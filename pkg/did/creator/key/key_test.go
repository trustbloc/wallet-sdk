/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package key_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-go/doc/jose/jwk"

	kmsspi "github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/wallet-sdk/pkg/did/creator/key"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
)

func TestCreate(t *testing.T) {
	t.Run("Using an ED25519 key", func(t *testing.T) {
		localKMS := createTestKMS(t)

		_, jsonWebKey, err := localKMS.Create(kmsspi.ED25519)
		require.NoError(t, err)

		didDoc, err := key.Create(jsonWebKey)
		require.NoError(t, err)
		require.NotNil(t, didDoc)
	})
	t.Run("Using a P-384 key", func(t *testing.T) {
		localKMS := createTestKMS(t)

		_, jsonWebKey, err := localKMS.Create(kmsspi.ECDSAP384IEEEP1363)
		require.NoError(t, err)

		didDoc, err := key.Create(jsonWebKey)
		require.NoError(t, err)
		require.NotNil(t, didDoc)
	})
	t.Run("Nil JWK", func(t *testing.T) {
		didDoc, err := key.Create(nil)
		require.Contains(t, err.Error(), "jwk object cannot be nil")
		require.Nil(t, didDoc)
	})
	t.Run("Fail to get public key bytes", func(t *testing.T) {
		didDoc, err := key.Create(&jwk.JWK{Crv: "Ed25519"})
		require.Contains(t, err.Error(), "unsupported public key type in kid ''")
		require.Nil(t, didDoc)
	})
	t.Run("Fail to create verification method from JWK", func(t *testing.T) {
		didDoc, err := key.Create(&jwk.JWK{})
		require.Contains(t, err.Error(),
			"convert JWK to public key bytes: unsupported public key type in kid ''")
		require.Nil(t, didDoc)
	})
}

func createTestKMS(t *testing.T) *localkms.LocalKMS {
	t.Helper()

	kmsStore := localkms.NewMemKMSStore()

	localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: kmsStore})
	require.NoError(t, err)

	return localKMS
}
