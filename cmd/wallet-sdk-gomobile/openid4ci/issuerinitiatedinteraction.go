/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4ci provides APIs for wallets to receive verifiable credentials via OIDC for Credential Issuance.
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

// IssuerInitiatedInteraction represents a single issuer-instantiated OpenID4CI interaction between a wallet and an
// issuer. This type can be used if you have received a credential offer from an issuer in some form.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// An IssuerInitiatedInteraction is a stateful object, and is intended for going through the full flow only once
// after which it should be discarded. Any new interactions should use a fresh IssuerInitiatedInteraction instance.
type IssuerInitiatedInteraction struct {
	goAPIInteraction *openid4cigoapi.IssuerInitiatedInteraction
	crypto           api.Crypto
	oTel             *otel.Trace
}

// NewIssuerInitiatedInteraction creates a new OpenID4CI IssuerInitiatedInteraction.
func NewIssuerInitiatedInteraction( //nolint: dupl // Similar looking but for different objects with different uses
	args *IssuerInitiatedInteractionArgs,
	opts *InteractionOpts,
) (*IssuerInitiatedInteraction, error) {
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

	goAPIInteraction, err := openid4cigoapi.NewIssuerInitiatedInteraction(args.initiateIssuanceURI, goAPIClientConfig)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, oTel)
	}

	return &IssuerInitiatedInteraction{
		crypto:           args.crypto,
		goAPIInteraction: goAPIInteraction,
		oTel:             oTel,
	}, nil
}

