/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	arieskms "github.com/trustbloc/kms-go/kms"
	kmsapi "github.com/trustbloc/kms-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/pkg/localkms"
)

func TestNewLocalKMS(t *testing.T) {
	localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: nil})
	require.EqualError(t, err, "cfg.Storage cannot be nil")
	require.Nil(t, localKMS)
}

func TestLocalKMS_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS := createTestKMS(t)

		keyID, pkJWK, err := localKMS.Create(kmsapi.ED25519Type)
		require.NoError(t, err)
		require.NotEmpty(t, keyID)
		require.NotNil(t, pkJWK)

		keyID, pkJWK, err = localKMS.Create(kmsapi.ECDSAP384IEEEP1363)
		require.NoError(t, err)
		require.NotEmpty(t, keyID)
		require.NotNil(t, pkJWK)
	})

	t.Run("Invalid key type", func(t *testing.T) {
		localKMS := createTestKMS(t)

		_, _, err := localKMS.Create("INVALID")
		require.Error(t, err)
	})

	t.Run("key type not supported", func(t *testing.T) {
		localKMS := createTestKMS(t)

		_, _, err := localKMS.Create(kmsapi.AES256GCMType)
		require.Error(t, err)
	})
}

func TestLocalKMS_GetKey(t *testing.T) {
	localKMS := createTestKMS(t)

	key, err := localKMS.ExportPubKey("KeyID")
	require.EqualError(t, err, "not implemented")
	require.Nil(t, key)
}

func TestLocalKMS_CustomStore(t *testing.T) {
	localKMS, err := localkms.NewLocalKMS(localkms.Config{
		Storage: newMockStorage(),
	})
	require.NoError(t, err)

	key, err := localKMS.ExportPubKey("KeyID")
	require.EqualError(t, err, "not implemented")
	require.Nil(t, key)
}

func TestLocalKMS_GetCrypto(t *testing.T) {
	localKMS := createTestKMS(t)

	crypto := localKMS.GetCrypto()
	require.NotNil(t, crypto)
}

func createTestKMS(t *testing.T) *localkms.LocalKMS {
	t.Helper()

	kmsStore := localkms.NewMemKMSStore()

	localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: kmsStore})
	require.NoError(t, err)

	return localKMS
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
