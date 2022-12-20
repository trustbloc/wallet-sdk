/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

const (
	// Ed25519VerificationKey2018 is a supported DID verification type.
	Ed25519VerificationKey2018 = "Ed25519VerificationKey2018"

	// EdDSA is  signature algorithm for Ed25519VerificationKey2018.
	EdDSA = "EdDSA"
)

// JWSSigner utility class used to sign jwt using api.Crypto.
type JWSSigner struct {
	keyID     string
	algorithm string
	crypto    api.Crypto
}

// NewJWSSigner creates jwt signer.
func NewJWSSigner(keyID, verificationType string, crypto api.Crypto) (*JWSSigner, error) {
	algorithm, err := getSignAlgorithmForVerificationType(verificationType)
	if err != nil {
		return nil, err
	}

	return &JWSSigner{
		keyID:     keyID,
		algorithm: algorithm,
		crypto:    crypto,
	}, nil
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

func getSignAlgorithmForVerificationType(verificationType string) (string, error) {
	if verificationType == Ed25519VerificationKey2018 {
		return EdDSA, nil
	}

	return "", fmt.Errorf("jwt signer: currently only %s is supported", Ed25519VerificationKey2018)
}
