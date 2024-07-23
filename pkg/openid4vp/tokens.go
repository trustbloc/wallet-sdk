/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"github.com/trustbloc/vc-go/verifiable"
)

// CustomClaims represents custom claims to be added to id_token.
type CustomClaims struct {
	ScopeClaims map[string]interface{}
}

type idTokenClaims struct {
	Scope         map[string]interface{} `json:"_scope,omitempty"` //nolint: tagliatelle
	AttestationVP string                 `json:"_attestation_vp"`
	Nonce         string                 `json:"nonce"`
	Exp           int64                  `json:"exp"`
	Iss           string                 `json:"iss"`
	Sub           string                 `json:"sub"`
	Aud           string                 `json:"aud"`
	Nbf           int64                  `json:"nbf"`
	Iat           int64                  `json:"iat"`
	Jti           string                 `json:"jti"`

	// Deprecated: Use top-level "presentation_submission" instead.
	VPToken idTokenVPToken `json:"_vp_token"`
}

type idTokenVPToken struct {
	PresentationSubmission interface{} `json:"presentation_submission"` //nolint: tagliatelle
}

type vpTokenClaims struct {
	VP    *verifiable.Presentation `json:"vp"`
	Nonce string                   `json:"nonce"`
	Exp   int64                    `json:"exp"`
	Iss   string                   `json:"iss"`
	Aud   string                   `json:"aud"`
	Nbf   int64                    `json:"nbf"`
	Iat   int64                    `json:"iat"`
	Jti   string                   `json:"jti"`
}
