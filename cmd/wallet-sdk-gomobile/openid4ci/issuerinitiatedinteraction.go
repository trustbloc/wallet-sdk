/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4ci provides APIs for wallets to receive verifiable credentials via OIDC for Credential Issuance.
package openid4ci

import (
	"errors"
	"fmt"

	verifiableapi "github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/otel"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
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
	credentials, _, err := i.requestCredentialWithPreAuth(vm, opts)
	if err != nil {
		return nil, err
	}

	return toGomobileCredentials(credentials), nil
}

// RequestCredentialWithPreAuthV2 requests credentials using a pre-authorized code flow.
// Returns an array of credentials with config IDs, which map to CredentialConfigurationSupported in the
// issuer's metadata.
func (i *IssuerInitiatedInteraction) RequestCredentialWithPreAuthV2(
	vm *api.VerificationMethod, opts *RequestCredentialWithPreAuthOpts,
) (*verifiable.CredentialsArrayV2, error) {
	credentials, configIDs, err := i.requestCredentialWithPreAuth(vm, opts)
	if err != nil {
		return nil, err
	}

	return toGomobileCredentialsV2(credentials, configIDs), nil
}

func (i *IssuerInitiatedInteraction) requestCredentialWithPreAuth(
	vm *api.VerificationMethod, opts *RequestCredentialWithPreAuthOpts,
) ([]*verifiableapi.Credential, []string, error) {
	if opts == nil {
		opts = NewRequestCredentialWithPreAuthOpts()
	}

	signer, err := createSigner(vm, i.crypto)
	if err != nil {
		return nil, nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	goOpts := []openid4cigoapi.RequestCredentialWithPreAuthOpt{openid4cigoapi.WithPIN(opts.pin)}

	if opts.attestationVM != nil {
		attestationSigner, attErr := createSigner(opts.attestationVM, i.crypto)
		if attErr != nil {
			return nil, nil, wrapper.ToMobileErrorWithTrace(attErr, i.oTel)
		}

		goOpts = append(goOpts, openid4cigoapi.WithAttestationVC(attestationSigner, opts.attestationVC))
	}

	credentials, err := i.goAPIInteraction.RequestCredentialWithPreAuth(signer, goOpts...)
	if err != nil {
		return nil, nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	configIDs := i.goAPIInteraction.CredentialConfigIDs()

	if len(credentials) != len(configIDs) {
		return nil, nil, fmt.Errorf("mismatch in the number of credentials and configuration IDs: "+
			"expected %d but got %d", len(credentials), len(configIDs))
	}

	return credentials, configIDs, nil
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

// OfferedCredentialsTypes returns types of offered credentials.
func (i *IssuerInitiatedInteraction) OfferedCredentialsTypes() *api.StringArrayArray {
	return api.StringArrayArrayFromGoArray(i.goAPIInteraction.OfferedCredentialsTypes())
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

// IssuerTrustInfo returns issuer trust info like, did, domain, credential type, format.
func (i *IssuerInitiatedInteraction) IssuerTrustInfo() (*IssuerTrustInfo, error) {
	trustInfo, err := i.goAPIInteraction.IssuerTrustInfo()
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, i.oTel)
	}

	var credentialOffers []*CredentialOffer

	for _, offer := range trustInfo.CredentialsSupported {
		for _, t := range offer.Types {
			if t != "VerifiableCredential" {
				credentialOffers = append(credentialOffers, &CredentialOffer{
					CredentialType:   t,
					CredentialFormat: offer.Format,
				})
			}
		}
	}

	return &IssuerTrustInfo{
		DID:              trustInfo.DID,
		Domain:           trustInfo.Domain,
		CredentialOffers: credentialOffers,
	}, nil
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

// Acknowledgment return not nil Acknowledgment if the issuer requires to be acknowledged that
// the user accepts or rejects credentials.
func (i *IssuerInitiatedInteraction) Acknowledgment() (*Acknowledgment, error) {
	acknowledgment, err := i.goAPIInteraction.Acknowledgment()
	if err != nil {
		return nil, err
	}

	return &Acknowledgment{
		acknowledgment: acknowledgment,
	}, nil
}
