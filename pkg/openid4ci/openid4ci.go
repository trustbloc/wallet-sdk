/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4ci provides APIs for wallets to receive verifiable credentials via OIDC for Credential Issuance.
package openid4ci

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
)

// NewInteraction creates a new OpenID4CI Interaction.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// Calling this function represents taking the first step in the flow.
// This function takes in an Initiate Issuance Request object from an Issuer (as defined in
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-5.1), encoded using URL query
// parameters. This object is intended for going through the full flow only once (i.e. one interaction), after which
// it should be discarded. Any new interactions should use a fresh Interaction instance.
func NewInteraction(requestURI string) (*Interaction, error) {
	requestURIParsed, err := url.Parse(requestURI)
	if err != nil {
		return nil, err
	}

	initiationRequest := &InitiationRequest{}

	initiationRequest.Issuer = requestURIParsed.Query().Get("issuer")
	initiationRequest.CredentialTypes = requestURIParsed.Query()["credential_type"]
	initiationRequest.PreAuthorizedCode = requestURIParsed.Query().Get("pre-authorized_code")

	userPINRequiredString := requestURIParsed.Query().Get("user_pin_required")

	if userPINRequiredString != "" {
		userPINRequired, err := strconv.ParseBool(userPINRequiredString)
		if err != nil {
			return nil, err
		}

		initiationRequest.UserPINRequired = userPINRequired
	}

	initiationRequest.OpState = requestURIParsed.Query().Get("op_state")

	return &Interaction{initiationRequest: initiationRequest}, nil
}

// Authorize is used by a wallet to authorize an issuer's OIDC Verifiable Credential Issuance Request.
// After initializing the Interaction object with an Issuance Request, this should be the first method you call in
// order to continue with the flow.
// It only supports the pre-authorized flow in its current implementation.
// Once the authorization flow is implemented, the following section of the spec will be relevant:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-6
func (i *Interaction) Authorize() (*AuthorizeResult, error) {
	if i.initiationRequest.PreAuthorizedCode == "" {
		return nil, errors.New("pre-authorized code is required (authorization flow not implemented)")
	}

	authorizeResult := &AuthorizeResult{
		UserPINRequired: i.initiationRequest.UserPINRequired,
	}

	return authorizeResult, nil
}

// RequestCredential is the final step in the interaction. This is called after the wallet is authorized and is ready
// to receive credential(s).
// Relevant sections of the spec:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-7
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-8
func (i *Interaction) RequestCredential(credentialRequest *CredentialRequest) ([]*CredentialResponse, error) {
	if i.initiationRequest.UserPINRequired && credentialRequest.UserPIN == "" {
		return nil, errors.New("PIN required (per initiation request)")
	}

	return []*CredentialResponse{getSampleCredentialResponse()}, nil
}

func getSampleCredentialResponse() *CredentialResponse {
	return &CredentialResponse{
		Credential: &verifiable.Credential{ID: "SampleID"},
		Format:     "SampleFormat",
	}
}
