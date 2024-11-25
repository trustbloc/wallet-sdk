/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/kms-go/doc/jose"
	"github.com/trustbloc/vc-go/jwt"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	noopmetricslogger "github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const (
	jwtProofTypeHeader              = "openid4vci-proof+jwt"
	attestationInitReqEventText     = "perform init attestation request"
	attestationCompleteReqEventText = "perform complete attestation request"
	challengeVCLifetime             = 5
)

// ServerAPI separates the HTTP client from the main client logic.
type ServerAPI interface {
	AttestationInit(req AttestWalletInitRequest, attestationURL string) (*AttestWalletInitResponse, error)
	AttestationComplete(req AttestWalletCompleteRequest, attestationURL string) (*AttestWalletCompleteResponse, error)
}

// Client is the client for the attestation service.
type Client struct {
	documentLoader ld.DocumentLoader
	attestationURL string
	api            ServerAPI
}

// ClientConfig holds the configuration for the attestation client.
type ClientConfig struct {
	HTTPClient     *http.Client
	MetricsLogger  api.MetricsLogger
	DocumentLoader ld.DocumentLoader
	AttestationURL string
	CustomAPI      ServerAPI
}

// NewClient returns a new attestation client.
func NewClient(config *ClientConfig) *Client {
	servAPI := config.CustomAPI
	if servAPI == nil {
		metricsLogger := config.MetricsLogger

		if metricsLogger == nil {
			metricsLogger = noopmetricslogger.NewMetricsLogger()
		}

		servAPI = &serverAPI{
			httpClient:    config.HTTPClient,
			metricsLogger: metricsLogger,
		}
	}

	return &Client{
		api:            servAPI,
		documentLoader: config.DocumentLoader,
		attestationURL: strings.TrimRight(config.AttestationURL, "/"),
	}
}

// GetAttestationVC requests an attestation VC from the attestation service.
func (c *Client) GetAttestationVC(
	attestationRequest AttestWalletInitRequest,
	signer api.JWTSigner,
) (*verifiable.Credential, error) {
	initResp, err := c.api.AttestationInit(attestationRequest, c.attestationURL)
	if err != nil {
		return nil, err
	}

	completeResp, err := c.attestationComplete(initResp.SessionID, initResp.Challenge, signer)
	if err != nil {
		return nil, err
	}

	attestationVC, err := verifiable.ParseCredential(
		[]byte(completeResp.WalletAttestationVC),
		verifiable.WithDisabledProofCheck(),
		verifiable.WithJSONLDDocumentLoader(c.documentLoader),
	)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			ErrorModule,
			ParseAttestationVCFailedCode,
			ParseAttestationVCFailedError,
			fmt.Errorf("parse attestation vc: %w", err))
	}

	return attestationVC, nil
}

func (c *Client) attestationComplete(
	sessionID,
	challenge string,
	signer api.JWTSigner,
) (*AttestWalletCompleteResponse, error) {
	did, err := getSignerDID(signer)
	if err != nil {
		return nil, err
	}

	claims := &jwtProofClaims{
		Issuer:   did,
		Audience: c.attestationURL,
		IssuedAt: time.Now().Unix(),
		Exp:      time.Now().Add(time.Minute * challengeVCLifetime).Unix(),
		Nonce:    challenge,
	}

	headers := jose.Headers{
		jose.HeaderType: jwtProofTypeHeader,
	}

	signedJWT, err := jwt.NewSigned(claims, jwt.SignParameters{AdditionalHeaders: headers}, signer)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			ErrorModule,
			JWTSigningFailedCode,
			JWTSigningFailedError,
			fmt.Errorf("failed to sign JWT: %w", err))
	}

	jws, err := signedJWT.Serialize(false)
	if err != nil {
		return nil, fmt.Errorf("serialize signed jwt: %w", err)
	}

	req := AttestWalletCompleteRequest{
		AssuranceLevel: "low",
		Proof: Proof{
			Jwt:       jws,
			ProofType: "jwt",
		},
		SessionID: sessionID,
	}

	return c.api.AttestationComplete(req, c.attestationURL)
}

func getSignerDID(jwtSigner api.JWTSigner) (string, error) {
	kidParts := strings.Split(jwtSigner.GetKeyID(), "#")
	if len(kidParts) < 2 { //nolint: mnd
		return "", walleterror.NewExecutionError(
			ErrorModule,
			KeyIDMissingDIDPartCode,
			KeyIDMissingDIDPartError,
			fmt.Errorf("key ID (%s) is missing the DID part", jwtSigner.GetKeyID()))
	}

	return kidParts[0], nil
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type serverAPI struct {
	httpClient    httpClient
	metricsLogger api.MetricsLogger
}

func (a *serverAPI) AttestationInit(
	req AttestWalletInitRequest,
	attestationURL string,
) (*AttestWalletInitResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var resp AttestWalletInitResponse

	err = httprequest.New(a.httpClient, a.metricsLogger).DoAndParse(
		http.MethodPost,
		attestationURL+"/init",
		"application/json",
		bytes.NewBuffer(body), attestationInitReqEventText, "", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("do request %s: %w", string(body), err)
	}

	return &resp, nil
}

func (a *serverAPI) AttestationComplete(
	req AttestWalletCompleteRequest,
	attestationURL string,
) (*AttestWalletCompleteResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var resp AttestWalletCompleteResponse

	err = httprequest.New(a.httpClient, a.metricsLogger).DoAndParse(
		http.MethodPost,
		attestationURL+"/complete",
		"application/json",
		bytes.NewBuffer(body), attestationCompleteReqEventText, "", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	return &resp, nil
}
