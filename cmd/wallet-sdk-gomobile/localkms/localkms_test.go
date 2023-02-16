/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
)

func TestNewKMS(t *testing.T) {
	kms, err := localkms.NewKMS(nil)
	require.EqualError(t, err, "kmsStore cannot be nil")
	require.Nil(t, kms)
}

func TestLocalKMS_Create(t *testing.T) {
	localKMS := createTestKMS(t)

	keyHandle, err := localKMS.Create(localkms.KeyTypeED25519)
	require.NoError(t, err)
	require.NotNil(t, keyHandle.JWK)
	require.NotEmpty(t, keyHandle.ID())
}

func TestLocalKMS_ExportPubKey(t *testing.T) {
	localKMS := createTestKMS(t)

	keyHandle, err := localKMS.Create(localkms.KeyTypeED25519)
	require.NoError(t, err)
	require.NotNil(t, keyHandle.JWK)
	require.NotEmpty(t, keyHandle.ID())

	_, err = localKMS.ExportPubKey(keyHandle.ID())
	require.Contains(t, err.Error(), "not implemented")
}

func TestLocalKMS_GetCrypto(t *testing.T) {
	localKMS := createTestKMS(t)

	crypto := localKMS.GetCrypto()
	require.NotNil(t, crypto)
}

func createTestKMS(t *testing.T) *localkms.KMS {
	t.Helper()

	kmsStore := localkms.NewMemKMSStore()

	localKMS, err := localkms.NewKMS(kmsStore)
	require.NoError(t, err)

	return localKMS
}
