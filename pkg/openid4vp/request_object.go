/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import "github.com/trustbloc/vc-go/presexch"

type clientIDScheme string

const (
	didScheme         clientIDScheme = "did"
	redirectURIScheme clientIDScheme = "redirect_uri"
)

type requestObject struct {
	JTI                    string                           `json:"jti"`
	IAT                    int64                            `json:"iat"`
	Issuer                 string                           `json:"iss"`
	ResponseType           string                           `json:"response_type"` //nolint: tagliatelle
	ResponseMode           string                           `json:"response_mode"` //nolint: tagliatelle
	ResponseURI            string                           `json:"response_uri"`
	Scope                  string                           `json:"scope"`
	Nonce                  string                           `json:"nonce"`
	ClientID               string                           `json:"client_id"` //nolint: tagliatelle
	ClientIDScheme         clientIDScheme                   `json:"client_id_scheme"`
	State                  string                           `json:"state"`
	Exp                    int64                            `json:"exp"`
	ClientMetadata         clientMetadata                   `json:"client_metadata"`
	PresentationDefinition *presexch.PresentationDefinition `json:"presentation_definition"`

	// Deprecated: Deprecated in OID4VP-ID2. Use response_uri instead.
	RedirectURI string `json:"redirect_uri"`
	// Deprecated: Deprecated in OID4VP-ID2. Use client_metadata instead.
	Registration requestObjectRegistration `json:"registration"`
	// Deprecated: Deprecated in OID4VP-ID2. Use top-level "presentation_definition" instead.
	Claims requestObjectClaims `json:"claims"`
}

type clientMetadata struct {
	ClientName                  string           `json:"client_name"`                    //nolint: tagliatelle
	ClientPurpose               string           `json:"client_purpose"`                 //nolint: tagliatelle
	ClientLogoURI               string           `json:"logo_uri"`                       //nolint: tagliatelle
	SubjectSyntaxTypesSupported []string         `json:"subject_syntax_types_supported"` //nolint: tagliatelle
	VPFormats                   *presexch.Format `json:"vp_formats"`                     //nolint: tagliatelle
}

type requestObjectRegistration struct {
	ClientName                  string           `json:"client_name"`
	SubjectSyntaxTypesSupported []string         `json:"subject_syntax_types_supported"`
	VPFormats                   *presexch.Format `json:"vp_formats"`
	ClientPurpose               string           `json:"client_purpose"`
	LogoURI                     string           `json:"logo_uri"`
}

type requestObjectClaims struct {
	VPToken struct {
		PresentationDefinition *presexch.PresentationDefinition `json:"presentation_definition"`
	} `json:"vp_token"`
}
