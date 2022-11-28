/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package localkms contains a KMS implementation that uses local storage.
// It is not intended for production use and may not be secure.
package localkms

import (
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapilocalkms "github.com/trustbloc/wallet-sdk/pkg/localkms"
)

// KeyTypeED25519 is the name recognized by the Create method for creating an ED25519 keyset.
const KeyTypeED25519 = arieskms.ED25519

// Store defines the storage capability for local KMS.
type Store interface {
	// Put stores the given key under the given keysetID.
	Put(keysetID string, key []byte) error
	// Get retrieves the key stored under the given keysetID. If no key is found, then an error is returned.
	Get(keysetID string) (key []byte, err error)
	// Delete deletes the key stored under the given keysetID.
	Delete(keysetID string) error
}

// KMS is a KMS implementation that uses local storage.
// It is not intended for production use and may not be secure.
type KMS struct {
	goAPILocalKMS *goapilocalkms.LocalKMS
}

// NewKMS returns a new local KMS instance.
func NewKMS(kmsStore Store) (*KMS, error) {
	goAPILocalKMS, err := goapilocalkms.NewLocalKMS(&goapilocalkms.Config{
		Storage: kmsStore,
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

// GetSigningAlgorithm returns signing algorithm associated with the given keyID.
func (k *KMS) GetSigningAlgorithm(keyID string) (string, error) {
	return k.goAPILocalKMS.GetSigningAlgorithm(keyID)
}

// GetCrypto returns Crypto instance that can perform crypto ops with keys created by this kms.
func (k *KMS) GetCrypto() api.Crypto {
	return k.goAPILocalKMS.GetCrypto()
}
