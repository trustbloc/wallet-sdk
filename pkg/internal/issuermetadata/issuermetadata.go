/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package issuermetadata contains a function for fetching issuer metadata.
package issuermetadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/trustbloc/vc-go/jwt"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	"github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

const fetchIssuerMetadataViaGETReqEventText = "Fetch issuer metadata via an HTTP GET request to %s"

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Get gets an issuer's metadata by doing a lookup on its OpenID configuration endpoint.
// issuerURI is expected to be the base URL for the issuer.
func Get(issuerURI string, httpClient httpClient, metricsLogger api.MetricsLogger, parentEvent string,
	signatureVerifier jwt.ProofChecker,
) (*issuer.Metadata, error) {
	if metricsLogger == nil {
		metricsLogger = noop.NewMetricsLogger()
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: api.DefaultHTTPTimeout}
	}

	metadataEndpoint := strings.TrimSuffix(issuerURI, "/") + "/.well-known/openid-credential-issuer"

	responseBytes, err := httprequest.New(httpClient, metricsLogger).Do(
		http.MethodGet, metadataEndpoint, "", nil,
		fmt.Sprintf(fetchIssuerMetadataViaGETReqEventText, metadataEndpoint), parentEvent, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get response from the issuer's metadata endpoint: %w", err)
	}

	return responseBytesToIssuerMetadataObject(responseBytes, signatureVerifier)
}

func responseBytesToIssuerMetadataObject(responseBytes []byte,
	signatureVerifier jwt.ProofChecker,
) (*issuer.Metadata, error) {
	// The issuer metadata can come in one of two formats - either directly as JSON, or as a JWT.
	var metadata issuer.Metadata

	// First, try parsing directly as JSON.
	err := json.Unmarshal(responseBytes, &metadata)
	if err != nil {
		return nil, fmt.Errorf("decode metadata: %w", err)
	}

	if metadata.SignedMetadata != "" {
		return issuerMetadataObjectFromJWT(metadata.SignedMetadata, signatureVerifier, err)
	}

	return &metadata, nil
}

// errUnmarshal is the error that happened when the response bytes couldn't be unmarshalled into the issuer metadata
// struct directly. It's passed here so that it can be included in the error message in case the response
// is also not a JWT. This gives the caller additional information that can help them to more easily debug the cause
// of the parsing failure.
func issuerMetadataObjectFromJWT(signedMetadata string, signatureVerifier jwt.ProofChecker,
	errUnmarshal error,
) (*issuer.Metadata, error) {
	// Try to parse it as a JWT.
	// But first, make sure a signature verifier was passed in. If it wasn't, then the jwt.Parse call below will
	// panic.
	if signatureVerifier == nil {
		return nil, errors.New("missing signature verifier")
	}

	jsonWebToken, payload, errParseJWT := jwt.ParseAndCheckProof(signedMetadata, signatureVerifier, false)
	if errParseJWT != nil {
		return nil, fmt.Errorf("failed to parse the response from the issuer's OpenID Credential Issuer "+
			"endpoint as JSON or as a JWT: %w", errors.Join(errUnmarshal, errParseJWT))
	}

	var metadata issuer.Metadata

	err := json.Unmarshal(payload, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the marshalled issuer metadata JSON into an "+
			"issuer metadata object OpenID configuration endpoint: %w", err)
	}

	kid, ok := jsonWebToken.Headers.KeyID()
	if !ok {
		return nil, errors.New("issuer's OpenID configuration endpoint returned a JWT, but the kid header " +
			"value is missing or is not a string")
	}

	metadata.SetJWTKID(kid)

	return &metadata, nil
}
