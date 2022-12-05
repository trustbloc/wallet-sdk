/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/util/didsignjwt"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

// Interaction represents a single OpenID4CI interaction between a wallet and an issuer. The methods defined on this
// object are used to help guide the calling code through the OpenID4CI flow.
type Interaction struct {
	initiationRequest *InitiationRequest
	userDID           string
	clientID          string
	issuerMetadata    *issuer.Metadata
	vcs               []string // base64url encoded
	signerProvider    didsignjwt.SignerGetter
	didResolver       *didResolverWrapper
}

// InitiationRequest represents the Issuance Initiation Request object received from an issuer as defined in
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-5.1.
type InitiationRequest struct {
	IssuerURI         string   `json:"issuer,omitempty"`
	CredentialTypes   []string `json:"credential_type,omitempty"`
	PreAuthorizedCode string   `json:"pre-authorized_code,omitempty"`
	UserPINRequired   bool     `json:"user_pin_required,omitempty"`
	OpState           string   `json:"op_state,omitempty"`
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

// CredentialResponse is the object returned from the Client.Callback method.
// It contains the issued credential (as base64url-encoded JSON) and the credential's format.
type CredentialResponse struct {
	Credential string `json:"credential,omitempty"` // Optional for deferred credential flow
	Format     string `json:"format,omitempty"`
}

type tokenResponse struct {
	AccessToken     string `json:"access_token,omitempty"`
	TokenType       string `json:"token_type,omitempty"`
	ExpiresIn       int    `json:"expires_in,omitempty"`
	RefreshToken    string `json:"refresh_token,omitempty"`
	CNonce          string `json:"c_nonce,omitempty"`
	CNonceExpiresIn int    `json:"c_nonce_expires_in,omitempty"`
}

type credentialRequest struct {
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`
	DID    string `json:"did,omitempty"`
	Proof  proof  `json:"proof,omitempty"`
}

type proof struct {
	ProofType       string `json:"proof_type,omitempty"`
	JWT             string `json:"jwt,omitempty"`
	CNonce          string `json:"c_nonce,omitempty"`
	CNonceExpiresIn int    `json:"c_nonce_expires_in,omitempty"`
}
