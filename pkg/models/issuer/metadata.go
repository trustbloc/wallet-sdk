/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package issuer contains models for representing an issuer's metadata.
package issuer

// Metadata represents metadata about an issuer as obtained from their .well-known OpenID configuration.
type Metadata struct {
	CredentialIssuer     string                `json:"credential_issuer,omitempty"`
	AuthorizationServer  string                `json:"authorization_server,omitempty"`
	CredentialEndpoint   string                `json:"credential_endpoint,omitempty"`
	CredentialsSupported []SupportedCredential `json:"credentials_supported,omitempty"`
	// IssuerDisplays represents display information for the issuer's name in various locales.
	IssuerDisplays []Display `json:"display,omitempty"`
}

// SupportedCredential represents metadata about a credential type that a credential issuer can issue.
type SupportedCredential struct {
	Format                               string               `json:"format,omitempty"`
	Types                                []string             `json:"types,omitempty"`
	ID                                   string               `json:"id,omitempty"`
	Overview                             []CredentialOverview `json:"display,omitempty"`
	CredentialSubject                    map[string]Claim     `json:"credentialSubject,omitempty"`
	CryptographicBindingMethodsSupported []string             `json:"cryptographic_binding_methods_supported,omitempty"`
	CryptographicSuitesSupported         []string             `json:"cryptographic_suites_supported,omitempty"`
}

// CredentialOverview represents display data for a credential as a whole.
// Display data for specific claims (e.g. first name, date of birth, etc.) are in SupportedCredential.CredentialSubject
// (in the parent object above).
type CredentialOverview struct {
	Name            string `json:"name,omitempty"`
	Locale          string `json:"locale,omitempty"`
	Logo            *Logo  `json:"logo,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
}

// Claim represents display data for a specific claim in (potentially) multiple locales.
// Each ClaimDisplay represents display data for a single locale.
type Claim struct {
	// Displays represents display data for a specific claim in various locales.
	Displays  []Display `json:"display,omitempty"`
	ValueType string    `json:"value_type,omitempty"`
}

// Logo represents display information for a logo.
type Logo struct {
	URL     string `json:"url,omitempty"`
	AltText string `json:"alt_text,omitempty"`
}

// Display represents display information for some piece of data in a specific locale.
type Display struct {
	Name   string `json:"name,omitempty"`
	Locale string `json:"locale,omitempty"`
}
