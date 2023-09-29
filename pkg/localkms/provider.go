/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms

import (
	arieskms "github.com/trustbloc/kms-go/kms"
)

// MemKMSStore represents an in-memory database of keysets.
type MemKMSStore struct {
	keys map[string][]byte
}

// NewMemKMSStore returns a new MemKMSStore.
func NewMemKMSStore() *MemKMSStore {
	return &MemKMSStore{keys: map[string][]byte{}}
}

// Put stores the given key under the given keysetID.
func (k *MemKMSStore) Put(keysetID string, keyset []byte) error {
	k.keys[keysetID] = keyset

	return nil
}

// Get retrieves the key stored under the given keysetID. If no key is found, then an error is returned.
func (k *MemKMSStore) Get(keysetID string) ([]byte, error) {
	key, exists := k.keys[keysetID]
	if !exists {
		return nil, arieskms.ErrKeyNotFound
	}

	return key, nil
}

// Delete deletes the key stored under the given keysetID.
// This won't normally be used since we don't expose the underlyinng Rotate method from the Aries local KMS
// in the local KMS implementation here. However, if someone uses the deprecated GetAriesKMS method, they could
// potentially call Rotate directly, so the Delete method is implemented here just in case.
func (k *MemKMSStore) Delete(keysetID string) error {
	delete(k.keys, keysetID)

	return nil
}
