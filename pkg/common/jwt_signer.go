/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	cryptolib "crypto"
	"encoding/base64"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/doc/util/jwkkid"
	"github.com/hyperledger/aries-framework-go/pkg/kms"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/models"
)

const (
	// Ed25519VerificationKey2018 is a supported DID verification type.
	Ed25519VerificationKey2018 = "Ed25519VerificationKey2018"
	// JSONWebKey2020 is a supported DID verification type.
	JSONWebKey2020 = "JsonWebKey2020"
	// EdDSA is  signature algorithm for Ed25519VerificationKey2018.
	EdDSA = "EdDSA"
)

// JWSSigner utility class used to sign jwt using api.Crypto.
type JWSSigner struct {
	keyID     string
	algorithm string
	cryptoKID string
	crypto    api.Crypto
}

// NewJWSSigner creates jwt signer.
func NewJWSSigner(vm *models.VerificationMethod, crypto api.Crypto) (*JWSSigner, error) {
	algorithm, cryptoKID, err := algAndThumbprint(vm)
	if err != nil {
		return nil, err
	}

	return &JWSSigner{
		keyID:     vm.ID,
		algorithm: algorithm,
		crypto:    crypto,
		cryptoKID: cryptoKID,
	}, nil
}

// returns: alg, thumbprint, error
func algAndThumbprint(vm *models.VerificationMethod) (string, string, error) {
	if vm.Type == Ed25519VerificationKey2018 {
		tp, err := jwkkid.CreateKID(vm.Key.Raw, kms.ED25519Type)
		if err != nil {
			return "", "", fmt.Errorf("generating thumbprint for ed25519 key: %w", err)
		}

		return EdDSA, tp, nil
	}

	if vm.Type != JSONWebKey2020 {
		return "", "", fmt.Errorf("verification method type '%s' not supported", vm.Type)
	}

	if vm.Key.JSONWebKey == nil {
		return "", "", fmt.Errorf("missing jwk for %s verification method", JSONWebKey2020)
	}

	tpBytes, err := vm.Key.JSONWebKey.Thumbprint(cryptolib.SHA256)
	if err != nil {
		return "", "", fmt.Errorf("creating crypto thumbprint for JWK: %w", err)
	}

	tp := base64.RawURLEncoding.EncodeToString(tpBytes)

	// TODO: https://github.com/trustbloc/wallet-sdk/issues/161 handle more key types
	if vm.Key.JSONWebKey.Crv == "Ed25519" {
		return EdDSA, tp, nil
	}

	return "", "", fmt.Errorf("jwt signer: currently only Ed25519 is supported")
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
