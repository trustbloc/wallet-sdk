/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema

// ResolvedDisplayData represents display information for some issued credentials based on an issuer's metadata.
type ResolvedDisplayData struct {
	IssuerDisplay      *ResolvedIssuerDisplay `json:"issuer_display,omitempty"`
	CredentialDisplays []CredentialDisplay    `json:"credential_displays,omitempty"`
}

// ResolvedIssuerDisplay represents display information about the issuer of some credential(s).
type ResolvedIssuerDisplay struct {
	Name            string `json:"name,omitempty"`
	Locale          string `json:"locale,omitempty"`
	URL             string `json:"url,omitempty"`
	Logo            *Logo  `json:"logo,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
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
	// RawID is the raw field name (key) from the VC associated with this claim.
	// It's not localized or formatted for display.
	RawID      string      `json:"raw_id,omitempty"`
	Label      string      `json:"label,omitempty"`
	ValueType  string      `json:"value_type,omitempty"`
	RawValue   string      `json:"raw_value,omitempty"`
	Value      *string     `json:"value,omitempty"`
	Order      *int        `json:"order,omitempty"`
	Pattern    string      `json:"pattern,omitempty"`
	Mask       string      `json:"mask,omitempty"`
	Locale     string      `json:"locale,omitempty"`
	Attachment *Attachment `json:"attachment,omitempty"`
}

// ResolvedData represents display information for the credentials based on an issuer's metadata.
type ResolvedData struct {
	LocalizedIssuer []ResolvedIssuerDisplay `json:"localized_issuer,omitempty"`
	Credential      []Credential            `json:"credentials,omitempty"`
}

// Credential represents display data for a credential.
// Display data for specific claims (e.g. first name, date of birth, etc.) are in Subject.
type Credential struct {
	LocalizedOverview []CredentialOverview `json:"localized_overview,omitempty"`
	Subject           []Subject            `json:"subjects,omitempty"`
}

// Subject represents display data for a specific credential subject.
type Subject struct {
	RawID           string      `json:"raw_id,omitempty"`
	LocalizedLabels []Label     `json:"localized_labels,omitempty"`
	ValueType       string      `json:"value_type,omitempty"`
	RawValue        string      `json:"raw_value,omitempty"`
	Value           *string     `json:"value,omitempty"`
	Order           *int        `json:"order,omitempty"`
	Pattern         string      `json:"pattern,omitempty"`
	Mask            string      `json:"mask,omitempty"`
	Attachment      *Attachment `json:"attachment,omitempty"`
}

// Label represents display information for localizaed credential subject label.
type Label struct {
	Name   string `json:"name,omitempty"`
	Locale string `json:"locale,omitempty"`
}

// Logo represents display information for a logo.
type Logo struct {
	URL     string `json:"uri,omitempty"`
	AltText string `json:"alt_text,omitempty"`
}

// Attachment contains data for display for a vc attachment.
type Attachment struct {
	ID          string   `json:"id,omitempty"`
	Type        []string `json:"type,omitempty"`
	MimeType    string   `json:"mimeType,omitempty"`
	Description string   `json:"description,omitempty"`
	URI         string   `json:"uri,omitempty"`
	Hash        string   `json:"hash,omitempty"`
	HashAlg     string   `json:"hash-alg,omitempty"`
}
