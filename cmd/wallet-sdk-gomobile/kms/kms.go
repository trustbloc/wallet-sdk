/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package kms contains a KMS implementation.
package kms

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// A KeyManager manages key creation, storage, retrieval, and other related functionality.
type KeyManager struct{}

// NewKeyManager returns a new KeyManager instance.
func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

// Create a new key/keyset/key handle of type keyType
// Some key types may require additional attributes described in `opts`. // TODO: Format of opts to be determined.
// Returns a key handle, which contains both the key ID and actual private key bytes.
func (k *KeyManager) Create(keyType, opts string) (*api.KeyHandle, error) {
	return &api.KeyHandle{
		Key:     []byte("Key Bytes"),
		KeyType: keyType,
	}, nil
}

// Rotate a key referenced by keyID and return a new handle of a keyset including old key and
// new key of type keyType. It also returns the updated keyID as the first return value
// Some key types may require additional attributes described in `opts` // TODO: Format of opts to be determined.
// Returns: a key handle, which contains both the new key ID and actual private key bytes.
func (k *KeyManager) Rotate(keyType, keyID, opts string) (*api.KeyHandle, error) {
	return &api.KeyHandle{
		Key:     []byte("Key Bytes"),
		KeyType: keyType,
	}, nil
}

// Import will import privKey into the KMS storage for the given keyType then returns the new key id and
// the newly persisted KeyHandle.
// privKey possible types are: *ecdsa.PrivateKey and ed25519.PrivateKey. // TODO: Determine how these restrictions work.
// kt possible types are signing key types only (ECDSA keys or Ed25519)
// opts allows setting the keysetID of the imported key using WithKeyID() option. If the ID is already used,
// then an error is returned. // TODO: Format of opts to be determined.
// Returns: a key handle, which contains both the new key ID and actual private key bytes.
// An error/exception will be returned/thrown if there is an import failure (key empty, invalid, doesn't match
// keyType, unsupported keyType or storing of key failed)
// TODO: Consider renaming this method to avoid keyword collision and automatic renaming to _import in Java.
func (k *KeyManager) Import(privateKey *api.KeyHandle, keyType, opts string) (*api.KeyHandle, error) {
	return privateKey, nil
}

// GetKeyHandle return the key handle for the given keyID
// Returns the private key handle
//   - Error if failure.
func (k *KeyManager) GetKeyHandle(keyID string) (*api.KeyHandle, error) {
	return &api.KeyHandle{
		Key:     []byte("Key Bytes"),
		KeyType: "SomeKeyType",
	}, nil
}

// Export will fetch a key referenced by id then gets its public key in raw bytes and returns it.
// The key must be an asymmetric key.
// Returns:
//   - A key handle, which contains both the key type and public key bytes.
//   - Error if it fails to export the public key bytes.
func (k *KeyManager) Export(keyID string) (*api.KeyHandle, error) {
	return &api.KeyHandle{
		Key:     []byte("SomeKey"),
		KeyType: "SomeKeyType",
	}, nil
}
