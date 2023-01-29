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

	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapilocalkms "github.com/trustbloc/wallet-sdk/pkg/localkms"
)

// KeyTypeED25519 is the name recognized by the Create method for creating an ED25519 keyset.
const KeyTypeED25519 = arieskms.ED25519

// Result indicates the result of a key retrieval operation (see Store.Get for more info).
type Result struct {
	// Found indicates whether a key was found stored under the given keysetID.
	// If this is false, then Key should be nil. If this is true, then Key should not be nil.
	Found bool
	// Key is the retrieved key bytes.
	Key []byte
}

// Store defines the storage capability for local KMS.
type Store interface {
	// Put stores the given key under the given keysetID.
	Put(keysetID string, key []byte) error
	// Get retrieves the key stored under the given keysetID.
	// The returned result indicates whether a key was found and, if so, the key bytes.
	// If a key was not found, then Result.Found should be set accordingly - no error should be returned in this case.
	Get(keysetID string) (*Result, error)
}

// KMS is a KMS implementation that uses Google's Tink crypto library.
// Private keys may intermittently reside in local memory with this implementation so
// keep this consideration in mind when deciding whether to use this or not.
type KMS struct {
	goAPILocalKMS *goapilocalkms.LocalKMS
}

// NewKMS returns a new local KMS instance.
func NewKMS(kmsStore Store) (*KMS, error) {
	if kmsStore == nil {
		return nil, errors.New("kmsStore cannot be nil")
	}

	goAPILocalKMS, err := goapilocalkms.NewLocalKMS(goapilocalkms.Config{
		Storage: &kmsStoreWrapper{store: kmsStore},
	})
	if err != nil {
		return nil, err
	}

	return &KMS{goAPILocalKMS: goAPILocalKMS}, nil
}

// Create creates a keyset of the given keyType and then writes it to storage.
// The keyID and raw public key bytes of the newly generated keyset are returned.
// Currently, this method only supports creating ED25519 keys.
func (k *KMS) Create(keyType string) (*api.KeyHandle, error) {
	keyID, key, err := k.goAPILocalKMS.Create(arieskms.KeyType(keyType))
	if err != nil {
		return nil, err
	}

	return &api.KeyHandle{
		PubKey: key,
		KeyID:  keyID,
	}, nil
}

// ExportPubKey returns the public key associated with the given keyID as raw bytes.
func (k *KMS) ExportPubKey(keyID string) ([]byte, error) {
	return k.goAPILocalKMS.ExportPubKey(keyID)
}

// GetCrypto returns Crypto instance that can perform crypto ops with keys created by this kms.
func (k *KMS) GetCrypto() api.Crypto {
	return k.goAPILocalKMS.GetCrypto()
}

// kmsStoreWrapper is a wrapper around the Store interface defined here. Its purpose is to convert any Store.Get
// calls from the wrapped Store implementation into equivalent Aries local KMS store interface Get calls.
type kmsStoreWrapper struct {
	store Store
}

func (k *kmsStoreWrapper) Put(keysetID string, key []byte) error {
	return k.store.Put(keysetID, key)
}

func (k *kmsStoreWrapper) Get(keysetID string) ([]byte, error) {
	result, err := k.store.Get(keysetID)
	if err != nil {
		return nil, err
	}

	if !result.Found {
		return nil, arieskms.ErrKeyNotFound
	}

	return result.Key, nil
}

// Delete isn't used since we don't expose the Rotate method from the underlying Aries local KMS.
// This method is just here as it's required to satisfy the Aries KMS store interface.
func (k *kmsStoreWrapper) Delete(keysetID string) error {
	return nil
}
