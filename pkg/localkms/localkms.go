/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package localkms contains a KMS implementation that uses Google's Tink crypto library.
// Private keys may intermittently reside in local memory with this implementation so
// keep this consideration in mind when deciding whether to use this or not.
package localkms

import (
	"errors"

	"github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk"
	"github.com/hyperledger/aries-framework-go/pkg/doc/util/jwkkid"
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"
	arieslocalkms "github.com/hyperledger/aries-framework-go/pkg/kms/localkms"

	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// LocalKMS is a KMS implementation that uses Google's Tink crypto library.
// Private keys may intermittently reside in local memory with this implementation so
// keep this consideration in mind when deciding whether to use this or not.
type LocalKMS struct {
	ariesLocalKMS *arieslocalkms.LocalKMS
	ariesCrypto   crypto.Crypto
}

// Config is config for local kms constructor.
type Config struct {
	Storage arieskms.Store
}

// NewLocalKMS returns a new Local KMS.
func NewLocalKMS(cfg Config) (*LocalKMS, error) {
	if cfg.Storage == nil {
		return nil, errors.New("cfg.Storage cannot be nil")
	}

	ariesLocalKMS, err := arieslocalkms.New("ThisIs://Unused", &storageProvider{
		Storage: cfg.Storage,
	})
	if err != nil {
		return nil, walleterror.NewExecutionError(module, InitialisationFailedCode, InitialisationFailedError, err)
	}

	ariesCrypto, err := tinkcrypto.New()
	if err != nil {
		return nil, walleterror.NewExecutionError(module, InitialisationFailedCode, InitialisationFailedError, err)
	}

	return &LocalKMS{ariesLocalKMS: ariesLocalKMS, ariesCrypto: ariesCrypto}, nil
}

// Create creates a keyset of the given keyType and then writes it to storage.
//
// Returns:
//   - key ID for the newly generated keyset.
//   - JSON byte string containing the JWK for the keyset's public key.
//   - JWK object for the same public key.
func (k *LocalKMS) Create(keyType arieskms.KeyType) (string, *jwk.JWK, error) {
	keyID, publicKey, err := k.ariesLocalKMS.CreateAndExportPubKeyBytes(keyType)
	if err != nil {
		return "", nil, walleterror.NewExecutionError(module, CreateKeyFailedCode, CreateKeyFailedError, err)
	}

	pkJWK, err := jwkkid.BuildJWK(publicKey, keyType)
	if err != nil {
		return "", nil, walleterror.NewExecutionError(module, CreateKeyFailedCode, CreateKeyFailedError, err)
	}

	pkJWK.KeyID = keyID

	return keyID, pkJWK, nil
}

// ExportPubKey returns the public key associated with the given keyID as a JWK byte string.
func (k *LocalKMS) ExportPubKey(string) (*jwk.JWK, error) {
	return nil, errors.New("not implemented")
}

// GetCrypto returns Crypto instance that can perform crypto ops with keys created by this kms.
func (k *LocalKMS) GetCrypto() goapi.Crypto {
	return &AriesCryptoWrapper{
		cryptosKMS:    k.ariesLocalKMS,
		wrappedCrypto: k.ariesCrypto,
	}
}
