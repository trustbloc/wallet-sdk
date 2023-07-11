/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oauth2

import (
	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// ClientMetadata represents a set of client metadata values.
type ClientMetadata struct {
	RedirectURIs            []string           `json:"redirect_uris,omitempty"`
	TokenEndpointAuthMethod string             `json:"token_endpoint_auth_method,omitempty"`
	GrantTypes              []string           `json:"grant_types,omitempty"`
	ResponseTypes           []string           `json:"response_types,omitempty"`
	ClientName              string             `json:"client_name,omitempty"`
	ClientURI               string             `json:"client_uri,omitempty"`
	LogoURI                 string             `json:"logo_uri,omitempty"`
	Scope                   string             `json:"scope,omitempty"` // Space-separated strings
	Contacts                []string           `json:"contacts,omitempty"`
	TOSURI                  string             `json:"tos_uri,omitempty"`
	PolicyURI               string             `json:"policy_uri,omitempty"`
	JWKSetURI               string             `json:"jwks_uri,omitempty"`
	JWKSet                  *api.JSONWebKeySet `json:"jwks,omitempty"`
	SoftwareID              string             `json:"software_id,omitempty"`
	SoftwareVersion         string             `json:"software_version,omitempty"`
	// TODO: This is a temporary workaround for VCS. To be removed.
	IssuerState string `json:"issuer_state,omitempty"`
}

// RegisteredMetadata specifies what metadata was actually registered by the authorization server (which may differ
// from the client metadata in the request).
type RegisteredMetadata ClientMetadata

// RegisterClientResponse represents a response to a new client registration request.
type RegisterClientResponse struct {
	ClientID              string `json:"client_id"`
	ClientSecret          string `json:"client_secret,omitempty"`
	ClientIDIssuedAt      *int   `json:"client_id_issued_at,omitempty"`
	ClientSecretExpiresAt *int   `json:"client_secret_expires_at,omitempty"`
	RegisteredMetadata
}
