/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema

// ResolvedDisplayData represents display information for some issued credentials based on an issuer's metadata.
type ResolvedDisplayData struct {
	IssuerDisplay      *ResolvedCredentialIssuerDisplay `json:"issuer_display,omitempty"`
	CredentialDisplays []CredentialDisplay              `json:"credential_displays,omitempty"`
}

// ResolvedCredentialIssuerDisplay represents display information about the issuer of some credential(s).
type ResolvedCredentialIssuerDisplay struct {
	Name   string `json:"name,omitempty"`
	Locale string `json:"locale,omitempty"`
}

// CredentialDisplay represents display data for a credential.
// Display data for specific claims (e.g. first name, date of birth, etc.) are in Claims.
type CredentialDisplay struct {
	Overview *CredentialOverview `json:"overview,omitempty"`
	Claims   []ResolvedClaim     `json:"claims,omitempty"`
}

// CredentialOverview represents display data for a credential as a whole.
// Display data for specific claims (e.g. first name, date of birth, etc.) are in CredentialDisplay.Claims
// (in the parent object above).
type CredentialOverview struct {
	Name            string `json:"name,omitempty"`
	Locale          string `json:"locale,omitempty"`
	Logo            *Logo  `json:"logo,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
}

// ResolvedClaim represents display data for a specific claim.
type ResolvedClaim struct {
	Label  string `json:"label,omitempty"`
	Value  string `json:"value,omitempty"`
	Locale string `json:"locale,omitempty"`
}

// Logo represents display information for a logo.
type Logo struct {
	URL     string `json:"url,omitempty"`
	AltText string `json:"alternative_text,omitempty"`
}
