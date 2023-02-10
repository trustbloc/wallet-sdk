/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package issuer contains models for representing an issuer's metadata.
package issuer

// Metadata represents metadata about an issuer as obtained from their .well-known OpenID configuration.
type Metadata struct {
	Issuer                    string                         `json:"issuer,omitempty"`
	AuthorizationEndpoint     string                         `json:"authorization_endpoint,omitempty"`
	TokenEndpoint             string                         `json:"token_endpoint,omitempty"`
	PushedAuthRequestEndpoint string                         `json:"pushed_authorization_request_endpoint,omitempty"`
	RequirePushedAuthRequests bool                           `json:"require_pushed_authorization_requests,omitempty"`
	CredentialEndpoint        string                         `json:"credential_endpoint,omitempty"`
	CredentialsSupported      map[string]SupportedCredential `json:"credentials_supported,omitempty"`
	CredentialIssuer          *CredentialIssuer              `json:"credential_issuer,omitempty"`
}

// SupportedCredential represents metadata about a credential type that a credential issuer can issue.
type SupportedCredential struct {
	Formats                              map[string]Format    `json:"formats,omitempty"`
	Overview                             []CredentialOverview `json:"display,omitempty"`
	Claims                               map[string]Claim     `json:"claims,omitempty"`
	CryptographicBindingMethodsSupported []string             `json:"cryptographic_binding_methods_supported,omitempty"`
	CryptographicSuitesSupported         []string             `json:"cryptographic_suites_supported,omitempty"`
}

// CredentialOverview represents display data for a credential as a whole.
// Display data for specific claims (e.g. first name, date of birth, etc.) are in SupportedCredential.Claims
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
	Displays  []ClaimDisplay `json:"display,omitempty"`
	ValueType string         `json:"value_type,omitempty"`
}

// Logo represents display information for a logo.
type Logo struct {
	URL     string `json:"url,omitempty"`
	AltText string `json:"alternative_text,omitempty"`
}

// Format represents a single supported VC format within a SupportedCredential object.
type Format struct {
	CryptographicBindingMethodsSupported []string `json:"cryptographic_binding_methods_supported,omitempty"`
	CryptographicSuitesSupported         []string `json:"cryptographic_suites_supported,omitempty"`
	Types                                []string `json:"types,omitempty"`
}

// CredentialIssuer represents display information about the issuer of some credential(s) for (potentially) multiple
// locales.
// Each Display represents display data for a single locale.
type CredentialIssuer struct {
	Displays []CredentialIssuerDisplay `json:"display,omitempty"`
}

// ClaimDisplay represents display data for a specific claim for a single locale.
type ClaimDisplay struct {
	Name   string `json:"name,omitempty"`
	Locale string `json:"locale,omitempty"`
}

// CredentialIssuerDisplay represents display information for the issuer of some credential(s) for a single locale.
type CredentialIssuerDisplay struct {
	Name   string `json:"name,omitempty"`
	Locale string `json:"locale,omitempty"`
}
