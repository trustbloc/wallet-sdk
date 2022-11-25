/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema

import (
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

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
	Overview *issuer.CredentialDisplay `json:"overview,omitempty"`
	Claims   []ResolvedClaim           `json:"claims,omitempty"`
}

// ResolvedClaim represents display data for a specific claim.
type ResolvedClaim struct {
	Label  string `json:"label,omitempty"`
	Value  string `json:"value,omitempty"`
	Locale string `json:"locale,omitempty"`
}