// CreateAuthorizationURL creates an authorization URL that can be opened in a browser to proceed to the login page.
// It is the first step in the authorization code flow.
// It creates the authorization URL that can be opened in a browser to proceed to the login page.
// This method can only be used if the issuer supports authorization code grants.
// Check the issuer's capabilities first using the methods available on this IssuerInitiatedInteraction object.
// If scopes are needed, pass them in using the CreateAuthorizationURLOpts object.
func (i *IssuerInitiatedInteraction) CreateAuthorizationURL(clientID, redirectURI string,
	opts *CreateAuthorizationURLOpts,
) (string, error) {
	goAPIOpts := convertToGoAPICreateAuthURLOpts(opts)

	authorizationURL, err := i.goAPIInteraction.CreateAuthorizationURL(clientID, redirectURI, goAPIOpts...)
	if err != nil {
		return "", wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return authorizationURL, nil
}

// RequestCredentialWithPreAuth requests credential(s) from the issuer. This method can only be used for the
// pre-authorized code flow, where it acts as the final step in the interaction with the issuer.
// For the equivalent method for the authorization code flow, see RequestCredentialWithAuth instead.
// If a PIN is required (which can be checked via the Capabilities method), then it must be passed
// into this method via the SetPIN method on the RequestCredentialWithPreAuthOpts object.
func (i *IssuerInitiatedInteraction) RequestCredentialWithPreAuth(
	vm *api.VerificationMethod, opts *RequestCredentialWithPreAuthOpts,
) (*verifiable.CredentialsArray, error) {
	if opts == nil {
		opts = NewRequestCredentialWithPreAuthOpts()
	}

	signer, err := createSigner(vm, i.crypto)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
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
func (i *IssuerInitiatedInteraction) RequestCredentialWithAuth(vm *api.VerificationMethod,
	redirectURIWithAuthCode string,
	opts *RequestCredentialWithAuthOpts, //nolint: revive // The opts param is reserved for future use.
) (*verifiable.CredentialsArray, error) {
	signer, err := createSigner(vm, i.crypto)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
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
func (i *IssuerInitiatedInteraction) RequestCredential(
	vm *api.VerificationMethod,
) (*verifiable.CredentialsArray, error) {
	return i.RequestCredentialWithPreAuth(vm, nil)
}

// RequestCredentialWithPIN is equivalent to calling the RequestCredentialWithPreAuth method with a PIN set.
//
// Deprecated: This method exists only to prevent existing code from breaking. It will be removed in a future version
// as the method name can be confusing and the lack of an options object means it's not possible to easily add new
// options in the future without breaking existing code.
func (i *IssuerInitiatedInteraction) RequestCredentialWithPIN(
	vm *api.VerificationMethod, pin string,
) (*verifiable.CredentialsArray, error) {
	return i.RequestCredentialWithPreAuth(vm, NewRequestCredentialWithPreAuthOpts().SetPIN(pin))
}

// IssuerURI returns the issuer's URI from the initiation request. It's useful to store this somewhere in case
// there's a later need to refresh credential display data using the latest display information from the issuer.
func (i *IssuerInitiatedInteraction) IssuerURI() string {
	return i.goAPIInteraction.IssuerURI()
}

// PreAuthorizedCodeGrantTypeSupported indicates whether an issuer supports the pre-authorized code grant type.
func (i *IssuerInitiatedInteraction) PreAuthorizedCodeGrantTypeSupported() bool {
	return i.goAPIInteraction.PreAuthorizedCodeGrantTypeSupported()
}

// PreAuthorizedCodeGrantParams returns an object that can be used to determine an issuer's pre-authorized code grant
// parameters. The caller should call the PreAuthorizedCodeGrantTypeSupported method first and only call this method to
// get the params if PreAuthorizedCodeGrantTypeSupported returns true.
// This method returns an error if (and only if) PreAuthorizedCodeGrantTypeSupported returns false.
func (i *IssuerInitiatedInteraction) PreAuthorizedCodeGrantParams() (*PreAuthorizedCodeGrantParams, error) {
	goAPIPreAuthorizedCodeGrantParams, err := i.goAPIInteraction.PreAuthorizedCodeGrantParams()
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return &PreAuthorizedCodeGrantParams{
		goAPIPreAuthorizedCodeGrantParams: goAPIPreAuthorizedCodeGrantParams,
	}, nil
}

// AuthorizationCodeGrantTypeSupported indicates whether an issuer supports the authorization code grant type.
func (i *IssuerInitiatedInteraction) AuthorizationCodeGrantTypeSupported() bool {
	return i.goAPIInteraction.AuthorizationCodeGrantTypeSupported()
}

// AuthorizationCodeGrantParams returns an object that can be used to determine the issuer's authorization code grant
// parameters. The caller should call the AuthorizationCodeGrantTypeSupported method first and only call this method to
// get the params if AuthorizationCodeGrantTypeSupported returns true.
// This method returns an error if (and only if) AuthorizationCodeGrantTypeSupported returns false.
func (i *IssuerInitiatedInteraction) AuthorizationCodeGrantParams() (*AuthorizationCodeGrantParams, error) {
	goAPIAuthorizationCodeGrantParams, err := i.goAPIInteraction.AuthorizationCodeGrantParams()
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return &AuthorizationCodeGrantParams{
		goAPIAuthorizationCodeGrantParams: goAPIAuthorizationCodeGrantParams,
		oTel:                              i.oTel,
	}, nil
}

// DynamicClientRegistrationSupported indicates whether the issuer supports dynamic client registration.
func (i *IssuerInitiatedInteraction) DynamicClientRegistrationSupported() (bool, error) {
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
func (i *IssuerInitiatedInteraction) DynamicClientRegistrationEndpoint() (string, error) {
	endpoint, err := i.goAPIInteraction.DynamicClientRegistrationEndpoint()
	if err != nil {
		return "", wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return endpoint, nil
}

// IssuerMetadata returns the issuer's metadata.
func (i *IssuerInitiatedInteraction) IssuerMetadata() (*IssuerMetadata, error) {
	goAPIIssuerMetadata, err := i.goAPIInteraction.IssuerMetadata()
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return &IssuerMetadata{issuerMetadata: goAPIIssuerMetadata}, nil
}

// VerifyIssuer verifies the issuer via its issuer metadata. If successful, then the service URL is returned.
// An error means that either the issuer failed the verification check, or something went wrong during the
// process (and so a verification status could not be determined).
func (i *IssuerInitiatedInteraction) VerifyIssuer() (string, error) {
	serviceURL, err := i.goAPIInteraction.VerifyIssuer()
	if err != nil {
		return "", wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	return serviceURL, nil
}

// OTelTraceID returns the OpenTelemetry trace ID.
// If OpenTelemetry has been disabled, then an empty string is returned.
func (i *IssuerInitiatedInteraction) OTelTraceID() string {
	traceID := ""
	if i.oTel != nil {
		traceID = i.oTel.TraceID()
	}

	return traceID
}

// RequireAcknowledgment if true indicates that the issuer requires to be acknowledged if
// the user accepts or rejects credentials.
func (i *IssuerInitiatedInteraction) RequireAcknowledgment() (bool, error) {
	return i.goAPIInteraction.RequireAcknowledgment()
}

// AcknowledgeSuccess acknowledges the issuer that the client accepted credentials.
func (i *IssuerInitiatedInteraction) AcknowledgeSuccess() error {
	return i.goAPIInteraction.AcknowledgeSuccess()
}

// AcknowledgeReject acknowledges the issuer that the client rejected credentials.
func (i *IssuerInitiatedInteraction) AcknowledgeReject() error {
	return i.goAPIInteraction.AcknowledgeReject()
}
