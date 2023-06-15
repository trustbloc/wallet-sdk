/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4ci provides APIs for wallets to receive verifiable credentials via OIDC for Credential Issuance.
package openid4ci

import (
	"errors"

	afgoverifiable "github.com/hyperledger/aries-framework-go/component/models/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/otel"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

// Interaction represents a single OpenID4CI interaction between a wallet and an issuer.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// An Interaction is a stateful object, and is intended for going through the full flow only once
// after which it should be discarded. Any new interactions should use a fresh Interaction instance.
type Interaction struct {
	goAPIInteraction *openid4cigoapi.Interaction
	crypto           api.Crypto
	oTel             *otel.Trace
}

// NewInteraction creates a new OpenID4CI Interaction.
func NewInteraction(args *InteractionArgs, opts *InteractionOpts) (*Interaction, error) {
	if args == nil {
		return nil, errors.New("args object must be provided")
	}

	if opts == nil {
		opts = NewInteractionOpts()
	}

	var oTel *otel.Trace

	if !opts.disableOpenTelemetry {
		var err error

		oTel, err = otel.NewTrace()
		if err != nil {
			return nil, wrapper.ToMobileError(err)
		}

		opts.AddHeader(oTel.TraceHeader())
	}

	goAPIClientConfig := createGoAPIClientConfig(args, opts)

	goAPIInteraction, err := openid4cigoapi.NewInteraction(args.initiateIssuanceURI, goAPIClientConfig)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, oTel)
	}

	return &Interaction{
		crypto:           args.crypto,
		goAPIInteraction: goAPIInteraction,
		oTel:             oTel,
	}, nil
}

// CreateAuthorizationURL creates an authorization URL that can be opened in a browser to proceed to the login page.
// It is the first step in the authorization code flow.
// It creates the authorization URL that can be opened in a browser to proceed to the login page.
// This method can only be used if the issuer supports authorization code grants.
// Check the issuer's capabilities first using the Capabilities method.
func (i *Interaction) CreateAuthorizationURL(clientID, redirectURI string,
	opts *CreateAuthorizationURLOpts,
) (string, error) {
	if opts == nil {
		opts = NewCreateAuthorizationURLOpts()
	}

	if opts.scopes == nil {
		opts.scopes = api.NewStringArray()
	}

	return i.goAPIInteraction.CreateAuthorizationURL(clientID, redirectURI,
		openid4cigoapi.WithScopes(opts.scopes.Strings))
}

// RequestCredentialWithPreAuth requests credential(s) from the issuer. This method can only be used for the
// pre-authorized code flow, where it acts as the final step in the interaction with the issuer.
// For the equivalent method for the authorization code flow, see RequestCredentialWithAuth instead.
// If a PIN is required (which can be checked via the Capabilities method), then it must be passed
// into this method via the SetPIN method on the RequestCredentialWithPreAuthOpts object.
func (i *Interaction) RequestCredentialWithPreAuth(
	vm *api.VerificationMethod, opts *RequestCredentialWithPreAuthOpts,
) (*verifiable.CredentialsArray, error) {
	if opts == nil {
		opts = NewRequestCredentialWithPreAuthOpts()
	}

	signer, err := i.createSigner(vm)
	if err != nil {
		return nil, err
	}

	credentials, err := i.goAPIInteraction.RequestCredentialWithPreAuth(signer, openid4cigoapi.WithPIN(opts.pin))
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return toGomobileCredentials(credentials), nil
}

// RequestCredentialWithAuth requests credential(s) from the issuer. This method can only be used for the
// authorization code flow, where it acts as the final step in the interaction with the issuer.
// For the equivalent method for the pre-authorized code flow, see RequestCredentialWithPreAuth instead.
//
// RequestCredentialWithAuth should be called only once all authorization pre-requisite steps have been completed.
// The redirect URI that you pass in here should look like the redirect URI that you passed in to the
// CreateAuthorizationURL, except that now it has some URL query parameters appended to it.
func (i *Interaction) RequestCredentialWithAuth(vm *api.VerificationMethod,
	redirectURIWithAuthCode string, opts *RequestCredentialWithAuthOpts,
) (*verifiable.CredentialsArray, error) {
	signer, err := i.createSigner(vm)
	if err != nil {
		return nil, err
	}

	credentials, err := i.goAPIInteraction.RequestCredentialWithAuth(signer, redirectURIWithAuthCode)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return toGomobileCredentials(credentials), nil
}

// RequestCredential is equivalent to calling the RequestCredentialWithPreAuth method with no PIN set.
//
// Deprecated: This method exists only to prevent existing code from breaking. It will be removed in a future version
// as the method name can be confusing.
func (i *Interaction) RequestCredential(vm *api.VerificationMethod) (*verifiable.CredentialsArray, error) {
	return i.RequestCredentialWithPreAuth(vm, nil)
}

