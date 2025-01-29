/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package attestation provides APIs for wallets to receive attestation credential.
package attestation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gowebpki/jcs"

	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/memstorage/legacy"
	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/otel"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	attestationgoapi "github.com/trustbloc/wallet-sdk/pkg/attestation"
)

// GetAttestationPayloadHash returns the SHA256 hash of the given attestation payload JSON.
// JSON is canonicalized according to RFC 8785.
func GetAttestationPayloadHash(attestationPayloadJSON string) (string, error) {
	transformed, err := jcs.Transform([]byte(attestationPayloadJSON))
	if err != nil {
		return "", err
	}

	return hashSHA256(transformed), nil
}

func hashSHA256(data []byte) string {
	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}

// Client is a client for the attestation API.
type Client struct {
	impl   *attestationgoapi.Client
	crypto api.Crypto
	oTel   *otel.Trace
}

// CreateClientArgs holds the arguments for creating a new client.
type CreateClientArgs struct {
	attestationURL                   string
	crypto                           api.Crypto
	disableOpenTelemetry             bool
	metricsLogger                    api.MetricsLogger
	additionalHeaders                api.Headers
	httpTimeout                      *time.Duration
	disableHTTPClientTLSVerification bool
	documentLoader                   api.LDDocumentLoader
	customAPI                        attestationgoapi.ServerAPI
}

// NewCreateClientArgs creates a new CreateClientArgs object.
func NewCreateClientArgs(attestationURL string, crypto api.Crypto) *CreateClientArgs {
	return &CreateClientArgs{
		attestationURL:       attestationURL,
		crypto:               crypto,
		disableOpenTelemetry: false,
	}
}

// DisableOpenTelemetry disables OpenTelemetry.
func (a *CreateClientArgs) DisableOpenTelemetry() *CreateClientArgs {
	a.disableOpenTelemetry = true

	return a
}

// SetHTTPTimeoutNanoseconds sets the timeout (in nanoseconds) for HTTP calls.
// Passing in 0 will disable timeouts.
func (a *CreateClientArgs) SetHTTPTimeoutNanoseconds(timeout int64) *CreateClientArgs {
	timeoutDuration := time.Duration(timeout)
	a.httpTimeout = &timeoutDuration

	return a
}

// AddHeader adds the given HTTP header to all REST calls made to the issuer during the OpenID4CI flow.
func (a *CreateClientArgs) AddHeader(header *api.Header) *CreateClientArgs {
	a.additionalHeaders.Add(header)

	return a
}

// DisableHTTPClientTLSVerify disables tls verification, should be used only for test purposes.
func (a *CreateClientArgs) DisableHTTPClientTLSVerify() *CreateClientArgs {
	a.disableHTTPClientTLSVerification = true

	return a
}

// SetDocumentLoader sets the document loader to use when parsing VCs received from the issuer.
// If no document loader is explicitly set, then a network-based loader will be used.
func (a *CreateClientArgs) SetDocumentLoader(documentLoader api.LDDocumentLoader) *CreateClientArgs {
	a.documentLoader = documentLoader

	return a
}

// SetMetricsLogger sets a metrics logger to use.
func (a *CreateClientArgs) SetMetricsLogger(metricsLogger api.MetricsLogger) *CreateClientArgs {
	a.metricsLogger = metricsLogger

	return a
}

// NewClient creates a new attestation client.
func NewClient(
	args *CreateClientArgs,
) (*Client, error) {
	if args == nil {
		return nil, wrapper.ToMobileError(walleterror.NewInvalidSDKUsageError(
			attestationgoapi.ErrorModule, errors.New("args object must be provided")))
	}

	var oTel *otel.Trace

	if !args.disableOpenTelemetry {
		var err error

		oTel, err = otel.NewTrace()
		if err != nil {
			return nil, wrapper.ToMobileError(err)
		}

		args.AddHeader(oTel.TraceHeader())
	}

	goAPIClientConfig, err := createGoAPIClientConfig(args)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, oTel)
	}

	client := attestationgoapi.NewClient(goAPIClientConfig)

	return &Client{
		impl:   client,
		crypto: args.crypto,
		oTel:   oTel,
	}, nil
}

// GetAttestationVC requests an attestation VC from the attestation service.
func (c *Client) GetAttestationVC(
	vm *api.VerificationMethod,
	attestationPayloadJSON string,
) (*verifiable.Credential, error) {
	signer, err := createSigner(vm, c.crypto)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, c.oTel)
	}

	var attestationRequest attestationgoapi.AttestWalletInitRequest

	err = json.Unmarshal([]byte(attestationPayloadJSON), &attestationRequest.Payload)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(walleterror.NewInvalidSDKUsageError(
			attestationgoapi.ErrorModule, fmt.Errorf("parse attestation vc payload: %w", err)), c.oTel)
	}

	credential, err := c.impl.GetAttestationVC(attestationRequest, signer)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, c.oTel)
	}

	return verifiable.NewCredential(credential), nil
}

func createGoAPIClientConfig(args *CreateClientArgs) (*attestationgoapi.ClientConfig, error) {
	httpClient := wrapper.NewHTTPClient(args.httpTimeout, args.additionalHeaders, args.disableHTTPClientTLSVerification)

	goAPIClientConfig := &attestationgoapi.ClientConfig{
		AttestationURL: args.attestationURL,
		MetricsLogger:  &wrapper.MobileMetricsLoggerWrapper{MobileAPIMetricsLogger: args.metricsLogger},
		HTTPClient:     httpClient,
		CustomAPI:      args.customAPI,
	}

	if args.documentLoader != nil {
		documentLoaderWrapper := &wrapper.DocumentLoaderWrapper{
			DocumentLoader: args.documentLoader,
		}

		goAPIClientConfig.DocumentLoader = documentLoaderWrapper
	} else {
		dlHTTPClient := wrapper.NewHTTPClient(args.httpTimeout, api.Headers{}, args.disableHTTPClientTLSVerification)

		var err error

		goAPIClientConfig.DocumentLoader, err = common.CreateJSONLDDocumentLoader(dlHTTPClient, legacy.NewProvider())
		if err != nil {
			return nil, err
		}
	}

	return goAPIClientConfig, nil
}

func createSigner(vm *api.VerificationMethod, crypto api.Crypto) (*common.JWSSigner, error) {
	if vm == nil {
		return nil, walleterror.NewInvalidSDKUsageError(openid4cigoapi.ErrorModule,
			errors.New("verification method must be provided"))
	}

	signer, err := common.NewJWSSigner(vm.ToSDKVerificationMethod(), crypto)
	if err != nil {
		return nil, err
	}

	return signer, nil
}
