/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
)

type idTokenVPToken struct {
	PresentationSubmission interface{} `json:"presentation_submission"` //nolint: tagliatelle
}

type idTokenClaims struct {
	VPToken idTokenVPToken `json:"_vp_token"` //nolint: tagliatelle
	Nonce   string         `json:"nonce"`
	Exp     int64          `json:"exp"`
	Iss     string         `json:"iss"`
	Sub     string         `json:"sub"`
}

type vpTokenClaims struct {
	VP    *verifiable.Presentation `json:"vp"`
	Nonce string                   `json:"nonce"`
	Exp   int64                    `json:"exp"`
	Iss   string                   `json:"iss"`
}
