/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
)

// Interaction represents a single OpenID4CI interaction between a wallet and an issuer. The methods defined on this
// object are used to help guide the calling code through the OpenID4CI flow.
type Interaction struct {
	initiationRequest *InitiationRequest
}

// InitiationRequest represents the Issuance Initiation Request object received from an Issuer as defined in
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-5.1.
type InitiationRequest struct {
	Issuer            string   `json:"issuer,omitempty"`
	CredentialTypes   []string `json:"credential_type,omitempty"`     //nolint: tagliatelle // required by spec
	PreAuthorizedCode string   `json:"pre-authorized_code,omitempty"` //nolint: tagliatelle // required by spec
	UserPINRequired   bool     `json:"user_pin_required,omitempty"`   //nolint: tagliatelle // required by spec
	OpState           string   `json:"op_state,omitempty"`            //nolint: tagliatelle // required by spec
}

// AuthorizeResult is the object returned from the Client.Authorize method.
// An empty/missing AuthorizationRedirectEndpoint indicates that the wallet is pre-authorized.
type AuthorizeResult struct {
	AuthorizationRedirectEndpoint string
	UserPINRequired               bool
}

// CredentialRequest represents the data (required and optional) that is used in the final step of the OpenID4CI flow,
// where the wallet requests the credential from the issuer.
type CredentialRequest struct {
	UserPIN string
}

// CredentialResponse is the object returned from the Client.Callback method.
// It contains the issued credentials and their respective formats.
type CredentialResponse struct {
	Credential *verifiable.Credential `json:"credential,omitempty"` // Optional for deferred credential flow
	Format     string                 `json:"format,omitempty"`
}
