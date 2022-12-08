/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ion_test

import (
	"errors"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/stretchr/testify/require"

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
		localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
		require.NoError(t, err)

		kid, pk, err := localKMS.Create(kms.ED25519Type)
		require.NoError(t, err)

		vm := &did.VerificationMethod{
			ID:    "#" + kid,
			Value: pk,
			Type:  "Ed25519VerificationKey2018",
		}

		c, err := NewCreator(localKMS)
		require.NoError(t, err)

		doc, err := c.Create(vm)
		require.NoError(t, err)
		require.NotNil(t, doc)
		require.NotNil(t, doc.DIDDocument)
		require.NotEmpty(t, doc.DIDDocument.VerificationMethod)
		require.NotNil(t, doc.DIDDocument.VerificationMethod[0])
		require.Equal(t, pk, doc.DIDDocument.VerificationMethod[0].Value)
	})

	t.Run("fail to create update/recovery keys", func(t *testing.T) {
		expectErr := errors.New("expected error")

		badKMS := mockKeyWriter(func(keyType kms.KeyType) (string, []byte, error) {
			return "", nil, expectErr
		})

		c, err := NewCreator(badKMS)
		require.NoError(t, err)

		doc, err := c.Create(nil)
		require.ErrorIs(t, err, expectErr)
		require.Nil(t, doc)
	})
}

type mockKeyWriter func(keyType kms.KeyType) (string, []byte, error)

func (kw mockKeyWriter) Create(keyType kms.KeyType) (string, []byte, error) {
	return kw(keyType)
}
