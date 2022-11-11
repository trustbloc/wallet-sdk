/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import "github.com/hyperledger/aries-framework-go/pkg/doc/presexch"

type requestObject struct {
	JTI          string                    `json:"jti"`
	IAT          int64                     `json:"iat"`
	ResponseType string                    `json:"response_type"` //nolint: tagliatelle
	ResponseMode string                    `json:"response_mode"` //nolint: tagliatelle
	Scope        string                    `json:"scope"`
	Nonce        string                    `json:"nonce"`
	ClientID     string                    `json:"client_id"`    //nolint: tagliatelle
	RedirectURI  string                    `json:"redirect_uri"` //nolint: tagliatelle
	State        string                    `json:"state"`
	Exp          int64                     `json:"exp"`
	Registration requestObjectRegistration `json:"registration"`
	Claims       requestObjectClaims       `json:"claims"`
}

type requestObjectRegistration struct {
	ClientName                  string           `json:"client_name"`                    //nolint: tagliatelle
	SubjectSyntaxTypesSupported []string         `json:"subject_syntax_types_supported"` //nolint: tagliatelle
	VPFormats                   *presexch.Format `json:"vp_formats"`                     //nolint: tagliatelle
	ClientPurpose               string           `json:"client_purpose"`                 //nolint: tagliatelle
}

type requestObjectClaims struct {
	VPToken vpToken `json:"vp_token"` //nolint: tagliatelle
}
type vpToken struct {
	PresentationDefinition *presexch.PresentationDefinition `json:"presentation_definition"` //nolint: tagliatelle
}
