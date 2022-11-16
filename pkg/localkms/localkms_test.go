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
		localKMS, err := localkms.NewLocalKMS()
		require.NoError(t, err)

		keyID, key, err := localKMS.Create(arieskms.ED25519Type)
		require.NoError(t, err)
		require.NotEmpty(t, keyID)
		require.NotEmpty(t, key)
	})
	t.Run("Unsupported key type", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS()
		require.NoError(t, err)

		keyID, key, err := localKMS.Create("SomeUnsupportedKeyType")
		require.EqualError(t, err, "key type SomeUnsupportedKeyType not supported")
		require.Empty(t, keyID)
		require.Empty(t, key)
	})
}

func TestLocalKMS_GetKey(t *testing.T) {
	localKMS, err := localkms.NewLocalKMS()
	require.NoError(t, err)

	key, err := localKMS.ExportPubKey("KeyID")
	require.EqualError(t, err, "not implemented")
	require.Empty(t, key)
}

func TestLocalKMS_GetSignAlgorithm(t *testing.T) {
	localKMS, err := localkms.NewLocalKMS()
	require.NoError(t, err)

	key, err := localKMS.GetSigningAlgorithm("KeyID")
	require.EqualError(t, err, "not implemented")
	require.Empty(t, key)
}
