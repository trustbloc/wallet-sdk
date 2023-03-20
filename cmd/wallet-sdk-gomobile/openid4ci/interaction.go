/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4ci provides APIs for wallets to receive verifiable credentials via OIDC for Credential Issuance.
package openid4ci

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/walleterror"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
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

// CredentialRequestOpts represents the data (required and optional) that is used in the
// final step of the OpenID4CI flow, where the wallet requests the credential from the issuer.
type CredentialRequestOpts struct {
	UserPIN string
}

// NewCredentialRequestOpts returns a new NewCredentialRequestOpts object.
func NewCredentialRequestOpts(userPIN string) *CredentialRequestOpts {
	return &CredentialRequestOpts{UserPIN: userPIN}
}

// ClientConfig contains various parameters for an OpenID4CI Interaction.
// ActivityLogger is optional, but if provided then activities will be logged there.
// If not provided, then no activities will be logged.
type ClientConfig struct {
	ClientID                string
	Crypto                  api.Crypto
	DIDResolver             api.DIDResolver
	ActivityLogger          api.ActivityLogger
	MetricsLogger           api.MetricsLogger
	disableVCProofChecks    bool
	httpClientSkipTLSVerify bool
}

// NewClientConfig creates the client config object.
// ActivityLogger is optional, but if provided then activities will be logged there.
// If not provided, then no activities will be logged.
func NewClientConfig(clientID string, crypto api.Crypto,
	didRes api.DIDResolver, activityLogger api.ActivityLogger,
) *ClientConfig {
	return &ClientConfig{
		ClientID:       clientID,
		Crypto:         crypto,
		DIDResolver:    didRes,
		ActivityLogger: activityLogger,
	}
}

// DisableVCProofChecks disables VC proof checks during the OpenID4CI interaction flow.
func (c *ClientConfig) DisableVCProofChecks() {
	c.disableVCProofChecks = true
}

// DisableHTTPClientTLSVerify disables tls verification, should be used only for test purposes.
func (c *ClientConfig) DisableHTTPClientTLSVerify() {
	c.httpClientSkipTLSVerify = true
}

// NewInteraction creates a new OpenID4CI Interaction.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// Calling this function represents taking the first step in the flow.
// This function takes in an Initiate Issuance Request object from an issuer (as defined in
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-5.1), encoded using URL query
// parameters. This object is intended for going through the full flow only once (i.e. one interaction), after which
// it should be discarded. Any new interactions should use a fresh Interaction instance.
// If no ActivityLogger is provided (via the ClientConfig object), then no activity logging will take place.
func NewInteraction(
	initiateIssuanceURI string, config *ClientConfig,
) (*Interaction, error) {
	goAPIClientConfig := unwrapConfig(config)

	goAPIInteraction, err := openid4cigoapi.NewInteraction(initiateIssuanceURI, goAPIClientConfig)
	if err != nil {
		return nil, walleterror.ToMobileError(err)
	}

	return &Interaction{
		crypto:           config.Crypto,
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
		return nil, walleterror.ToMobileError(err)
	}

	authorizationResult := &AuthorizeResult{
		AuthorizationRedirectEndpoint: authorizationResultGoAPI.AuthorizationRedirectEndpoint,
		UserPINRequired:               authorizationResultGoAPI.UserPINRequired,
	}

	return authorizationResult, nil
}

// RequestCredential is the final step in the interaction.
// This is called after the wallet is authorized and is ready to receive credential(s).
// Relevant sections of the spec:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html#name-credential-endpoint
func (i *Interaction) RequestCredential(
	credentialRequestOpts *CredentialRequestOpts,
	vm *api.VerificationMethod,
) (*api.VerifiableCredentialsArray, error) {
	if credentialRequestOpts == nil {
		credentialRequestOpts = &CredentialRequestOpts{}
	}

	goAPICredentialRequest := &openid4cigoapi.CredentialRequestOpts{UserPIN: credentialRequestOpts.UserPIN}

	signer, err := common.NewJWSSigner(vm.ToSDKVerificationMethod(), i.crypto)
	if err != nil {
		return nil, walleterror.ToMobileError(err)
	}

	credentials, err := i.goAPIInteraction.RequestCredential(goAPICredentialRequest, signer)
	if err != nil {
		return nil, walleterror.ToMobileError(err)
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

func unwrapConfig(config *ClientConfig) *openid4cigoapi.ClientConfig {
	activityLogger := createGoAPIActivityLogger(config.ActivityLogger)

	httpClient := common.DefaultHTTPClient()
	if config.httpClientSkipTLSVerify {
		httpClient = common.InsecureHTTPClient()
	}

	return &openid4cigoapi.ClientConfig{
		ClientID:             config.ClientID,
		DIDResolver:          &wrapper.VDRResolverWrapper{DIDResolver: config.DIDResolver},
		ActivityLogger:       activityLogger,
		MetricsLogger:        &wrapper.MobileMetricsLoggerWrapper{MobileAPIMetricsLogger: config.MetricsLogger},
		DisableVCProofChecks: config.disableVCProofChecks,
		HTTPClient:           httpClient,
	}
}

func createGoAPIActivityLogger(mobileAPIActivityLogger api.ActivityLogger) goapi.ActivityLogger {
	if mobileAPIActivityLogger == nil {
		return nil // Will result in activity logging being disabled in the OpenID4CI Interaction object.
	}

	return &wrapper.MobileActivityLoggerWrapper{MobileAPIActivityLogger: mobileAPIActivityLogger}
}
