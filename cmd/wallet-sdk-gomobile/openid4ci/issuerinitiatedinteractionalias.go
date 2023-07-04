/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
)

// This Interaction object in this file is an alias for an IssuerInitiatedInteraction object and performs the same
// functionality. This only exists for backwards compatibility with code that hasn't been updated to the latest version
// of Wallet-SDK and will be removed in a future version.
// Gomobile doesn't seem to work correctly when Go's built-in type alias feature, so instead the Interaction type here
// acts like a pass-through wrapper for an IssuerInitiatedInteraction.

// Interaction is an alias for IssuerInitiatedInteraction.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
type Interaction struct {
	issuerInitiatedInteraction *IssuerInitiatedInteraction
}

// InteractionArgs is an alias for IssuerInitiatedInteractionArgs.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
type InteractionArgs struct {
	initiateIssuanceURI string
	crypto              api.Crypto
	didResolver         api.DIDResolver
}

// NewInteractionArgs is an alias for NewIssuerInitiatedInteractionArgs.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func NewInteractionArgs(initiateIssuanceURI string, crypto api.Crypto,
	didResolver api.DIDResolver,
) *InteractionArgs {
	return &InteractionArgs{
		initiateIssuanceURI: initiateIssuanceURI,
		crypto:              crypto,
		didResolver:         didResolver,
	}
}

// NewInteraction is an alias for NewIssuerInitiatedInteraction.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func NewInteraction(args *InteractionArgs, opts *InteractionOpts) (*Interaction, error) {
	issuerInitiatedInteractionArgs := NewIssuerInitiatedInteractionArgs(args.initiateIssuanceURI, args.crypto,
		args.didResolver)

	issuerInitiatedInteraction, err := NewIssuerInitiatedInteraction(issuerInitiatedInteractionArgs, opts)
	if err != nil {
		return nil, err
	}

	return &Interaction{issuerInitiatedInteraction: issuerInitiatedInteraction}, nil
}

// CreateAuthorizationURL is an alias for the method with the same name on the IssuerInitiatedInteraction object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) CreateAuthorizationURL(clientID, redirectURI string,
	opts *CreateAuthorizationURLOpts,
) (string, error) {
	return i.issuerInitiatedInteraction.CreateAuthorizationURL(clientID, redirectURI, opts)
}

// RequestCredentialWithPreAuth is an alias for the method with the same name on the IssuerInitiatedInteraction object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) RequestCredentialWithPreAuth(
	vm *api.VerificationMethod, opts *RequestCredentialWithPreAuthOpts,
) (*verifiable.CredentialsArray, error) {
	return i.issuerInitiatedInteraction.RequestCredentialWithPreAuth(vm, opts)
}

// RequestCredentialWithAuth is an alias for the method with the same name on the IssuerInitiatedInteraction object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) RequestCredentialWithAuth(vm *api.VerificationMethod,
	redirectURIWithAuthCode string, opts *RequestCredentialWithAuthOpts,
) (*verifiable.CredentialsArray, error) {
	return i.issuerInitiatedInteraction.RequestCredentialWithAuth(vm, redirectURIWithAuthCode, opts)
}

// RequestCredential is equivalent to calling the RequestCredentialWithPreAuth method with no PIN set.
//
// Deprecated: This method exists only to prevent existing code from breaking. It will be removed in a future version
// as the method name can be confusing.
func (i *Interaction) RequestCredential(vm *api.VerificationMethod) (*verifiable.CredentialsArray, error) {
	return i.issuerInitiatedInteraction.RequestCredentialWithPreAuth(vm, nil)
}

// RequestCredentialWithPIN is equivalent to calling the RequestCredentialWithPreAuth method with a PIN set.
//
// Deprecated: This method exists only to prevent existing code from breaking. It will be removed in a future version
// as the method name can be confusing and the lack of an options object means it's not possible to easily add new
// options in the future without breaking existing code.
func (i *Interaction) RequestCredentialWithPIN(
	vm *api.VerificationMethod, pin string,
) (*verifiable.CredentialsArray, error) {
	return i.issuerInitiatedInteraction.RequestCredentialWithPreAuth(vm,
		NewRequestCredentialWithPreAuthOpts().SetPIN(pin))
}

// IssuerURI is an alias for the method with the same name on the IssuerInitiatedInteraction object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) IssuerURI() string {
	return i.issuerInitiatedInteraction.IssuerURI()
}

// PreAuthorizedCodeGrantTypeSupported is an alias for the method with the same name on the IssuerInitiatedInteraction
// object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) PreAuthorizedCodeGrantTypeSupported() bool {
	return i.issuerInitiatedInteraction.PreAuthorizedCodeGrantTypeSupported()
}

// PreAuthorizedCodeGrantParams is an alias for the method with the same name on the IssuerInitiatedInteraction object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) PreAuthorizedCodeGrantParams() (*PreAuthorizedCodeGrantParams, error) {
	return i.issuerInitiatedInteraction.PreAuthorizedCodeGrantParams()
}

// AuthorizationCodeGrantTypeSupported is an alias for the method with the same name on the IssuerInitiatedInteraction
// object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) AuthorizationCodeGrantTypeSupported() bool {
	return i.issuerInitiatedInteraction.AuthorizationCodeGrantTypeSupported()
}

// AuthorizationCodeGrantParams is an alias for the method with the same name on the IssuerInitiatedInteraction object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) AuthorizationCodeGrantParams() (*AuthorizationCodeGrantParams, error) {
	return i.issuerInitiatedInteraction.AuthorizationCodeGrantParams()
}

// DynamicClientRegistrationSupported is an alias for the method with the same name on the IssuerInitiatedInteraction
// object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) DynamicClientRegistrationSupported() (bool, error) {
	return i.issuerInitiatedInteraction.DynamicClientRegistrationSupported()
}

// DynamicClientRegistrationEndpoint is an alias for the method with the same name on the IssuerInitiatedInteraction
// object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) DynamicClientRegistrationEndpoint() (string, error) {
	return i.issuerInitiatedInteraction.DynamicClientRegistrationEndpoint()
}

// OTelTraceID is an alias for the method with the same name on the IssuerInitiatedInteraction object.
//
// Deprecated: This only exists for backwards compatibility with code that hasn't been updated to the latest version of
// Wallet-SDK and will be removed in a future version.
func (i *Interaction) OTelTraceID() string {
	return i.issuerInitiatedInteraction.OTelTraceID()
}
