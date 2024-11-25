/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oauth2

import (
	"strings"

	"github.com/trustbloc/kms-go/doc/jose/jwk"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	goapioauth2 "github.com/trustbloc/wallet-sdk/pkg/oauth2"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// ClientMetadata represents a set of client metadata values.
type ClientMetadata struct {
	goAPIClientMetadata *goapioauth2.ClientMetadata
}

// NewClientMetadata creates a new ClientMetadata object.
func NewClientMetadata() *ClientMetadata {
	return &ClientMetadata{goAPIClientMetadata: &goapioauth2.ClientMetadata{}}
}

// RedirectURIs returns the redirect URIs.
func (c *ClientMetadata) RedirectURIs() *api.StringArray {
	return &api.StringArray{Strings: c.goAPIClientMetadata.RedirectURIs}
}

// SetRedirectURIs sets the redirect URIs.
func (c *ClientMetadata) SetRedirectURIs(redirectURIs *api.StringArray) {
	if redirectURIs == nil {
		redirectURIs = api.NewStringArray()
	}

	c.goAPIClientMetadata.RedirectURIs = redirectURIs.Strings
}

// TokenEndpointAuthMethod returns the token endpoint authentication method.
func (c *ClientMetadata) TokenEndpointAuthMethod() string {
	return c.goAPIClientMetadata.TokenEndpointAuthMethod
}

// SetTokenEndpointAuthMethod sets the token endpoint authentication method.
func (c *ClientMetadata) SetTokenEndpointAuthMethod(tokenEndpointAuthMethod string) {
	c.goAPIClientMetadata.TokenEndpointAuthMethod = tokenEndpointAuthMethod
}

// GrantTypes returns the grant types.
func (c *ClientMetadata) GrantTypes() *api.StringArray {
	return &api.StringArray{Strings: c.goAPIClientMetadata.GrantTypes}
}

// SetGrantTypes sets the grant types.
func (c *ClientMetadata) SetGrantTypes(grantTypes *api.StringArray) {
	if grantTypes == nil {
		grantTypes = api.NewStringArray()
	}

	c.goAPIClientMetadata.GrantTypes = grantTypes.Strings
}

// ResponseTypes returns the response types.
func (c *ClientMetadata) ResponseTypes() *api.StringArray {
	return &api.StringArray{Strings: c.goAPIClientMetadata.ResponseTypes}
}

// SetResponseTypes sets the response types.
func (c *ClientMetadata) SetResponseTypes(responseTypes *api.StringArray) {
	if responseTypes == nil {
		responseTypes = api.NewStringArray()
	}

	c.goAPIClientMetadata.ResponseTypes = responseTypes.Strings
}

// ClientName returns the client name.
func (c *ClientMetadata) ClientName() string {
	return c.goAPIClientMetadata.ClientName
}

// SetClientName sets the client name.
func (c *ClientMetadata) SetClientName(clientName string) {
	c.goAPIClientMetadata.ClientName = clientName
}

// ClientURI returns the client URI.
func (c *ClientMetadata) ClientURI() string {
	return c.goAPIClientMetadata.ClientURI
}

// SetClientURI sets the client URI.
func (c *ClientMetadata) SetClientURI(clientURI string) {
	c.goAPIClientMetadata.ClientURI = clientURI
}

// LogoURI returns the logo URI.
func (c *ClientMetadata) LogoURI() string {
	return c.goAPIClientMetadata.LogoURI
}

// SetLogoURI sets the logo URI.
func (c *ClientMetadata) SetLogoURI(logoURI string) {
	c.goAPIClientMetadata.LogoURI = logoURI
}

// Scopes returns the scopes.
func (c *ClientMetadata) Scopes() *api.StringArray {
	scopesStrings := strings.Split(c.goAPIClientMetadata.Scope, " ")

	if len(scopesStrings) == 1 && scopesStrings[0] == "" {
		return nil
	}

	scopes := &api.StringArray{Strings: scopesStrings}

	return scopes
}

// SetScopes sets the scope values.
func (c *ClientMetadata) SetScopes(scopes *api.StringArray) {
	if scopes == nil || scopes.Length() == 0 {
		c.goAPIClientMetadata.Scope = ""

		return
	}

	var sb strings.Builder

	numOfScopes := scopes.Length()

	indexOfLastScope := numOfScopes - 1

	for i := range indexOfLastScope {
		sb.WriteString(scopes.AtIndex(i))
		sb.WriteString(" ")
	}

	sb.WriteString(scopes.AtIndex(indexOfLastScope))

	c.goAPIClientMetadata.Scope = sb.String()
}

// Contacts returns the contacts.
func (c *ClientMetadata) Contacts() *api.StringArray {
	return &api.StringArray{Strings: c.goAPIClientMetadata.Contacts}
}

// SetContacts sets the contacts.
func (c *ClientMetadata) SetContacts(contacts *api.StringArray) {
	if contacts == nil {
		contacts = api.NewStringArray()
	}

	c.goAPIClientMetadata.Contacts = contacts.Strings
}

// TOSURI returns the TOS (terms of service) document URI.
func (c *ClientMetadata) TOSURI() string {
	return c.goAPIClientMetadata.TOSURI
}

// SetTOSURI sets the TOS (terms of service) document URI.
func (c *ClientMetadata) SetTOSURI(tosURI string) {
	c.goAPIClientMetadata.TOSURI = tosURI
}

// PolicyURI returns the privacy policy document URI.
func (c *ClientMetadata) PolicyURI() string {
	return c.goAPIClientMetadata.PolicyURI
}

// SetPolicyURI sets the privacy policy document URI.
func (c *ClientMetadata) SetPolicyURI(policyURI string) {
	c.goAPIClientMetadata.PolicyURI = policyURI
}

// JWKSetURI returns the JWK Set document URI.
func (c *ClientMetadata) JWKSetURI() string {
	return c.goAPIClientMetadata.JWKSetURI
}

// SetJWKSetURI sets the JWK Set document URI.
func (c *ClientMetadata) SetJWKSetURI(jwksURI string) {
	c.goAPIClientMetadata.JWKSetURI = jwksURI
}

// JWKSet returns the JWK Set document value. If the client metadata doesn't have one, then nil is returned instead.
func (c *ClientMetadata) JWKSet() *api.JSONWebKeySet {
	if c.goAPIClientMetadata.JWKSet == nil {
		return nil
	}

	jsonWebKeySet := api.JSONWebKeySet{JWKs: make([]api.JSONWebKey, len(c.goAPIClientMetadata.JWKSet.JWKs))}

	for i := range c.goAPIClientMetadata.JWKSet.JWKs {
		jsonWebKeySet.JWKs[i] = api.JSONWebKey{JWK: &c.goAPIClientMetadata.JWKSet.JWKs[i]}
	}

	return &jsonWebKeySet
}

// SetJWKSet sets the JWK Set document value.
func (c *ClientMetadata) SetJWKSet(jwks *api.JSONWebKeySet) {
	goAPIJSONWebKeySet := goapi.JSONWebKeySet{JWKs: make([]jwk.JWK, len(jwks.JWKs))}

	for i := range jwks.JWKs {
		goAPIJSONWebKeySet.JWKs[i] = *jwks.JWKs[i].JWK
	}

	c.goAPIClientMetadata.JWKSet = &goAPIJSONWebKeySet
}

// SoftwareID returns the software ID.
func (c *ClientMetadata) SoftwareID() string {
	return c.goAPIClientMetadata.SoftwareID
}

// SetSoftwareID sets the software ID.
func (c *ClientMetadata) SetSoftwareID(softwareID string) {
	c.goAPIClientMetadata.SoftwareID = softwareID
}

// SoftwareVersion returns the software version.
func (c *ClientMetadata) SoftwareVersion() string {
	return c.goAPIClientMetadata.SoftwareVersion
}

// SetSoftwareVersion sets the software version.
func (c *ClientMetadata) SetSoftwareVersion(softwareVersion string) {
	c.goAPIClientMetadata.SoftwareVersion = softwareVersion
}

// IssuerState returns the issuer state.
func (c *ClientMetadata) IssuerState() string {
	return c.goAPIClientMetadata.IssuerState
}

// SetIssuerState sets the issuer state.
func (c *ClientMetadata) SetIssuerState(issuerState string) {
	c.goAPIClientMetadata.IssuerState = issuerState
}
