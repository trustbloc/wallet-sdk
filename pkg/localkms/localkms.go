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

	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/secretlock/noop"
	arieskms "github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/kms-go/wrapper/api"
	"github.com/trustbloc/kms-go/wrapper/localsuite"

	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// LocalKMS is a KMS implementation that uses Google's Tink crypto library.
// Private keys may intermittently reside in local memory with this implementation so
// keep this consideration in mind when deciding whether to use this or not.
type LocalKMS struct {
	AriesSuite api.Suite
	keyCreator api.KeyCreator
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

	suite, err := localsuite.NewLocalCryptoSuite("ThisIs://Unused", cfg.Storage, &noop.NoLock{})
	if err != nil {
		return nil, walleterror.NewExecutionError(module, InitialisationFailedCode, InitialisationFailedError, err)
	}

	keyCreator, err := suite.KeyCreator()
	if err != nil {
		return nil, walleterror.NewExecutionError(module, InitialisationFailedCode, InitialisationFailedError, err)
	}

	return &LocalKMS{
		AriesSuite: suite,
		keyCreator: keyCreator,
	}, nil
}

// Create creates a keyset of the given keyType and then writes it to storage.
//
// Returns:
//   - key ID for the newly generated keyset.
//   - JWK object for the keyset's public key.
func (k *LocalKMS) Create(keyType arieskms.KeyType) (string, *jwk.JWK, error) {
	pkJWK, err := k.keyCreator.Create(keyType)
	if err != nil {
		return "", nil, walleterror.NewExecutionError(module, CreateKeyFailedCode, CreateKeyFailedError, err)
	}

	return pkJWK.KeyID, pkJWK, nil
}

// ExportPubKey returns the public key associated with the given keyID as a JWK byte string.
func (k *LocalKMS) ExportPubKey(string) (*jwk.JWK, error) {
	return nil, errors.New("not implemented")
}

// GetCrypto returns Crypto instance that can perform crypto ops with keys created by this kms.
func (k *LocalKMS) GetCrypto() goapi.Crypto {
	return &AriesCryptoWrapper{
		cryptoSuite: k.AriesSuite,
	}
}
