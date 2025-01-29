/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"time"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"

	"github.com/trustbloc/wallet-sdk/pkg/walleterror"

	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// WalletInitiatedInteraction represents a single wallet-instantiated OpenID4CI interaction between a wallet and an
// issuer.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// An WalletInitiatedInteraction is a stateful object, and is intended for going through the full flow only once
// after which it should be discarded. Any new interactions should use a fresh WalletInitiatedInteraction instance.
type WalletInitiatedInteraction struct {
	interaction *interaction

	credentialFormat  string
	credentialTypes   []string
	credentialContext []string
}

// NewWalletInitiatedInteraction creates a new OpenID4CI WalletInitiatedInteraction.
// If no ActivityLogger is provided (via the ClientConfig object), then no activity logging will take place.
func NewWalletInitiatedInteraction(issuerURI string, config *ClientConfig) (*WalletInitiatedInteraction, error) {
	timeStartNewInteraction := time.Now()

	err := validateRequiredParameters(config)
	if err != nil {
		return nil, walleterror.NewInvalidSDKUsageError(ErrorModule, err)
	}

	setDefaults(config)

	return &WalletInitiatedInteraction{
			interaction: &interaction{
				issuerURI:            issuerURI,
				didResolver:          config.DIDResolver,
				activityLogger:       config.ActivityLogger,
				metricsLogger:        config.MetricsLogger,
				disableVCProofChecks: config.DisableVCProofChecks,
				documentLoader:       config.DocumentLoader,
				httpClient:           config.HTTPClient,
			},
		}, config.MetricsLogger.Log(&api.MetricsEvent{
			Event:    newInteractionEventText,
			Duration: time.Since(timeStartNewInteraction),
		})
}

// SupportedCredential represents a specific credential (type and format) that an issuer can issue.
type SupportedCredential struct {
	Format string
	Types  []string
}

// SupportedCredentials returns the credential types and formats that an issuer can issue.
func (i *WalletInitiatedInteraction) SupportedCredentials() ([]SupportedCredential, error) {
	err := i.interaction.populateIssuerMetadata("Get supported credentials")
	if err != nil {
		return nil, err
	}

	credentialConf := i.interaction.issuerMetadata.CredentialConfigurationsSupported
	supportedCredentials := make([]SupportedCredential, 0, len(credentialConf))

	for _, credentialConfiguration := range credentialConf {
		supportedCredentials = append(supportedCredentials, SupportedCredential{
			Format: credentialConfiguration.Format,
			Types:  credentialConfiguration.CredentialDefinition.Type,
		})
	}

	return supportedCredentials, nil
}

// CreateAuthorizationURL creates an authorization URL that can be opened in a browser to proceed to the login page.
// It must be called before calling the RequestCredential method.
// It creates the authorization URL that can be opened in a browser to proceed to the login page.
// This method can only be used if the issuer supports authorization code grants.
// Check the issuer's capabilities first using the Capabilities method.
// If scopes are needed, pass them in using the WithScopes option.
func (i *WalletInitiatedInteraction) CreateAuthorizationURL(clientID, redirectURI, credentialFormat string,
	credentialTypes []string, opts ...CreateAuthorizationURLOpt,
) (string, error) {
	processedOpts := processCreateAuthorizationURLOpts(opts)

	authorizationURL, err := i.interaction.createAuthorizationURL(clientID, redirectURI, credentialFormat,
		credentialTypes, processedOpts.context, processedOpts.issuerState, processedOpts.scopes,
		processedOpts.useOAuthDiscoverableClientIDScheme)
	if err != nil {
		return "", err
	}

	i.credentialFormat = credentialFormat
	i.credentialTypes = credentialTypes
	i.credentialContext = processedOpts.context

	return authorizationURL, nil
}

// RequestCredential requests credential(s) from the issuer. This method is the final step in the
// interaction with the issuer.
// This method must be called only once all authorization pre-requisite steps have been completed.
// The redirect URI that you pass in here should look like the redirect URI that you passed in to the
// CreateAuthorizationURL, except that now it has some URL query parameters appended to it.
func (i *WalletInitiatedInteraction) RequestCredential(jwtSigner api.JWTSigner, redirectURIWithParams string,
) ([]*verifiable.Credential, error) {
	err := i.interaction.requestAccessToken(redirectURIWithParams)
	if err != nil {
		return nil, err
	}

	return i.interaction.requestCredentialWithAuth(jwtSigner, []string{i.credentialFormat},
		[][]string{i.credentialTypes}, [][]string{i.credentialContext})
}

// DynamicClientRegistrationSupported indicates whether the issuer supports dynamic client registration.
func (i *WalletInitiatedInteraction) DynamicClientRegistrationSupported() (bool, error) {
	return i.interaction.dynamicClientRegistrationSupported()
}

// DynamicClientRegistrationEndpoint returns the issuer's dynamic client registration endpoint.
// The caller should call the DynamicClientRegistrationSupported method first and only call this method
// if DynamicClientRegistrationSupported returns true.
// This method will return an error if the issuer does not support dynamic client registration.
func (i *WalletInitiatedInteraction) DynamicClientRegistrationEndpoint() (string, error) {
	return i.interaction.dynamicClientRegistrationEndpoint()
}

// IssuerMetadata returns the issuer's metadata.
func (i *WalletInitiatedInteraction) IssuerMetadata() (*issuer.Metadata, error) {
	err := i.interaction.populateIssuerMetadata(getIssuerMetadataEventText)
	if err != nil {
		return nil, err
	}

	return i.interaction.issuerMetadata, nil
}

// VerifyIssuer verifies the issuer via its issuer metadata. If successful, then the service URL is returned.
// An error means that either the issuer failed the verification check, or something went wrong during the
// process (and so a verification status could not be determined).
func (i *WalletInitiatedInteraction) VerifyIssuer() (string, error) {
	return i.interaction.verifyIssuer()
}
