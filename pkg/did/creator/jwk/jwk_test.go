/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package jwk_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/spi/kms"

	. "github.com/trustbloc/wallet-sdk/pkg/did/creator/jwk"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
)

func TestCreate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS := createTestKMS(t)

		_, jsonWebKey, err := localKMS.Create(kms.ED25519)
		require.NoError(t, err)

		didDoc, err := Create(jsonWebKey)
		require.NoError(t, err)
		require.NotNil(t, didDoc)
	})
	t.Run("Nil JWK", func(t *testing.T) {
		didDoc, err := Create(nil)
		require.Contains(t, err.Error(), "jwk object cannot be nil")
		require.Nil(t, didDoc)
	})
	t.Run("Fail to create verification method from JWK", func(t *testing.T) {
		didDoc, err := Create(&jwk.JWK{})
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
