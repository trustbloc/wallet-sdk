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
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"
	arieslocalkms "github.com/hyperledger/aries-framework-go/pkg/kms/localkms"

	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
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
func NewLocalKMS(cfg *Config) (*LocalKMS, error) {
	ariesLocalKMS, err := arieslocalkms.New("ThisIs://Unused", &InMemoryStorageProvider{
		Storage: cfg.Storage,
	})
	if err != nil {
		return nil, err
	}

	ariesCrypto, err := tinkcrypto.New()
	if err != nil {
		return nil, err
	}

	return &LocalKMS{ariesLocalKMS: ariesLocalKMS, ariesCrypto: ariesCrypto}, nil
}

// Create creates a keyset of the given keyType and then writes it to storage.
// The keyID and raw public key bytes of the newly generated keyset are returned.
func (k *LocalKMS) Create(keyType arieskms.KeyType) (string, []byte, error) {
	// TODO: for keys that support afgo JWK, return afgo JWK
	return k.ariesLocalKMS.CreateAndExportPubKeyBytes(keyType)
}

// ExportPubKey returns the public key associated with the given keyID as raw bytes.
func (k *LocalKMS) ExportPubKey(string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// GetCrypto returns Crypto instance that can perform crypto ops with keys created by this kms.
func (k *LocalKMS) GetCrypto() goapi.Crypto {
	return &AriesCryptoWrapper{
		cryptosKMS:    k.ariesLocalKMS,
		wrappedCrypto: k.ariesCrypto,
	}
}

// GetAriesKMS returns the underlying Aries local KMS instance.
//
// Deprecated: This method will be removed in a future version.
func (k *LocalKMS) GetAriesKMS() *arieslocalkms.LocalKMS {
	return k.ariesLocalKMS
}
