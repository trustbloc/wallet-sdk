/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms

import (
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock"
	"github.com/hyperledger/aries-framework-go/pkg/secretlock/noop"
)

// InMemoryStore represents an in-memory database of keysets.
type InMemoryStore struct {
	keys map[string][]byte
}

// NewInMemoryStore returns a new InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{keys: map[string][]byte{}}
}

// Put stores the given key under the given keysetID.
func (k *InMemoryStore) Put(keysetID string, keyset []byte) error {
	k.keys[keysetID] = keyset

	return nil
}

// Get retrieves the key stored under the given keysetID. If no key is found, then an error is returned.
func (k *InMemoryStore) Get(keysetID string) ([]byte, error) {
	key, exists := k.keys[keysetID]
	if !exists {
		return nil, arieskms.ErrKeyNotFound
	}

	return key, nil
}

// Delete deletes the key stored under the given keysetID.
func (k *InMemoryStore) Delete(keysetID string) error {
	delete(k.keys, keysetID)

	return nil
}

// InMemoryStorageProvider represents an in-memory storage provide that can be used to satisfy the Aries KMS
// Provider interface.
type InMemoryStorageProvider struct {
	Storage arieskms.Store
}

// NewInMemoryStorageProvider returns a new InMemoryStorageProvider.
func NewInMemoryStorageProvider() *InMemoryStorageProvider {
	return &InMemoryStorageProvider{Storage: NewInMemoryStore()}
}

// StorageProvider returns an in-memory arieskms.Store implemenation.
func (p *InMemoryStorageProvider) StorageProvider() arieskms.Store {
	if p.Storage != nil {
		return p.Storage
	}

	return NewInMemoryStore()
}

// SecretLock returns the Aries no-op secretlock.Service implementation.
func (p *InMemoryStorageProvider) SecretLock() secretlock.Service {
	return &noop.NoLock{}
}
