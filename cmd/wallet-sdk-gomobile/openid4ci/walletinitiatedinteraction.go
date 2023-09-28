/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/pkg/walleterror"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/otel"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

// WalletInitiatedInteraction represents a single wallet-instantiated OpenID4CI interaction between a wallet and an
// issuer.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// An WalletInitiatedInteraction is a stateful object, and is intended for going through the full flow only once
// after which it should be discarded. Any new interactions should use a fresh WalletInitiatedInteraction instance.
type WalletInitiatedInteraction struct {
	goAPIInteraction *openid4cigoapi.WalletInitiatedInteraction
	crypto           api.Crypto
	oTel             *otel.Trace
}

// WalletInitiatedInteractionArgs contains the required parameters for an WalletInitiatedInteraction.
type WalletInitiatedInteractionArgs struct {
	issuerURI   string
	crypto      api.Crypto
	didResolver api.DIDResolver
}

// NewWalletInitiatedInteractionArgs creates a new WalletInitiatedInteractionArgs object. All parameters are mandatory.
func NewWalletInitiatedInteractionArgs(issuerURI string, crypto api.Crypto,
	didResolver api.DIDResolver,
) *WalletInitiatedInteractionArgs {
	return &WalletInitiatedInteractionArgs{
		issuerURI:   issuerURI,
		crypto:      crypto,
		didResolver: didResolver,
	}
}

// NewWalletInitiatedInteraction creates a new OpenID4CI WalletInitiatedInteraction.
func NewWalletInitiatedInteraction( //nolint: dupl // Similar looking but for different objects with different uses
	args *WalletInitiatedInteractionArgs,
	opts *InteractionOpts,
) (*WalletInitiatedInteraction, error) {
	if args == nil {
		return nil, wrapper.ToMobileError(walleterror.NewInvalidSDKUsageError(
			openid4cigoapi.ErrorModule, errors.New("args object must be provided")))
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

	goAPIClientConfig, err := createGoAPIClientConfig(args.didResolver, opts)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, oTel)
	}

	goAPIInteraction, err := openid4cigoapi.NewWalletInitiatedInteraction(args.issuerURI, goAPIClientConfig)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, oTel)
	}

	return &WalletInitiatedInteraction{
		crypto:           args.crypto,
		goAPIInteraction: goAPIInteraction,
		oTel:             oTel,
	}, nil
}

// CreateAuthorizationURL creates an authorization URL that can be opened in a browser to proceed to the login page.
// It must be called before calling the RequestCredential method.
// It creates the authorization URL that can be opened in a browser to proceed to the login page.
// This method can only be used if the issuer supports authorization code grants.
// Check the issuer's capabilities first using the Capabilities method.
// If scopes are needed, pass them in using the CreateAuthorizationURLOpts object.
func (i *WalletInitiatedInteraction) CreateAuthorizationURL(clientID, redirectURI, credentialFormat string,
	credentialTypes *api.StringArray, opts *CreateAuthorizationURLOpts,
) (string, error) {
	goAPIOpts := convertToGoAPICreateAuthURLOpts(opts)

	if credentialTypes == nil {
		credentialTypes = api.NewStringArray()
	}

	authorizationURL, err := i.goAPIInteraction.CreateAuthorizationURL(clientID, redirectURI, credentialFormat,
		credentialTypes.Strings, goAPIOpts...)
	if err != nil {
		return "", wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return authorizationURL, nil
}

// RequestCredential requests credential(s) from the issuer. This method is the final step in the
// interaction with the issuer.
// This method must be called only once all authorization pre-requisite steps have been completed.
// The redirect URI that you pass in here should look like the redirect URI that you passed in to the
// CreateAuthorizationURL, except that now it has some URL query parameters appended to it.
func (i *WalletInitiatedInteraction) RequestCredential(vm *api.VerificationMethod,
	redirectURIWithAuthCode string,
	opts *RequestCredentialWithAuthOpts, //nolint: revive // The opts param is reserved for future use.
) (*verifiable.CredentialsArray, error) {
	signer, err := createSigner(vm, i.crypto)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	credentials, err := i.goAPIInteraction.RequestCredential(signer, redirectURIWithAuthCode)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return toGomobileCredentials(credentials), nil
}

// DynamicClientRegistrationSupported indicates whether the issuer supports dynamic client registration.
func (i *WalletInitiatedInteraction) DynamicClientRegistrationSupported() (bool, error) {
	supported, err := i.goAPIInteraction.DynamicClientRegistrationSupported()
	if err != nil {
		return false, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return supported, nil
}

// DynamicClientRegistrationEndpoint returns the issuer's dynamic client registration endpoint.
// The caller should call the DynamicClientRegistrationSupported method first and only call this method
// if DynamicClientRegistrationSupported returns true.
// This method will return an error if the issuer does not support dynamic client registration.
func (i *WalletInitiatedInteraction) DynamicClientRegistrationEndpoint() (string, error) {
	endpoint, err := i.goAPIInteraction.DynamicClientRegistrationEndpoint()
	if err != nil {
		return "", wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return endpoint, nil
}

// IssuerMetadata returns the issuer's metadata object.
func (i *WalletInitiatedInteraction) IssuerMetadata() (*IssuerMetadata, error) {
	goAPIIssuerMetadata, err := i.goAPIInteraction.IssuerMetadata()
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return &IssuerMetadata{issuerMetadata: goAPIIssuerMetadata}, nil
}

// VerifyIssuer verifies the issuer via its issuer metadata. If successful, then the service URL is returned.
// An error means that either the issuer failed the verification check, or something went wrong during the
// process (and so a verification status could not be determined).
func (i *WalletInitiatedInteraction) VerifyIssuer() (string, error) {
	serviceURL, err := i.goAPIInteraction.VerifyIssuer()
	if err != nil {
		return "", wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return serviceURL, nil
}
