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
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/walleterror"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
	goapiwalleterror "github.com/trustbloc/wallet-sdk/pkg/walleterror"
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
func NewInteraction(args *Args, opts *Opts) (*Interaction, error) {
	if args == nil {
		return nil, errors.New("args object must be provided")
	}

	if opts == nil {
		opts = NewOpts()
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
// Check the issuer's capabilities first using the IssuerCapabilities method.
func (i *Interaction) CreateAuthorizationURL(clientID, redirectURI string) (string, error) {
	return i.goAPIInteraction.CreateAuthorizationURL(clientID, redirectURI)
}

// CreateAuthorizationURLWithScopes is like CreateAuthorizationURL but allows OAuth2 scopes to be passed in.
func (i *Interaction) CreateAuthorizationURLWithScopes(clientID, redirectURI string,
	scopes *api.StringArray,
) (string, error) {
	if scopes == nil {
		scopes = api.NewStringArray()
	}

	return i.goAPIInteraction.CreateAuthorizationURLWithScopes(clientID, redirectURI, scopes.Strings)
}

// RequestCredentialWithAuth requests credential(s) from the issuer. This method can only be used for the
// authorization code flow, where it acts as the final step in the interaction with the issuer.
// For the equivalent methods for the pre-authorized code flow, see RequestCredential and RequestCredentialWithPIN
// instead.
//
// RequestCredentialWithAuth should be called only once all authorization pre-requisite steps have been completed.
// The redirect URI that you pass in here should look like the redirect URI that you passed in to the
// CreateAuthorizationURL, except that now it has some URL query parameters appended to it.
func (i *Interaction) RequestCredentialWithAuth(vm *api.VerificationMethod,
	redirectURIWithAuthCode string,
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

// RequestCredential requests credential(s) from the issuer. It is the final step in the interaction.
// This should be called only once any applicable authorization pre-requisite steps have been completed.
//
// For the pre-authorized code grant flow, this method can be called right after you've instantiated the Interaction
// object. If a user PIN is required (which can be checked via the IssuerCapabilities method), then
// RequestCredentialWithPIN must be called instead.
//
// For the authorization code grant flow, the CreateAuthorizationURL (or CreateAuthorizationURLWithScopes) and
// RequestAccessToken methods must be called first and any associated steps with those methods completed.
// See the docs on those methods for more details.

// RequestCredential requests credential(s) from the issuer. This method can only be used for the pre-authorized code
// flow, where it acts as the final step in the interaction with the issuer.
// For the equivalent method for the authorization code flow, see RequestCredentialWithAuth instead.
//
// If a PIN is required (which can be checked via the IssuerCapabilities method), then use instead.
func (i *Interaction) RequestCredential(vm *api.VerificationMethod) (*verifiable.CredentialsArray, error) {
	credentials, err := i.requestCredential(vm, "")
	if err != nil {
		parsedErr := walleterror.Parse(err.Error())

		if parsedErr.Category == openid4cigoapi.PINRequiredError {
			parsedErr.Details += ". Use the requestCredentialWithPIN method instead"
		}

		goAPIWalletErr := &goapiwalleterror.Error{
			Code:        parsedErr.Code,
			Scenario:    parsedErr.Category,
			ParentError: parsedErr.Details,
		}

		return nil, wrapper.ToMobileErrorWithTrace(goAPIWalletErr, i.oTel)
	}

	return credentials, nil
}

// RequestCredentialWithPIN is like RequestCredential, but takes in a PIN for use with the
// pre-authorized code grant flow. If the issuer does not require a PIN, then use RequestCredential instead. If you are
// using the authorization code grant flow, then use RequestCredentialWithAuth instead.
func (i *Interaction) RequestCredentialWithPIN(
	vm *api.VerificationMethod, pin string,
) (*verifiable.CredentialsArray, error) {
	return i.requestCredential(vm, pin)
}

// IssuerCapabilities returns an object which can be used to learn about an issuer's self-reported capabilities.
func (i *Interaction) IssuerCapabilities() *IssuerCapabilities {
	return &IssuerCapabilities{goAPIIssuerCapabilities: i.goAPIInteraction.IssuerCapabilities()}
}

// OTelTraceID returns open telemetry trace id.
func (i *Interaction) OTelTraceID() string {
	traceID := ""
	if i.oTel != nil {
		traceID = i.oTel.TraceID()
	}

	return traceID
}

func (i *Interaction) requestCredential(
	vm *api.VerificationMethod, pin string,
) (*verifiable.CredentialsArray, error) {
	signer, err := i.createSigner(vm)
	if err != nil {
		return nil, err
	}

	goAPICredentialRequest := &openid4cigoapi.CredentialRequestOpts{UserPIN: pin}

	credentials, err := i.goAPIInteraction.RequestCredential(goAPICredentialRequest, signer)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return toGomobileCredentials(credentials), nil
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

// IssuerURI returns the issuer's URI from the initiation request. It's useful to store this somewhere in case
// there's a later need to refresh credential display data using the latest display information from the issuer.
func (i *Interaction) IssuerURI() string {
	return i.goAPIInteraction.IssuerURI()
}

func createGoAPIClientConfig(config *Args,
	opts *Opts,
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
