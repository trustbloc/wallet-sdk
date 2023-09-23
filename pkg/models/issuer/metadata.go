/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package issuer contains models for representing an issuer's metadata.
package issuer

import (
	"errors"
	"strconv"
)

// Metadata represents metadata about an issuer as obtained from their .well-known OpenID configuration.
type Metadata struct {
	CredentialIssuer        string                   `json:"credential_issuer,omitempty"`
	AuthorizationServer     string                   `json:"authorization_server,omitempty"`
	CredentialEndpoint      string                   `json:"credential_endpoint,omitempty"`
	CredentialsSupported    []SupportedCredential    `json:"credentials_supported,omitempty"`
	LocalizedIssuerDisplays []LocalizedIssuerDisplay `json:"display,omitempty"`
	TokenEndpoint           string                   `json:"token_endpoint,omitempty"`
	jwtKID                  *string
}

// GetJWTKID returns the jwtKID field. This is exposed via this method instead of with an exported field because
// the linter expects all exported fields to have JSON tags, but the jwtKID field is only intended for use internally
// within Wallet-SDK.
func (m *Metadata) GetJWTKID() *string {
	return m.jwtKID
}

// SetJWTKID sets the jwtKID field.
func (m *Metadata) SetJWTKID(jwtKID string) {
	m.jwtKID = &jwtKID
}

// SupportedCredential represents metadata about a credential type that a credential issuer can issue.
type SupportedCredential struct {
	Format                               string                       `json:"format,omitempty"`
	Types                                []string                     `json:"types,omitempty"`
	ID                                   string                       `json:"id,omitempty"`
	LocalizedCredentialDisplays          []LocalizedCredentialDisplay `json:"display,omitempty"`
	CredentialSubject                    map[string]*Claim            `json:"credentialSubject,omitempty"`
	CryptographicBindingMethodsSupported []string                     `json:"cryptographic_binding_methods_supported,omitempty"` //nolint:lll // Formatter forces these line symbols to line up, and this is a long name
	CryptographicSuitesSupported         []string                     `json:"cryptographic_suites_supported,omitempty"`
}

// LocalizedCredentialDisplay represents display data for a credential as a whole for a certain locale.
// Display data for specific claims (e.g. first name, date of birth, etc.) are in SupportedCredential.CredentialSubject
// (in the parent object above).
type LocalizedCredentialDisplay struct {
	Name            string `json:"name,omitempty"`
	Locale          string `json:"locale,omitempty"`
	Logo            *Logo  `json:"logo,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
}

// Claim represents display data for a specific claim in (potentially) multiple locales.
// Each ClaimDisplay represents display data for a single locale.
type Claim struct {
	LocalizedClaimDisplays []LocalizedClaimDisplay `json:"display,omitempty"`
	ValueType              string                  `json:"value_type,omitempty"`
	Order                  interface{}             `json:"order,omitempty"`
	Pattern                string                  `json:"pattern,omitempty"`
	Mask                   string                  `json:"mask,omitempty"`
}

// OrderAsInt returns this Claim's Order value as an integer.
func (c *Claim) OrderAsInt() (int, error) {
	// If this issuer metadata was sent by the server as JSON, then it should unmarshal into a float64.
	orderAsFloat64, ok := c.Order.(float64)
	if ok {
		return int(orderAsFloat64), nil
	}

	// If it was sent as a JWT, then it'll be a string.
	orderAsString, ok := c.Order.(string)
	if ok {
		orderAsInt, err := strconv.Atoi(orderAsString)
		if err != nil {
			return -1, err
		}

		return orderAsInt, nil
	}

	// Other types aren't expected currently, so return an error.
	// If we do expect another type in the future, then we'll need to add support for it to this method.

	return -1, errors.New("order is nil or an unsupported type")
}

// Logo represents display information for a logo.
type Logo struct {
	URL     string `json:"url,omitempty"`
	AltText string `json:"alt_text,omitempty"`
}

// LocalizedIssuerDisplay represents display information for an issuer in a specific locale.
type LocalizedIssuerDisplay struct {
	Name            string `json:"name,omitempty"`
	Locale          string `json:"locale,omitempty"`
	URL             string `json:"url,omitempty"`
	Logo            *Logo  `json:"logo,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
}

// LocalizedClaimDisplay represents display information for a claim in a specific locale.
type LocalizedClaimDisplay struct {
	Name   string `json:"name,omitempty"`
	Locale string `json:"locale,omitempty"`
}
