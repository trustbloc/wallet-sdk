/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package localkms contains a KMS implementation that uses local storage.
// It is not intended for production use and may not be secure.
package localkms

import (
	"errors"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/crypto"
	"github.com/hyperledger/aries-framework-go/pkg/crypto/tinkcrypto"
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"
	arieslocalkms "github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

// LocalKMS is a KMS implementation that uses local storage.
// It is not intended for production use and may not be secure.
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
// Currently, this method only supports creating ED25519 keys.
func (k *LocalKMS) Create(keyType arieskms.KeyType) (string, []byte, error) {
	// The CreateAndExportPubKeyBytes method we use from the Aries Local KMS implementation returns raw bytes for
	// some key types, and marshalled key bytes for others. If we want to support other key types in the future
	// then we will need to ensure that either this method either only returns raw bytes for consistency (by converting
	// to raw bytes as needed) or that the KeyWriter interface documentation is updated to make the expected key format
	// clear.
	if keyType != arieskms.ED25519Type {
		return "", nil, fmt.Errorf("key type %s not supported", keyType)
	}

	return k.ariesLocalKMS.CreateAndExportPubKeyBytes(keyType)
}

// ExportPubKey returns the public key associated with the given keyID as raw bytes.
func (k *LocalKMS) ExportPubKey(string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// GetSigningAlgorithm returns sign algorithm associated with the given keyID.
func (k *LocalKMS) GetSigningAlgorithm(keyID string) (string, error) {
	return "", errors.New("not implemented")
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
