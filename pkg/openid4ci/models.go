/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "encoding/json"

// CredentialOffer represents the Credential Offer object as defined in
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html#section-4.1.1.
type CredentialOffer struct {
	CredentialIssuer string                            `json:"credential_issuer,omitempty"`
	Credentials      []Credentials                     `json:"credentials,omitempty"`
	Grants           map[string]map[string]interface{} `json:"grants,omitempty"`
}

// Credentials represents the credential format and types in a Credential Offer.
type Credentials struct {
	Format string   `json:"format,omitempty"`
	Types  []string `json:"types,omitempty"`
}

// AuthorizeResult is the object returned from the Client.Authorize method.
// An empty/missing AuthorizationRedirectEndpoint indicates that the wallet is pre-authorized.
type AuthorizeResult struct {
	AuthorizationRedirectEndpoint string
	UserPINRequired               bool
}

type authorizationDetails struct {
	Type      string   `json:"type,omitempty"`
	Locations []string `json:"locations,omitempty"`
	Types     []string `json:"types,omitempty"`
	Format    string   `json:"format,omitempty"`
}

// OpenIDConfig represents an issuer's OpenID configuration.
type OpenIDConfig struct {
	AuthorizationEndpoint  string   `json:"authorization_endpoint,omitempty"`
	ResponseTypesSupported []string `json:"response_types_supported,omitempty"`
	TokenEndpoint          string   `json:"token_endpoint,omitempty"`
	RegistrationEndpoint   *string  `json:"registration_endpoint,omitempty"`
}

// CredentialResponse is the object returned from the Client.Callback method.
// It contains the issued credential and the credential's format.
type CredentialResponse struct {
	Credential interface{} `json:"credential,omitempty"` // Optional for deferred credential flow.
	Format     string      `json:"format,omitempty"`
	AscID      string      `json:"ack_id"`
}

// SerializeToCredentialsBytes serializes underlying credential to proper bytes representation depending on
// credential format.
func (r *CredentialResponse) SerializeToCredentialsBytes() ([]byte, error) {
	// TODO: https://github.com/trustbloc/wallet-sdk/issues/456 check response.Format after
	// VCS starts return valid value.
	switch cred := r.Credential.(type) {
	case string:
		return []byte(cred), nil
	default:
		return json.Marshal(cred)
	}
}

type preAuthTokenResponse struct {
	AccessToken     string `json:"access_token,omitempty"`
	TokenType       string `json:"token_type,omitempty"`
	ExpiresIn       int    `json:"expires_in,omitempty"`
	RefreshToken    string `json:"refresh_token,omitempty"`
	CNonce          string `json:"c_nonce,omitempty"`
	CNonceExpiresIn int    `json:"c_nonce_expires_in,omitempty"`
}

type credentialRequest struct {
	Types  []string `json:"types,omitempty"`
	Format string   `json:"format,omitempty"`
	Proof  proof    `json:"proof,omitempty"`
}

type proof struct {
	ProofType       string `json:"proof_type,omitempty"`
	JWT             string `json:"jwt,omitempty"`
	CNonce          string `json:"c_nonce,omitempty"`
	CNonceExpiresIn int    `json:"c_nonce_expires_in,omitempty"`
}

type errorResponse struct {
	Error string `json:"error,omitempty"`
}

type acknowledgementRequest struct {
	Credentials []credentialAcknowledgement `json:"credentials"`
}

type credentialAcknowledgement struct {
	AckID            string `json:"ack_id"`
	Status           string `json:"status"`
	IssuerIdentifier string `json:"issuer_identifier"`
}
