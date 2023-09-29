/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ion_test

import (
	"crypto/x509"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/spi/kms"

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

		_, pkJWK, err := localKMS.Create(kms.ED25519Type)
		require.NoError(t, err)

		doc, err := CreateLongForm(pkJWK)
		require.NoError(t, err)
		require.NotNil(t, doc)
		require.NotNil(t, doc.DIDDocument)
		require.NotEmpty(t, doc.DIDDocument.VerificationMethod)
		require.NotNil(t, doc.DIDDocument.VerificationMethod[0])

		// localkms returns a key without these fields set, whereas the fields are set
		// when marshalling/unmarshalling in creating the did doc.
		pkJWK.Certificates = []*x509.Certificate{}
		pkJWK.CertificateThumbprintSHA1 = []byte{}
		pkJWK.CertificateThumbprintSHA256 = []byte{}

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

func TestCreateLongForm(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS := createTestKMS(t)

		_, signingJWK, err := localKMS.Create(kms.ED25519)
		require.NoError(t, err)

		didDoc, err := CreateLongForm(signingJWK)
		require.NoError(t, err)
		require.NotNil(t, didDoc)
	})
	t.Run("Nil JWK", func(t *testing.T) {
		didDoc, err := CreateLongForm(nil)
		require.Contains(t, err.Error(), "jwk object cannot be null/nil")
		require.Nil(t, didDoc)
	})
	t.Run("Fail to create verification method from JWK", func(t *testing.T) {
		jsonWebKey := &jwk.JWK{}

		didDoc, err := CreateLongForm(jsonWebKey)
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

type mockKeyWriter func(keyType kms.KeyType) (string, *jwk.JWK, error)

func (kw mockKeyWriter) Create(keyType kms.KeyType) (string, *jwk.JWK, error) {
	return kw(keyType)
}
