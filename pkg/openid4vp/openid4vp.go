/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4vp implements the OpenID4VP presentation flow.
package openid4vp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jwt"
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

const (
	requestURIPrefix = "openid-vc://?request_uri="
	tokenLiveTimeSec = 600
)

type jwtSignatureVerifier interface {
	Verify(joseHeaders jose.Headers, payload, signingInput, signature []byte) error
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Interaction is used to help with OpenID4VP operations.
type Interaction struct {
	authorizationRequest string
	signatureVerifier    jwtSignatureVerifier
	httpClient           httpClient

	requestObject *requestObject
}

type authorizedResponse struct {
	IDTokenJWS string
	VPTokenJWS string
	State      string
}

// New creates new openid4vp instance.
func New(authorizationRequest string, signatureVerifier jwtSignatureVerifier, httpClient httpClient) *Interaction {
	return &Interaction{
		authorizationRequest: authorizationRequest,
		signatureVerifier:    signatureVerifier,
		httpClient:           httpClient,
	}
}

// GetQuery creates query based on authorization request data.
func (o *Interaction) GetQuery() (*presexch.PresentationDefinition, error) {
	rawRequestObject, err := fetchRequestObject(o.httpClient, o.authorizationRequest)
	if err != nil {
		return nil, fmt.Errorf("fetch request object: %w", err)
	}

	requestObject, err := verifyAuthorizationRequestAndDecodeClaims(rawRequestObject, o.signatureVerifier)
	if err != nil {
		return nil, fmt.Errorf("verify authorization request: %w", err)
	}

	o.requestObject = requestObject

	return requestObject.Claims.VPToken.PresentationDefinition, nil
}

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredential(presentation *verifiable.Presentation, jwtSigner api.JWTSigner) error {
	response, err := createAuthorizedResponse(presentation, o.requestObject, jwtSigner)
	if err != nil {
		return fmt.Errorf("create authorized response failed: %w", err)
	}

	responseBody := fmt.Sprintf("id_token=%s&vp_token=%s&state=%s",
		response.IDTokenJWS, response.VPTokenJWS, response.State)

	err = sendAuthorizedResponse(o.httpClient, responseBody, o.requestObject.RedirectURI)
	if err != nil {
		return fmt.Errorf("send authorized response failed: %w", err)
	}

	return nil
}

func fetchRequestObject(httpClient httpClient, authorizationRequest string) (string, error) {
	if !strings.HasPrefix(authorizationRequest, requestURIPrefix) {
		return authorizationRequest, nil
	}

	endpointURL := strings.TrimPrefix(authorizationRequest, requestURIPrefix)

	respBytes, err := doHTTPRequest(httpClient, http.MethodGet, endpointURL, "", nil)
	if err != nil {
		return "", err
	}

	return string(respBytes), nil
}

func doHTTPRequest(httpClient httpClient, method, url, contentType string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close() //nolint: errcheck
	}()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"expected status code %d but got status code %d with response body %s instead",
			http.StatusOK, resp.StatusCode, respBytes)
	}

	return respBytes, nil
}

func verifyAuthorizationRequestAndDecodeClaims(
	rawRequestObject string,
	signatureVerifier jwtSignatureVerifier,
) (*requestObject, error) {
	requestObject := &requestObject{}

	err := verifyTokenSignature(rawRequestObject, requestObject, signatureVerifier)
	if err != nil {
		return nil, err
	}

	return requestObject, nil
}

func verifyTokenSignature(rawJwt string, claims interface{}, verifier jose.SignatureVerifier) error {
	jsonWebToken, err := jwt.Parse(rawJwt, jwt.WithSignatureVerifier(verifier))
	if err != nil {
		return fmt.Errorf("parse JWT: %w", err)
	}

	err = jsonWebToken.DecodeClaims(claims)
	if err != nil {
		return fmt.Errorf("decode claims: %w", err)
	}

	return nil
}

func createAuthorizedResponse(
	presentation *verifiable.Presentation,
	requestObject *requestObject,
	signer api.JWTSigner,
) (*authorizedResponse, error) {
	kidParts := strings.Split(signer.GetKeyID(), "#")
	if len(kidParts) < 2 { //nolint: gomnd
		return nil, fmt.Errorf("kid not containing did part %s", signer.GetKeyID())
	}

	did := kidParts[0]
	presentationSubmission := presentation.CustomFields["presentation_submission"]

	idToken := &idTokenClaims{
		VPToken: idTokenVPToken{
			PresentationSubmission: presentationSubmission,
		},
		Nonce: requestObject.Nonce,
		Exp:   time.Now().Unix() + tokenLiveTimeSec,
		Iss:   "https://self-issued.me/v2/openid-vc",
		Sub:   did,
	}

	presentation.CustomFields["presentation_submission"] = nil

	vpToken := vpTokenClaims{
		VP:    presentation,
		Nonce: requestObject.Nonce,
		Exp:   time.Now().Unix() + tokenLiveTimeSec,
		Iss:   did,
	}

	idTokenJWS, err := signToken(idToken, signer)
	if err != nil {
		return nil, fmt.Errorf("sign id_token: %w", err)
	}

	vpTokenJWS, err := signToken(vpToken, signer)
	if err != nil {
		return nil, fmt.Errorf("sign vp_token: %w", err)
	}

	return &authorizedResponse{IDTokenJWS: idTokenJWS, VPTokenJWS: vpTokenJWS, State: requestObject.State}, nil
}

func signToken(claims interface{}, signer api.JWTSigner) (string, error) {
	token, err := jwt.NewSigned(claims, nil, signer)
	if err != nil {
		return "", fmt.Errorf("sign token failed: %w", err)
	}

	tokenBytes, err := token.Serialize(false)
	if err != nil {
		return "", fmt.Errorf("serialize token failed: %w", err)
	}

	return tokenBytes, nil
}

func sendAuthorizedResponse(httpClient httpClient, responseBody, redirectURI string) error {
	_, err := doHTTPRequest(httpClient, http.MethodPost,
		redirectURI, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(responseBody)))
	if err != nil {
		return err
	}

	return nil
}
