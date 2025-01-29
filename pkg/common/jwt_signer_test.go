/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common_test

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/doc/jose/jwk/jwksupport"
	"github.com/trustbloc/kms-go/doc/util/jwkkid"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/jwt"

	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/models"
)

func TestNewJWSSigner(t *testing.T) {
	mockKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	mockJWK, err := jwkkid.BuildJWK(mockKey, kms.ED25519Type)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		successCases := []struct {
			name        string
			vm          *models.VerificationMethod
			expectedAlg string
		}{
			{
				name: "VM type " + common.Ed25519VerificationKey2018,
				vm: &models.VerificationMethod{
					ID:   "testKeyID",
					Type: common.Ed25519VerificationKey2018,
					Key: models.VerificationKey{
						Raw: mockKey,
					},
				},
				expectedAlg: common.EdDSA,
			},
			{
				name: "VM type " + common.JSONWebKey2020 + " with Ed25519 JWK",
				vm: &models.VerificationMethod{
					ID:   "testKeyID",
					Type: common.JSONWebKey2020,
					Key: models.VerificationKey{
						JSONWebKey: mockJWK,
					},
				},
				expectedAlg: common.EdDSA,
			},
			{
				name: "VM with P384 JWK",
				vm: &models.VerificationMethod{
					ID:   "testKeyID",
					Type: common.JSONWebKey2020,
					Key: models.VerificationKey{
						JSONWebKey: getECKey(t),
					},
				},
				expectedAlg: common.ES384,
			},
		}

		for _, successCase := range successCases {
			t.Run(successCase.name, func(t *testing.T) {
				signer, err := common.NewJWSSigner(
					successCase.vm,
					&cryptoMock{})
				require.NoError(t, err)
				require.NotNil(t, signer)
				alg := signer.Algorithm()
				require.Equal(t, successCase.expectedAlg, alg)
			})
		}
	})

	t.Run("Invalid raw ed25519 key", func(t *testing.T) {
		_, err := common.NewJWSSigner(
			&models.VerificationMethod{
				ID:   "testKeyID",
				Type: common.Ed25519VerificationKey2018,
			},
			&cryptoMock{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "generating thumbprint for ed25519 key")
	})

	t.Run("Invalid verificationType", func(t *testing.T) {
		_, err := common.NewJWSSigner(
			&models.VerificationMethod{
				ID:   "testKeyID",
				Type: "Invalid",
			},
			&cryptoMock{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "verification method type 'Invalid' not supported")
	})

	t.Run("missing JWK", func(t *testing.T) {
		_, err := common.NewJWSSigner(
			&models.VerificationMethod{
				ID:   "testKeyID",
				Type: common.JSONWebKey2020,
			},
			&cryptoMock{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing jwk")
	})

	t.Run("invalid JWK", func(t *testing.T) {
		_, err := common.NewJWSSigner(
			&models.VerificationMethod{
				ID:   "testKeyID",
				Type: common.JSONWebKey2020,
				Key: models.VerificationKey{
					JSONWebKey: &jwk.JWK{},
				},
			},
			&cryptoMock{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "creating crypto thumbprint for JWK")
	})

	t.Run("Missed crypto", func(t *testing.T) {
		_, err := common.NewJWSSigner(
			&models.VerificationMethod{
				ID:   "testKeyID",
				Type: common.Ed25519VerificationKey2018,
				Key: models.VerificationKey{
					Raw: mockKey,
				},
			},
			nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "NO_CRYPTO_PROVIDED")
	})
}

func TestJWSSigner_Sign(t *testing.T) {
	mockKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		signer, err := common.NewJWSSigner(
			&models.VerificationMethod{
				ID:   "testKeyID",
				Type: "Ed25519VerificationKey2018",
				Key:  models.VerificationKey{Raw: mockKey},
			},
			&cryptoMock{Signature: []byte("mock sig")})
		require.NoError(t, err)

		sig, err := signer.SignJWT(jwt.SignParameters{}, []byte("test data"))

		require.NoError(t, err)
		require.Equal(t, sig, []byte("mock sig"))

		require.Equal(t, "testKeyID", signer.GetKeyID())
		require.Equal(t, "EdDSA", signer.Algorithm())

		headers, err := signer.CreateJWTHeaders(jwt.SignParameters{})
		require.NoError(t, err)
		require.Equal(t, "testKeyID", headers["kid"])
		require.Equal(t, "EdDSA", headers["alg"])
	})

	t.Run("Failed", func(t *testing.T) {
		s, err := common.NewJWSSigner(
			&models.VerificationMethod{
				ID:   "testKeyID",
				Type: "Ed25519VerificationKey2018",
				Key:  models.VerificationKey{Raw: mockKey},
			},
			&cryptoMock{Err: errors.New("test error")})
		require.NoError(t, err)

		_, err = s.SignJWT(jwt.SignParameters{}, []byte("test data"))
		require.Error(t, err)
	})
}

func getECKey(t *testing.T) *jwk.JWK {
	t.Helper()

	crv := elliptic.P384()
	privateKey, err := ecdsa.GenerateKey(crv, rand.Reader)
	require.NoError(t, err)

	j, err := jwksupport.JWKFromKey(privateKey)
	require.NoError(t, err)

	return j
}

type cryptoMock struct {
	Signature []byte
	Err       error
}

func (c *cryptoMock) Sign([]byte, string) ([]byte, error) {
	return c.Signature, c.Err
}

// Verify is not yet defined.
func (c *cryptoMock) Verify([]byte, []byte, string) error {
	return nil
}
