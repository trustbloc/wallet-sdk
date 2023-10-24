/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	cryptolib "crypto"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/trustbloc/kms-go/doc/jose"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/doc/util/jwkkid"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/jwt"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const (
	// Ed25519VerificationKey2018 is a supported DID verification type.
	Ed25519VerificationKey2018 = "Ed25519VerificationKey2018"
	// JSONWebKey2020 is a supported DID verification type.
	JSONWebKey2020 = "JsonWebKey2020"
	// EdDSA is  signature algorithm for Ed25519VerificationKey2018.
	EdDSA = "EdDSA"
	// ES384 is  signature algorithm for P-384 JSONWebKey2020.
	ES384 = "ES384"
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
		return nil, walleterror.NewValidationError(
			module,
			UnsupportedAlgorithmCode,
			UnsupportedAlgorithmError,
			err,
		)
	}

	if crypto == nil {
		return nil, walleterror.NewValidationError(
			module,
			NoCryptoProvidedCode,
			NoCryptoProvidedError,
			errors.New("crypto instance should be provided"),
		)
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

	alg, err := inferAlg(vm.Key.JSONWebKey)
	if err != nil {
		return "", "", err
	}

	return alg, tp, nil
}

func inferAlg(key *jwk.JWK) (string, error) {
	kt, err := key.KeyType()
	if err != nil {
		return "", err
	}

	algo, err := verifiable.KeyTypeToJWSAlgo(kt)
	if err != nil {
		return "", err
	}

	return algo.Name()
}

// GetKeyID return id of key used for signing.
func (s *JWSSigner) GetKeyID() string {
	return s.keyID
}

// Algorithm return jwt algorithm.
func (s *JWSSigner) Algorithm() string {
	return s.algorithm
}

// SignJWT signs jwt token.
func (s *JWSSigner) SignJWT(_ jwt.SignParameters, data []byte) ([]byte, error) {
	return s.crypto.Sign(data, s.cryptoKID)
}

// CreateJWTHeaders provides JWS headers.
func (s *JWSSigner) CreateJWTHeaders(_ jwt.SignParameters) (jose.Headers, error) {
	return jose.Headers{
		jose.HeaderKeyID:     s.keyID,
		jose.HeaderAlgorithm: s.algorithm,
	}, nil
}
