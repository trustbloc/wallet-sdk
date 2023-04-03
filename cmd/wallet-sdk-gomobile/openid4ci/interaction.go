/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4ci provides APIs for wallets to receive verifiable credentials via OIDC for Credential Issuance.
package openid4ci

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/walleterror"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
	goapiwalleterror "github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// Interaction represents a single OpenID4CI interaction between a wallet and an issuer. The methods defined on this
// object are used to help guide the calling code through the OpenID4CI flow.
type Interaction struct {
	goAPIInteraction *openid4cigoapi.Interaction
	crypto           api.Crypto
}

// AuthorizeResult is the object returned from the Client.Authorize method.
// An empty/missing AuthorizationRedirectEndpoint indicates that the wallet is pre-authorized.
type AuthorizeResult struct {
	AuthorizationRedirectEndpoint string
	UserPINRequired               bool
}

// NewInteraction creates a new OpenID4CI Interaction.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// Calling this function represents taking the first step in the flow.
// This function takes in an Initiate Issuance Request object from an issuer (as defined in
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-5.1), encoded using URL query
// parameters. This object is intended for going through the full flow only once (i.e. one interaction), after which
// it should be discarded. Any new interactions should use a fresh Interaction instance.
func NewInteraction(args *Args, opts *Opts) (*Interaction, error) {
	if args == nil {
		return nil, errors.New("args object must be provided")
	}

	if opts == nil {
		opts = NewOpts()
	}

	goAPIClientConfig := createGoAPIClientConfig(args, opts)

	goAPIInteraction, err := openid4cigoapi.NewInteraction(args.initiateIssuanceURI, goAPIClientConfig)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &Interaction{
		crypto:           args.crypto,
		goAPIInteraction: goAPIInteraction,
	}, nil
}

// Authorize is used by a wallet to authorize an issuer's OIDC Verifiable Credential Issuance Request.
// After initializing the Interaction object with an Issuance Request, this should be the first method you call in
// order to continue with the flow.
// It only supports the pre-authorized flow in its current implementation.
// Once the authorization flow is implemented, the following section of the spec will be relevant:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html#name-authorization-endpoint
func (i *Interaction) Authorize() (*AuthorizeResult, error) {
	authorizationResultGoAPI, err := i.goAPIInteraction.Authorize()
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	authorizationResult := &AuthorizeResult{
		AuthorizationRedirectEndpoint: authorizationResultGoAPI.AuthorizationRedirectEndpoint,
		UserPINRequired:               authorizationResultGoAPI.UserPINRequired,
	}

	return authorizationResult, nil
}

// RequestCredential is the final step in the interaction. It requests credential(s) from the issuer.
// If a PIN is needed, then use RequestCredentialWithPIN instead.
// This is called after the wallet is authorized and is ready to receive credential(s).
// Relevant section of the spec:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html#name-credential-endpoint
func (i *Interaction) RequestCredential(vm *api.VerificationMethod) (*api.VerifiableCredentialsArray, error) {
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

		return nil, goAPIWalletErr
	}

	return credentials, nil
}

// RequestCredentialWithPIN is the final step in the interaction. It requests credential(s) from the issuer with the
// given PIN. If no PIN is need, then RequestCredential can be used instead.
// This is called after the wallet is authorized and is ready to receive credential(s).
// Relevant section of the spec:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html#name-credential-endpoint
func (i *Interaction) RequestCredentialWithPIN(
	vm *api.VerificationMethod, pin string,
) (*api.VerifiableCredentialsArray, error) {
	return i.requestCredential(vm, pin)
}

func (i *Interaction) requestCredential(
	vm *api.VerificationMethod, pin string,
) (*api.VerifiableCredentialsArray, error) {
	if vm == nil {
		return nil, errors.New("verification method must be provided")
	}

	signer, err := common.NewJWSSigner(vm.ToSDKVerificationMethod(), i.crypto)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	goAPICredentialRequest := &openid4cigoapi.CredentialRequestOpts{UserPIN: pin}

	credentials, err := i.goAPIInteraction.RequestCredential(goAPICredentialRequest, signer)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	gomobileCredentials := api.NewVerifiableCredentialsArray()

	for i := range credentials {
		gomobileCredentials.Add(api.NewVerifiableCredential(credentials[i]))
	}

	return gomobileCredentials, nil
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

	httpClient := wrapper.NewHTTPClient()
	httpClient.AddHeaders(&opts.additionalHeaders)
	httpClient.DisableTLSVerification = opts.disableHTTPClientTLSVerification

	goAPIClientConfig := &openid4cigoapi.ClientConfig{
		ClientID:             config.clientID,
		DIDResolver:          &wrapper.VDRResolverWrapper{DIDResolver: config.didResolver},
		ActivityLogger:       activityLogger,
		MetricsLogger:        &wrapper.MobileMetricsLoggerWrapper{MobileAPIMetricsLogger: opts.metricsLogger},
		DisableVCProofChecks: opts.disableVCProofChecks,
		HTTPClient:           httpClient,
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
