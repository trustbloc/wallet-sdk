/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"testing"

	"github.com/go-jose/go-jose/v3"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/doc/util/jwkkid"
	"github.com/trustbloc/kms-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

func TestKeyHandle(t *testing.T) {
	pkb, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	key, err := jwkkid.BuildJWK(pkb, kms.ED25519Type)
	require.NoError(t, err)

	// for Ed25519, localkms returns a key without these fields set, whereas the
	// fields are set when marshalling/unmarshalling in creating the did doc.
	key.Certificates = []*x509.Certificate{}
	key.CertificateThumbprintSHA1 = []byte{}
	key.CertificateThumbprintSHA256 = []byte{}

	kid, err := jwkkid.CreateKID(pkb, kms.ED25519Type)
	require.NoError(t, err)

	key.KeyID = kid

	kh := &api.JSONWebKey{
		JWK: key,
	}

	t.Run("success", func(t *testing.T) {
		require.Equal(t, kid, kh.ID())

		keyString, e := kh.Serialize()
		require.NoError(t, e)

		require.NotEmpty(t, keyString)

		newKH, e := api.ParseJSONWebKey(keyString)
		require.NoError(t, e)

		require.Equal(t, kh, newKH)

		jwks := api.NewJSONWebKeySet()
		jwks.Append(kh)

		require.Equal(t, 1, jwks.Length())
		require.Equal(t, kh, jwks.AtIndex(0))
	})

	t.Run("fail to serialize", func(t *testing.T) {
		k := &api.JSONWebKey{}
		keyString, e := k.Serialize()
		require.Error(t, e)
		require.Empty(t, keyString)
		require.Contains(t, e.Error(), "json web key has no data to serialize")

		k = &api.JSONWebKey{
			JWK: &jwk.JWK{
				JSONWebKey: jose.JSONWebKey{
					Key: make(chan int),
				},
			},
		}
		keyString, e = k.Serialize()
		require.Error(t, e)
		require.Empty(t, keyString)
		require.Contains(t, e.Error(), "serializing json web key")
	})

	t.Run("fail to parse", func(t *testing.T) {
		empty, e := api.ParseJSONWebKey("{{{{{")
		require.Nil(t, empty)
		require.Error(t, e)
		require.Contains(t, e.Error(), "parsing json web key")
	})
}
