/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didjwk_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-go/doc/jose/jwk"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didjwk"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
)

func TestCreate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS := createTestKMS(t)

		jsonWebKey, err := localKMS.Create(localkms.KeyTypeED25519)
		require.NoError(t, err)

		didDoc, err := didjwk.Create(jsonWebKey)
		require.NoError(t, err)
		require.NotNil(t, didDoc)
	})
	t.Run("Nil JWK", func(t *testing.T) {
		didDoc, err := didjwk.Create(nil)
		require.Contains(t, err.Error(), "jwk object cannot be null/nil")
		require.Nil(t, didDoc)
	})
	t.Run("Fail to create verification method from JWK", func(t *testing.T) {
		didDoc, err := didjwk.Create(&api.JSONWebKey{JWK: &jwk.JWK{}})
		require.Contains(t, err.Error(),
			"convert JWK to public key bytes: unsupported public key type in kid ''")
		require.Nil(t, didDoc)
	})
}

func createTestKMS(t *testing.T) *localkms.KMS {
	t.Helper()

	kmsStore := localkms.NewMemKMSStore()

	localKMS, err := localkms.NewKMS(kmsStore)
	require.NoError(t, err)

	return localKMS
}
