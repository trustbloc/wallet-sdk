/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ion_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/did"

	"github.com/trustbloc/wallet-sdk/pkg/common"
	. "github.com/trustbloc/wallet-sdk/pkg/did/creator/ion"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
)

func TestNewCreator(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c, err := NewCreator(nil)
		require.NoError(t, err)
		require.NotNil(t, c)
	})
}

func TestCreator_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		localKMS := createTestKMS(t)

		kid, pkJWK, err := localKMS.Create(kms.ED25519Type)
		require.NoError(t, err)

		vm, err := did.NewVerificationMethodFromJWK("#"+kid, common.JSONWebKey2020, "", pkJWK)
		require.NoError(t, err)

		c, err := NewCreator(localKMS)
		require.NoError(t, err)

		doc, err := c.Create(vm)
		require.NoError(t, err)
		require.NotNil(t, doc)
		require.NotNil(t, doc.DIDDocument)
		require.NotEmpty(t, doc.DIDDocument.VerificationMethod)
		require.NotNil(t, doc.DIDDocument.VerificationMethod[0])
		require.Equal(t, pkJWK, doc.DIDDocument.VerificationMethod[0].JSONWebKey())
	})

	t.Run("fail to create update/recovery keys", func(t *testing.T) {
		expectErr := errors.New("expected error")

		badKMS := mockKeyWriter(func(keyType kms.KeyType) (string, *jwk.JWK, error) {
			return "", nil, expectErr
		})

		c, err := NewCreator(badKMS)
		require.NoError(t, err)

		doc, err := c.Create(nil)
		require.ErrorIs(t, err, expectErr)
		require.Nil(t, doc)
	})
}

func createTestKMS(t *testing.T) *localkms.LocalKMS {
	t.Helper()

	kmsStore := localkms.NewMemKMSStore()

	localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: kmsStore})
	require.NoError(t, err)

	return localKMS
}

type mockKeyWriter func(keyType kms.KeyType) (string, *jwk.JWK, error)

func (kw mockKeyWriter) Create(keyType kms.KeyType) (string, *jwk.JWK, error) {
	return kw(keyType)
}
