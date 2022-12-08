/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms_test

import (
	"testing"

	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/localkms"
)

func TestLocalKMS_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
		require.NoError(t, err)

		keyID, key, err := localKMS.Create(arieskms.ED25519Type)
		require.NoError(t, err)
		require.NotEmpty(t, keyID)
		require.NotEmpty(t, key)
	})
}

func TestLocalKMS_GetKey(t *testing.T) {
	localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
	require.NoError(t, err)

	key, err := localKMS.ExportPubKey("KeyID")
	require.EqualError(t, err, "not implemented")
	require.Empty(t, key)
}

func TestLocalKMS_CustomStore(t *testing.T) {
	localKMS, err := localkms.NewLocalKMS(&localkms.Config{
		Storage: newMockStorage(),
	})
	require.NoError(t, err)

	key, err := localKMS.ExportPubKey("KeyID")
	require.EqualError(t, err, "not implemented")
	require.Empty(t, key)
}

func TestLocalKMS_GetSignAlgorithm(t *testing.T) {
	localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
	require.NoError(t, err)

	key, err := localKMS.GetSigningAlgorithm("KeyID")
	require.EqualError(t, err, "not implemented")
	require.Empty(t, key)
}

func TestLocalKMS_GetCrypto(t *testing.T) {
	localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
	require.NoError(t, err)

	crypto := localKMS.GetCrypto()
	require.NotNil(t, crypto)
}

func TestLocalKMS_GetAriesKMS(t *testing.T) {
	localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
	require.NoError(t, err)

	ariesKMS := localKMS.GetAriesKMS()
	require.NotNil(t, ariesKMS)
}

type mockStorage struct {
	keys map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{keys: map[string][]byte{}}
}

func (k *mockStorage) Put(keysetID string, keyset []byte) error {
	k.keys[keysetID] = keyset

	return nil
}

func (k *mockStorage) Get(keysetID string) ([]byte, error) {
	key, exists := k.keys[keysetID]
	if !exists {
		return nil, arieskms.ErrKeyNotFound
	}

	return key, nil
}

func (k *mockStorage) Delete(keysetID string) error {
	delete(k.keys, keysetID)

	return nil
}
