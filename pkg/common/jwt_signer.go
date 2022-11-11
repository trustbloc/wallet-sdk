/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// JWSSigner utility class used to sign jwt using api.Crypto.
type JWSSigner struct {
	keyID     string
	algorithm string
	crypto    api.Crypto
}

// NewJWSSigner creates jwt signer.
func NewJWSSigner(keyID, algorithm string, crypto api.Crypto) *JWSSigner {
	return &JWSSigner{
		keyID:     keyID,
		algorithm: algorithm,
		crypto:    crypto,
	}
}

// GetKeyID return id of key used for signing.
func (s *JWSSigner) GetKeyID() string {
	return s.keyID
}

// Sign signs jwt token.
func (s *JWSSigner) Sign(data []byte) ([]byte, error) {
	return s.crypto.Sign(data, s.keyID)
}

// Headers provides JWS headers.
func (s *JWSSigner) Headers() jose.Headers {
	return jose.Headers{
		jose.HeaderKeyID:     s.keyID,
		jose.HeaderAlgorithm: s.algorithm,
	}
}
