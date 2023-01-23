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
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jwt"
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const (
	// Base operation type.
	logOperationTypeOIDCPresentation = "OIDCPresentation"

	// Sub-operation types.
	logOperationTypeGetQuery          = logOperationTypeOIDCPresentation + "_GetQuery"
	logOperationTypePresentCredential = logOperationTypeOIDCPresentation + "_PresentCredential"

	// Sub-sub-operation type.
	logOperationTypeFetchRequestObject = logOperationTypeGetQuery + "_FetchRequestObject"

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
	logger               api.Logger

	requestObject *requestObject
}

type authorizedResponse struct {
	IDTokenJWS string
	VPTokenJWS string
	State      string
}

// New creates new openid4vp instance.
func New(authorizationRequest string, signatureVerifier jwtSignatureVerifier, opts ...Opt) *Interaction {
	client, logger := processOpts(opts)

	return &Interaction{
		authorizationRequest: authorizationRequest,
		signatureVerifier:    signatureVerifier,
		httpClient:           client,
		logger:               logger,
	}
}

// GetQuery creates query based on authorization request data.
func (o *Interaction) GetQuery() (*presexch.PresentationDefinition, error) {
	rawRequestObject, err := o.fetchRequestObject(o.httpClient, o.authorizationRequest)
	if err != nil {
		errExecution := walleterror.NewExecutionError(
			module,
			RequestObjectFetchFailedCode,
			RequestObjectFetchFailedError,
			fmt.Errorf("fetch request object: %w", err))

		o.logFailure(logOperationTypeGetQuery, errExecution)

		return nil, errExecution
	}

	requestObject, err := verifyAuthorizationRequestAndDecodeClaims(rawRequestObject, o.signatureVerifier)
	if err != nil {
		errExecution := walleterror.NewExecutionError(
			module,
			VerifyAuthorizationRequestFailedCode,
			VerifyAuthorizationRequestFailedError,
			fmt.Errorf("verify authorization request: %w", err))

		o.logFailure(logOperationTypeGetQuery, errExecution)

		return nil, errExecution
	}

	o.logSuccess(logOperationTypeGetQuery)

	o.requestObject = requestObject

	return requestObject.Claims.VPToken.PresentationDefinition, nil
}

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredential(presentation *verifiable.Presentation, jwtSigner api.JWTSigner) error {
	response, err := createAuthorizedResponse(presentation, o.requestObject, jwtSigner)
	if err != nil {
		errExecution := walleterror.NewExecutionError(
			module,
			CreateAuthorizedResponseFailedCode,
			CreateAuthorizedResponseFailedError,
			fmt.Errorf("create authorized response failed: %w", err))

		o.logFailure(logOperationTypePresentCredential, errExecution)

		return errExecution
	}

	data := url.Values{}
	data.Set("id_token", response.IDTokenJWS)
	data.Set("vp_token", response.VPTokenJWS)
	data.Set("state", response.State)

	err = sendAuthorizedResponse(o.httpClient, data.Encode(), o.requestObject.RedirectURI)
	if err != nil {
		errExecution := walleterror.NewExecutionError(
			module,
			SendAuthorizedResponseFailedCode,
			SendAuthorizedResponseFailedError,
			fmt.Errorf("send authorized response failed: %w", err))

		o.logFailure(logOperationTypePresentCredential, errExecution)

		return errExecution
	}

	o.logSuccess(logOperationTypePresentCredential)

	return nil
}

func (o *Interaction) logFailure(operationType string, err error) {
	o.log(api.LogStatusFailure, operationType, err)
}

func (o *Interaction) logSuccess(operationType string) {
	o.log(api.LogStatusSuccess, operationType, nil)
}

func (o *Interaction) log(status, operationType string, err error) {
	currentTime := time.Now()

	logEntry := &api.LogEntry{
		ID:   uuid.New().String(),
		Type: api.LogTypeCredentialActivity,
		Time: &currentTime,
		Data: &api.LogData{
			Operation: operationType,
			Status:    status,
		},
	}

	if err != nil {
		logEntry.Data.Params = map[string]interface{}{"error": err.Error()}
	}

	o.logger.Log(logEntry)
}

// If an HTTP request has to be made to fetch the request object, then on a successful response the logger will be
// used to log a success message.
func (o *Interaction) fetchRequestObject(httpClient httpClient, authorizationRequest string) (string, error) {
	if !strings.HasPrefix(authorizationRequest, requestURIPrefix) {
		return authorizationRequest, nil
	}

	endpointURL := strings.TrimPrefix(authorizationRequest, requestURIPrefix)

	respBytes, err := doHTTPRequest(httpClient, http.MethodGet, endpointURL, "", nil)
	if err != nil {
		return "", err
	}

	o.logSuccess(logOperationTypeFetchRequestObject)

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
