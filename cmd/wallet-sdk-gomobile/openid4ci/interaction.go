/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4ci provides APIs for wallets to receive verifiable credentials via OIDC for Credential Issuance.
package openid4ci

import (
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/util/didsignjwt"

	"github.com/trustbloc/wallet-sdk/cmd/utilities/gomobilewrappers"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	openid4cigoapi "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

// Interaction represents a single OpenID4CI interaction between a wallet and an issuer. The methods defined on this
// object are used to help guide the calling code through the OpenID4CI flow.
type Interaction struct {
	goAPIInteraction *openid4cigoapi.Interaction
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

// ClientConfig contains the various required parameters for an OpenID4CI Interaction.
type ClientConfig struct {
	UserDID       string
	ClientID      string
	SignerCreator api.DIDJWTSignerCreator
	DIDResolver   api.DIDResolver
}

// NewClientConfig creates the client config object.
func NewClientConfig(userDID, clientID string, signerCreator api.DIDJWTSignerCreator,
	didRes api.DIDResolver,
) *ClientConfig {
	return &ClientConfig{
		UserDID:       userDID,
		ClientID:      clientID,
		SignerCreator: signerCreator,
		DIDResolver:   didRes,
	}
}

// NewInteraction creates a new OpenID4CI Interaction.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// Calling this function represents taking the first step in the flow.
// This function takes in an Initiate Issuance Request object from an issuer (as defined in
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-5.1), encoded using URL query
// parameters. This object is intended for going through the full flow only once (i.e. one interaction), after which
// it should be discarded. Any new interactions should use a fresh Interaction instance.
func NewInteraction(
	initiateIssuanceURI string, config *ClientConfig,
) (*Interaction, error) {
	goAPIClientConfig := unwrapConfig(config)

	goAPIInteraction, err := openid4cigoapi.NewInteraction(initiateIssuanceURI, goAPIClientConfig)
	if err != nil {
		return nil, err
	}

	return &Interaction{goAPIInteraction: goAPIInteraction}, nil
}

// Authorize is used by a wallet to authorize an issuer's OIDC Verifiable Credential Issuance Request.
// After initializing the Interaction object with an Issuance Request, this should be the first method you call in
// order to continue with the flow.
// It only supports the pre-authorized flow in its current implementation.
// Once the authorization flow is implemented, the following section of the spec will be relevant:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-6
func (i *Interaction) Authorize() (*AuthorizeResult, error) {
	authorizationResultGoAPI, err := i.goAPIInteraction.Authorize()
	if err != nil {
		return nil, err
	}

	authorizationResult := &AuthorizeResult{
		AuthorizationRedirectEndpoint: authorizationResultGoAPI.AuthorizationRedirectEndpoint,
		UserPINRequired:               authorizationResultGoAPI.UserPINRequired,
	}

	return authorizationResult, nil
}

// RequestCredential is the final step (or second last step, if the ResolveDisplay method isn't needed) in the
// interaction. This is called after the wallet is authorized and is ready to receive credential(s).
// Relevant sections of the spec:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-7
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-8
// The returned object is an array of Credential Responses (as a JSON array).
func (i *Interaction) RequestCredential(
	credentialRequest *CredentialRequestOpts,
) (*api.VerifiableCredentialsArray, error) {
	goAPICredentialRequest := &openid4cigoapi.CredentialRequestOpts{UserPIN: credentialRequest.UserPIN}

	credentialResponses, err := i.goAPIInteraction.RequestCredential(goAPICredentialRequest)
	if err != nil {
		return nil, err
	}

	result := api.NewVerifiableCredentialsArray()

	for _, response := range credentialResponses {
		result.Add(api.NewVerifiableCredential(response.Credential))
	}

	return result, nil
}

// ResolveDisplay is the optional final step that can be called after RequestCredential. It resolves display
// information for the credentials received in this interaction. The CredentialDisplays in the returned
// object correspond to the VCs received and are in the same order.
// If preferredLocale is not specified, then the first locale specified by the issuer's metadata will be used during
// resolution.
func (i *Interaction) ResolveDisplay(preferredLocale string) (*api.JSONObject, error) {
	resolvedDisplayData, err := i.goAPIInteraction.ResolveDisplay(preferredLocale)
	if err != nil {
		return nil, err
	}

	resolvedDisplayDataBytes, err := json.Marshal(resolvedDisplayData)
	if err != nil {
		return nil, err
	}

	return &api.JSONObject{Data: resolvedDisplayDataBytes}, nil
}

func unwrapConfig(config *ClientConfig) *openid4cigoapi.ClientConfig {
	goAPISignerGetter := func(vm *did.VerificationMethod) (didsignjwt.Signer, error) {
		vmBytes, err := workaroundMarshalVM(vm)
		if err != nil {
			return nil, err
		}

		goMobileSigner, err := config.SignerCreator.Create(&api.JSONObject{Data: vmBytes})
		if err != nil {
			return nil, fmt.Errorf("failed to create gomobile signer: %w", err)
		}

		return goMobileSigner, nil
	}

	return &openid4cigoapi.ClientConfig{
		UserDID:        config.UserDID,
		ClientID:       config.ClientID,
		SignerProvider: goAPISignerGetter,
		DIDResolver:    &gomobilewrappers.VDRResolverWrapper{DIDResolver: config.DIDResolver},
	}
}

func workaroundMarshalVM(vm *did.VerificationMethod) ([]byte, error) {
	rawVM := map[string]interface{}{
		"id":         vm.ID,
		"type":       vm.Type,
		"controller": vm.Controller,
	}

	jsonKey := vm.JSONWebKey()
	if jsonKey != nil {
		jwkBytes, err := jsonKey.MarshalJSON()
		if err != nil {
			return nil, err
		}

		rawVM["publicKeyJwk"] = json.RawMessage(jwkBytes)
	} else {
		rawVM["publicKeyBase58"] = base58.Encode(vm.Value)
	}

	return json.Marshal(rawVM)
}
