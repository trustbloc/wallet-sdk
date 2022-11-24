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

func TestLocalKMS_Create(t *testing.T) {
	localKMS, err := localkms.NewKMS(nil)
	require.NoError(t, err)

	keyHandle, err := localKMS.Create(localkms.KeyTypeED25519)
	require.NoError(t, err)
	require.NotEmpty(t, keyHandle.PubKey)
	require.NotEmpty(t, keyHandle.KeyID)
}

func TestLocalKMS_ExportPubKey(t *testing.T) {
	localKMS, err := localkms.NewKMS(nil)
	require.NoError(t, err)

	keyHandle, err := localKMS.Create(localkms.KeyTypeED25519)
	require.NoError(t, err)
	require.NotEmpty(t, keyHandle.PubKey)
	require.NotEmpty(t, keyHandle.KeyID)

	_, err = localKMS.ExportPubKey(keyHandle.KeyID)
	require.Contains(t, err.Error(), "not implemented")
}

func TestLocalKMS_GetSignAlgorithm(t *testing.T) {
	localKMS, err := localkms.NewKMS(nil)
	require.NoError(t, err)

	keyHandle, err := localKMS.Create(localkms.KeyTypeED25519)
	require.NoError(t, err)
	require.NotEmpty(t, keyHandle.PubKey)
	require.NotEmpty(t, keyHandle.KeyID)

	_, err = localKMS.GetSigningAlgorithm(keyHandle.KeyID)
	require.Contains(t, err.Error(), "not implemented")
}

func TestLocalKMS_GetCrypto(t *testing.T) {
	localKMS, err := localkms.NewKMS(nil)
	require.NoError(t, err)

	crypto := localKMS.GetCrypto()
	require.NotNil(t, crypto)
}
