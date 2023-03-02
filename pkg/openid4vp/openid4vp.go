/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4vp implements the OpenID4VP presentation flow.
package openid4vp

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jwt"
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const (
	requestURIPrefix = "openid-vc://?request_uri="
	tokenLiveTimeSec = 600

	activityLogOperation = "oidc-presentation"
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
	activityLogger       api.ActivityLogger
	didResolver          api.DIDResolver
	crypto               api.Crypto

	requestObject *requestObject
}

type authorizedResponse struct {
	IDTokenJWS string
	VPTokenJWS string
	State      string
}

// New creates new openid4vp instance.
// If no ActivityLogger is provided (via an option), then no activity logging will take place.
func New(
	authorizationRequest string,
	signatureVerifier jwtSignatureVerifier,
	didResolver api.DIDResolver,
	crypto api.Crypto,
	opts ...Opt,
) *Interaction {
	client, activityLogger := processOpts(opts)

	return &Interaction{
		authorizationRequest: authorizationRequest,
		signatureVerifier:    signatureVerifier,
		httpClient:           client,
		activityLogger:       activityLogger,
		didResolver:          didResolver,
		crypto:               crypto,
	}
}

// GetQuery creates query based on authorization request data.
func (o *Interaction) GetQuery() (*presexch.PresentationDefinition, error) {
	rawRequestObject, err := fetchRequestObject(o.httpClient, o.authorizationRequest)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			module,
			RequestObjectFetchFailedCode,
			RequestObjectFetchFailedError,
			fmt.Errorf("fetch request object: %w", err))
	}

	requestObject, err := verifyAuthorizationRequestAndDecodeClaims(rawRequestObject, o.signatureVerifier)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			module,
			VerifyAuthorizationRequestFailedCode,
			VerifyAuthorizationRequestFailedError,
			fmt.Errorf("verify authorization request: %w", err))
	}

	o.requestObject = requestObject

	return requestObject.Claims.VPToken.PresentationDefinition, nil
}

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredential(presentations []*verifiable.Presentation) error {
	response, err := createAuthorizedResponse(presentations, o.requestObject, o.didResolver, o.crypto)
	if err != nil {
		return walleterror.NewExecutionError(
			module,
			CreateAuthorizedResponseFailedCode,
			CreateAuthorizedResponseFailedError,
			fmt.Errorf("create authorized response failed: %w", err))
	}

	data := url.Values{}
	data.Set("id_token", response.IDTokenJWS)
	data.Set("vp_token", response.VPTokenJWS)
	data.Set("state", response.State)

	err = sendAuthorizedResponse(o.httpClient, data.Encode(), o.requestObject.RedirectURI)
	if err != nil {
		return walleterror.NewExecutionError(
			module,
			SendAuthorizedResponseFailedCode,
			SendAuthorizedResponseFailedError,
			fmt.Errorf("send authorized response failed: %w", err))
	}

	return o.activityLogger.Log(&api.Activity{
		ID:   uuid.New(),
		Type: api.LogTypeCredentialActivity,
		Time: time.Now(),
		Data: api.Data{
			Client:    o.requestObject.Registration.ClientName,
			Operation: activityLogOperation,
			Status:    api.ActivityLogStatusSuccess,
		},
	})
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

func doHTTPRequest(httpClient httpClient, method, endpointURL, contentType string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, endpointURL, body)
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

func createAuthorizedResponse( //nolint:funlen
	presentations []*verifiable.Presentation,
	requestObject *requestObject,
	didResolver api.DIDResolver,
	crypto api.Crypto,
) (*authorizedResponse, error) {
	if len(presentations) == 0 {
		return nil, fmt.Errorf("expected at least one presentation to present to verifier")
	}

	// TODO handle multiple presentations
	presentation := presentations[0]

	did, err := getRandomHolderDID(presentation)
	if err != nil {
		return nil, err
	}

	signer, err := getHolderSigner(did, didResolver, crypto)
	if err != nil {
		return nil, err
	}

	presentationSubmission := presentation.CustomFields["presentation_submission"]

	idToken := &idTokenClaims{
		VPToken: idTokenVPToken{
			PresentationSubmission: presentationSubmission,
		},
		Nonce: requestObject.Nonce,
		Exp:   time.Now().Unix() + tokenLiveTimeSec,
		Iss:   "https://self-issued.me/v2/openid-vc",
		Sub:   did,
		Aud:   requestObject.ClientID,
		Nbf:   time.Now().Unix(),
		Iat:   time.Now().Unix(),
		Jti:   uuid.NewString(),
	}

	presentation.CustomFields["presentation_submission"] = nil

	vpToken := vpTokenClaims{
		VP:    presentation,
		Nonce: requestObject.Nonce,
		Exp:   time.Now().Unix() + tokenLiveTimeSec,
		Iss:   did,
		Aud:   requestObject.ClientID,
		Nbf:   time.Now().Unix(),
		Iat:   time.Now().Unix(),
		Jti:   uuid.NewString(),
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

func getHolderSigner(holderDID string, didResolver api.DIDResolver, crypto api.Crypto) (api.JWTSigner, error) {
	docRes, err := didResolver.Resolve(holderDID)
	if err != nil {
		return nil, fmt.Errorf("resolve holder DID for signing: %w", err)
	}

	verificationMethods := docRes.DIDDocument.VerificationMethods(diddoc.AssertionMethod)

	if len(verificationMethods[diddoc.AssertionMethod]) == 0 {
		return nil, fmt.Errorf("holder DID has no assertion method for signing")
	}

	signingVM := verificationMethods[diddoc.AssertionMethod][0].VerificationMethod

	return common.NewJWSSigner(models.VerificationMethodFromDoc(&signingVM), crypto)
}

func getRandomHolderDID(pres *verifiable.Presentation) (string, error) {
	creds := pres.Credentials()

	if len(creds) == 0 {
		return "", fmt.Errorf("presentation has no credentials")
	}

	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(creds))))
	if err != nil {
		return "", err
	}

	selected := creds[idx.Int64()]

	var subjID string

	switch cred := selected.(type) {
	case *verifiable.Credential:
		subjID, err = verifiable.SubjectID(cred.Subject)
	case map[string]interface{}:
		subjID, err = verifiable.SubjectID(cred["credentialSubject"])
	}

	if err != nil || subjID == "" {
		return "", fmt.Errorf("presentation VC does not have a subject ID")
	}

	return subjID, nil
}
