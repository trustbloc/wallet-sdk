/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oauth2

import (
	"errors"
	"strings"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapioauth2 "github.com/trustbloc/wallet-sdk/pkg/oauth2"
)

// RegisterClientResponse represents a response to a new client registration request.
type RegisterClientResponse struct {
	goAPIRegisterClientResponse *goapioauth2.RegisterClientResponse
}

// ClientID returns the client ID.
func (r *RegisterClientResponse) ClientID() string {
	return r.goAPIRegisterClientResponse.ClientID
}

// ClientSecret returns the client secret.
func (r *RegisterClientResponse) ClientSecret() string {
	return r.goAPIRegisterClientResponse.ClientSecret
}

// HasClientIDIssuedAt indicates whether this RegisterClientResponse specifies when the client ID was issued.
func (r *RegisterClientResponse) HasClientIDIssuedAt() bool {
	return r.goAPIRegisterClientResponse.ClientIDIssuedAt != nil
}

// ClientIDIssuedAt returns the time at which the client ID was issued.
// The HasClientIDIssuedAt method should be called first to determine whether this RegisterClientResponse
// specifies when the client ID was issued before calling this method.
// This method returns an error if (and only if) HasClientIDIssuedAt returns false.
func (r *RegisterClientResponse) ClientIDIssuedAt() (int, error) {
	if !r.HasClientIDIssuedAt() {
		return -1, errors.New("the register client response object does not specify when the client ID was issued")
	}

	return *r.goAPIRegisterClientResponse.ClientIDIssuedAt, nil
}

// HasClientSecretExpiresAt indicates whether this RegisterClientResponse specifies when the client secret expires.
func (r *RegisterClientResponse) HasClientSecretExpiresAt() bool {
	return r.goAPIRegisterClientResponse.ClientSecretExpiresAt != nil
}

// ClientSecretExpiresAt returns the time at which the client secret will expire or 0 if it will not expire.
// The HasClientSecretExpiresAt method should be called first to determine whether this RegisterClientResponse
// specifies when the client secret will expire before calling this method.
// This method returns an error if (and only if) HasClientSecretExpiresAt returns false.
func (r *RegisterClientResponse) ClientSecretExpiresAt() (int, error) {
	if !r.HasClientSecretExpiresAt() {
		return -1, errors.New("the register client response object does not " +
			"specify when the client secret expires")
	}

	return *r.goAPIRegisterClientResponse.ClientSecretExpiresAt, nil
}

// RegisteredMetadata returns the RegisteredMetadata object, which can be used to determine what metadata was
// actually registered by the authorization server (which may differ from the client metadata in the request).
func (r *RegisterClientResponse) RegisteredMetadata() *RegisteredMetadata {
	return &RegisteredMetadata{goAPIRegisteredMetadata: r.goAPIRegisterClientResponse.RegisteredMetadata}
}

// RegisteredMetadata represents a set of registered metadata values.
type RegisteredMetadata struct {
	goAPIRegisteredMetadata goapioauth2.RegisteredMetadata
}

// RedirectURIs returns the redirect URIs.
func (c *RegisteredMetadata) RedirectURIs() *api.StringArray {
	return &api.StringArray{Strings: c.goAPIRegisteredMetadata.RedirectURIs}
}

// TokenEndpointAuthMethod returns the token endpoint authentication method.
func (c *RegisteredMetadata) TokenEndpointAuthMethod() string {
	return c.goAPIRegisteredMetadata.TokenEndpointAuthMethod
}

// GrantTypes returns the grant types.
func (c *RegisteredMetadata) GrantTypes() *api.StringArray {
	return &api.StringArray{Strings: c.goAPIRegisteredMetadata.GrantTypes}
}

// ResponseTypes returns the response types.
func (c *RegisteredMetadata) ResponseTypes() *api.StringArray {
	return &api.StringArray{Strings: c.goAPIRegisteredMetadata.ResponseTypes}
}

// ClientName returns the client name.
func (c *RegisteredMetadata) ClientName() string {
	return c.goAPIRegisteredMetadata.ClientName
}

// ClientURI returns the client URI.
func (c *RegisteredMetadata) ClientURI() string {
	return c.goAPIRegisteredMetadata.ClientURI
}

// LogoURI returns the logo URI.
func (c *RegisteredMetadata) LogoURI() string {
	return c.goAPIRegisteredMetadata.LogoURI
}

// Scopes returns the scopes.
func (c *RegisteredMetadata) Scopes() *api.StringArray {
	scopesStrings := strings.Split(c.goAPIRegisteredMetadata.Scope, " ")

	if len(scopesStrings) == 1 && scopesStrings[0] == "" {
		return nil
	}

	scopes := &api.StringArray{Strings: scopesStrings}

	return scopes
}

// Contacts returns the contacts.
func (c *RegisteredMetadata) Contacts() *api.StringArray {
	return &api.StringArray{Strings: c.goAPIRegisteredMetadata.Contacts}
}

// TOSURI returns the TOS (terms of service) document URI.
func (c *RegisteredMetadata) TOSURI() string {
	return c.goAPIRegisteredMetadata.TOSURI
}

// PolicyURI returns the privacy policy document URI.
func (c *RegisteredMetadata) PolicyURI() string {
	return c.goAPIRegisteredMetadata.PolicyURI
}

// JWKSetURI returns the JWK Set document URI.
func (c *RegisteredMetadata) JWKSetURI() string {
	return c.goAPIRegisteredMetadata.JWKSetURI
}

// JWKSet returns the JWK Set document value. If the client metadata doesn't have one, then nil is returned instead.
func (c *RegisteredMetadata) JWKSet() *api.JSONWebKeySet {
	if c.goAPIRegisteredMetadata.JWKSet == nil {
		return nil
	}

	jsonWebKeySet := api.JSONWebKeySet{JWKs: make([]api.JSONWebKey, len(c.goAPIRegisteredMetadata.JWKSet.JWKs))}

	for i := range c.goAPIRegisteredMetadata.JWKSet.JWKs {
		jsonWebKeySet.JWKs[i] = api.JSONWebKey{JWK: &c.goAPIRegisteredMetadata.JWKSet.JWKs[i]}
	}

	return &jsonWebKeySet
}

// SoftwareID returns the software ID.
func (c *RegisteredMetadata) SoftwareID() string {
	return c.goAPIRegisteredMetadata.SoftwareID
}

// SoftwareVersion returns the software version.
func (c *RegisteredMetadata) SoftwareVersion() string {
	return c.goAPIRegisteredMetadata.SoftwareVersion
}