// RequestCredentialWithPIN is equivalent to calling the RequestCredentialWithPreAuth method with a PIN set.
//
// Deprecated: This method exists only to prevent existing code from breaking. It will be removed in a future version
// as the method name can be confusing and the lack of an options object means it's not possible to easily add new
// options in the future without breaking existing code.
func (i *Interaction) RequestCredentialWithPIN(
	vm *api.VerificationMethod, pin string,
) (*verifiable.CredentialsArray, error) {
	return i.RequestCredentialWithPreAuth(vm, NewRequestCredentialWithPreAuthOpts().SetPIN(pin))
}

// IssuerURI returns the issuer's URI from the initiation request. It's useful to store this somewhere in case
// there's a later need to refresh credential display data using the latest display information from the issuer.
func (i *Interaction) IssuerURI() string {
	return i.goAPIInteraction.IssuerURI()
}

// PreAuthorizedCodeGrantTypeSupported indicates whether an issuer supports the pre-authorized code grant type.
func (i *Interaction) PreAuthorizedCodeGrantTypeSupported() bool {
	return i.goAPIInteraction.PreAuthorizedCodeGrantTypeSupported()
}

// PreAuthorizedCodeGrantParams returns an object that can be used to determine an issuer's pre-authorized code grant
// parameters. The caller should call the PreAuthorizedCodeGrantTypeSupported method first and only call this method to
// get the params if PreAuthorizedCodeGrantTypeSupported returns true. This method only returns an error if
// PreAuthorizedCodeGrantTypeSupported returns false, so the error return here can be safely ignored if
// PreAuthorizedCodeGrantTypeSupported returns true.
func (i *Interaction) PreAuthorizedCodeGrantParams() (*PreAuthorizedCodeGrantParams, error) {
	goAPIPreAuthorizedCodeGrantParams, err := i.goAPIInteraction.PreAuthorizedCodeGrantParams()
	if err != nil {
		return nil, err
	}

	return &PreAuthorizedCodeGrantParams{
		goAPIPreAuthorizedCodeGrantParams: goAPIPreAuthorizedCodeGrantParams,
	}, nil
}

// AuthorizationCodeGrantTypeSupported indicates whether an issuer supports the authorization code grant type.
func (i *Interaction) AuthorizationCodeGrantTypeSupported() bool {
	return i.goAPIInteraction.AuthorizationCodeGrantTypeSupported()
}

// OTelTraceID returns open telemetry trace id.
func (i *Interaction) OTelTraceID() string {
	traceID := ""
	if i.oTel != nil {
		traceID = i.oTel.TraceID()
	}

	return traceID
}

func (i *Interaction) createSigner(vm *api.VerificationMethod) (*common.JWSSigner, error) {
	if vm == nil {
		return nil, errors.New("verification method must be provided")
	}

	signer, err := common.NewJWSSigner(vm.ToSDKVerificationMethod(), i.crypto)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return signer, nil
}

func createGoAPIClientConfig(config *InteractionArgs,
	opts *InteractionOpts,
) *openid4cigoapi.ClientConfig {
	activityLogger := createGoAPIActivityLogger(opts.activityLogger)

	httpClient := wrapper.NewHTTPClient(opts.httpTimeout, opts.additionalHeaders, opts.disableHTTPClientTLSVerification)

	goAPIClientConfig := &openid4cigoapi.ClientConfig{
		DIDResolver:                      &wrapper.VDRResolverWrapper{DIDResolver: config.didResolver},
		ActivityLogger:                   activityLogger,
		MetricsLogger:                    &wrapper.MobileMetricsLoggerWrapper{MobileAPIMetricsLogger: opts.metricsLogger},
		DisableVCProofChecks:             opts.disableVCProofChecks,
		NetworkDocumentLoaderHTTPTimeout: opts.httpTimeout,
		HTTPClient:                       httpClient,
	}

	if opts.documentLoader != nil {
		documentLoaderWrapper := &wrapper.DocumentLoaderWrapper{
			DocumentLoader: opts.documentLoader,
		}

		goAPIClientConfig.DocumentLoader = documentLoaderWrapper
	}

	return goAPIClientConfig
}

func createGoAPIActivityLogger(mobileAPIActivityLogger api.ActivityLogger) goapi.ActivityLogger {
	if mobileAPIActivityLogger == nil {
		return nil // Will result in activity logging being disabled in the OpenID4CI Interaction object.
	}

	return &wrapper.MobileActivityLoggerWrapper{MobileAPIActivityLogger: mobileAPIActivityLogger}
}

func toGomobileCredentials(credentials []*afgoverifiable.Credential) *verifiable.CredentialsArray {
	gomobileCredentials := verifiable.NewCredentialsArray()

	for i := range credentials {
		gomobileCredentials.Add(verifiable.NewCredential(credentials[i]))
	}

	return gomobileCredentials
}
