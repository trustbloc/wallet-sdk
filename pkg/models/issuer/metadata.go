/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package issuer contains models for representing an issuer's metadata.
package issuer

// Metadata represents metadata about an issuer as obtained from their .well-known OpenID configuration.
type Metadata struct {
	CredentialIssuer        string                   `json:"credential_issuer,omitempty"`
	AuthorizationServer     string                   `json:"authorization_server,omitempty"`
	CredentialEndpoint      string                   `json:"credential_endpoint,omitempty"`
	CredentialsSupported    []SupportedCredential    `json:"credentials_supported,omitempty"`
	LocalizedIssuerDisplays []LocalizedIssuerDisplay `json:"display,omitempty"`
}

// SupportedCredential represents metadata about a credential type that a credential issuer can issue.
type SupportedCredential struct {
	Format                               string                       `json:"format,omitempty"`
	Types                                []string                     `json:"types,omitempty"`
	ID                                   string                       `json:"id,omitempty"`
	LocalizedCredentialDisplays          []LocalizedCredentialDisplay `json:"display,omitempty"`
	CredentialSubject                    map[string]Claim             `json:"credentialSubject,omitempty"`
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
	Order                  *int                    `json:"order,omitempty"`
	Pattern                string                  `json:"pattern,omitempty"`
	Mask                   string                  `json:"mask,omitempty"`
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
