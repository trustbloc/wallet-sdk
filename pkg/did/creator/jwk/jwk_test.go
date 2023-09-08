/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package jwk_test

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/doc/jose/jwk/jwksupport"
	"github.com/trustbloc/kms-go/doc/util/jwkkid"
	"github.com/trustbloc/kms-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/pkg/did/creator"
	. "github.com/trustbloc/wallet-sdk/pkg/did/creator/jwk"
)

func TestNewCreator(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := NewCreator()
		require.NotNil(t, c)
	})
}

func TestCreator_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testcases := []struct {
			name    string
			jwkFunc makeJWK
		}{
			{
				name:    "ed25519",
				jwkFunc: ed225519JWK,
			},
			{
				name:    "P256",
				jwkFunc: p256JWK,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				pubJWK := tc.jwkFunc(t)

				c := NewCreator()

				vm, err := did.NewVerificationMethodFromJWK("0", creator.JSONWebKey2020, "", pubJWK)
				require.NoError(t, err)

				jwkDID, err := c.Create(vm)
				require.NoError(t, err)

				require.NotNil(t, jwkDID)
				require.NotNil(t, jwkDID.DIDDocument)
				require.NotEmpty(t, jwkDID.DIDDocument.VerificationMethod)
				require.NotNil(t, jwkDID.DIDDocument.VerificationMethod[0])

				expectedThumb, err := pubJWK.Thumbprint(crypto.SHA256)
				require.NoError(t, err)

				actualThumb, err := jwkDID.DIDDocument.VerificationMethod[0].JSONWebKey().Thumbprint(crypto.SHA256)
				require.NoError(t, err)

				require.Equal(t, expectedThumb, actualThumb)
			})
		}
	})

	t.Run("fail to create did with invalid verification method", func(t *testing.T) {
		c := NewCreator()

		_, err := c.Create(&did.VerificationMethod{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating did:jwk DID Document")
	})
}

type makeJWK func(t *testing.T) *jwk.JWK

func p256JWK(t *testing.T) *jwk.JWK {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	pub := priv.PublicKey

	pubJWK, err := jwksupport.JWKFromKey(&pub)
	require.NoError(t, err)

	return pubJWK
}

func ed225519JWK(t *testing.T) *jwk.JWK {
	t.Helper()

	edpub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pubJWK, err := jwkkid.BuildJWK(edpub, kms.ED25519Type)
	require.NoError(t, err)

	return pubJWK
}
