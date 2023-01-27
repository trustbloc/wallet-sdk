/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms

// MemKMSStore is a simple in-memory KMS store implementation.
type MemKMSStore struct {
	keys map[string][]byte
}

// NewMemKMSStore returns a new MemKMSStore.
func NewMemKMSStore() *MemKMSStore {
	return &MemKMSStore{keys: map[string][]byte{}}
}

// Put stores the given key under the given keysetID.
func (m *MemKMSStore) Put(keysetID string, key []byte) error {
	m.keys[keysetID] = key

	return nil
}

// Get retrieves the key stored under the given keysetID.
// The returned result indicates whether a key was found and, if so, the key bytes.
// If a key was not found, then Result.Found will be false and no error will be returned.
func (m *MemKMSStore) Get(keysetID string) (*Result, error) {
	key, exists := m.keys[keysetID]
	if !exists {
		return &Result{Found: false}, nil
	}

	return &Result{Found: true, Key: key}, nil
}
