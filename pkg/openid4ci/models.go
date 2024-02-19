/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"encoding/json"
	"time"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

// CredentialOffer represents the Credential Offer object as defined in
// https://openid.github.io/OpenID4VCI/openid-4-verifiable-credential-issuance-wg-draft.html#section-4.1.1.
type CredentialOffer struct {
	CredentialIssuer           string                            `json:"credential_issuer,omitempty"`
	CredentialConfigurationIDs []string                          `json:"credential_configuration_ids"`
	Grants                     map[string]map[string]interface{} `json:"grants,omitempty"`
}

// AuthorizeResult is the object returned from the Client.Authorize method.
// An empty/missing AuthorizationRedirectEndpoint indicates that the wallet is pre-authorized.
type AuthorizeResult struct {
	AuthorizationRedirectEndpoint string
	UserPINRequired               bool
}

// authorizationDetails is a model to convey the details about the Credentials the Client wants to obtain.
type authorizationDetails struct {
	// REQUIRED when Format parameter is not present.
	// String specifying a unique identifier of the Credential being described in the
	// credential_configurations_supported map in the Credential Issuer Metadata.
	// The referenced object in the credential_configurations_supported map conveys the details,
	// such as the format, for issuance of the requested Credential.
	// It MUST NOT be present if format parameter is present.
	CredentialConfigurationID string `json:"credential_configuration_id,omitempty"`

	// Object containing the detailed description of the credential type.
	CredentialDefinition *issuer.CredentialDefinition `json:"credential_definition,omitempty"`

	// REQUIRED when CredentialConfigurationID parameter is not present.
	// String identifying the format of the Credential the Wallet needs.
	// This Credential format identifier determines further claims in the authorization details object needed
	// to identify the Credential type in the requested format.
	// It MUST NOT be present if credential_configuration_id parameter is present.
	Format string `json:"format,omitempty"`

	// An array of strings that allows a client to specify the location of the resource server(s)
	// allowing the Authorization Server to mint audience restricted access tokens.
	Locations []string `json:"locations,omitempty"`

	// String that determines the authorization details type. MUST be set to "openid_credential" for OIDC4VC.
	Type string `json:"type"`
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

type universalAuthToken struct {
	AccessToken  string    `json:"access_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
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
