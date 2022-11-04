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

type inMemoryStorageProvider struct {
	keys map[string][]byte
}

func newInMemoryStorageProvider() *inMemoryStorageProvider {
	return &inMemoryStorageProvider{keys: map[string][]byte{}}
}

func (k *inMemoryStorageProvider) Put(keysetID string, keyset []byte) error {
	k.keys[keysetID] = keyset

	return nil
}

func (k *inMemoryStorageProvider) Get(keysetID string) ([]byte, error) {
	key, exists := k.keys[keysetID]
	if !exists {
		return nil, arieskms.ErrKeyNotFound
	}

	return key, nil
}

func (k *inMemoryStorageProvider) Delete(keysetID string) error {
	delete(k.keys, keysetID)

	return nil
}

type provider struct{}

func (p *provider) StorageProvider() arieskms.Store {
	return newInMemoryStorageProvider()
}

func (p *provider) SecretLock() secretlock.Service {
	return &noop.NoLock{}
}
