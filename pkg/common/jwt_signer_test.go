/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common_test

import (
	"errors"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/models"
)

func TestNewJWSSigner(t *testing.T) {
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
				},
				expectedAlg: common.EdDSA,
			},
			{
				name: "VM type " + common.JSONWebKey2020 + " with Ed25519 JWK",
				vm: &models.VerificationMethod{
					ID:   "testKeyID",
					Type: common.JSONWebKey2020,
					Key: models.VerificationKey{
						JSONWebKey: &jwk.JWK{
							Crv: "Ed25519",
						},
					},
				},
				expectedAlg: common.EdDSA,
			},
		}

		for _, successCase := range successCases {
			t.Run(successCase.name, func(t *testing.T) {
				signer, err := common.NewJWSSigner(
					successCase.vm,
					&cryptoMock{})
				require.NoError(t, err)
				require.NotNil(t, signer)
				alg, hasAlg := signer.Headers().Algorithm()
				require.True(t, hasAlg)
				require.Equal(t, successCase.expectedAlg, alg)
			})
		}
	})

	t.Run("Invalid verificationType", func(t *testing.T) {
		_, err := common.NewJWSSigner(
			&models.VerificationMethod{
				ID:   "testKeyID",
				Type: "Invalid",
			},
			&cryptoMock{})
		require.Error(t, err)
	})
}

func TestJWSSigner_Sign(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		signer, err := common.NewJWSSigner(
			&models.VerificationMethod{
				ID:   "testKeyID",
				Type: "Ed25519VerificationKey2018",
			},
			&cryptoMock{Signature: []byte("mock sig")})
		require.NoError(t, err)

		sig, err := signer.Sign([]byte("test data"))

		require.NoError(t, err)
		require.Equal(t, sig, []byte("mock sig"))

		require.Equal(t, signer.Headers()["kid"], "testKeyID")
		require.Equal(t, signer.GetKeyID(), "testKeyID")
		require.Equal(t, signer.Headers()["alg"], "EdDSA")
	})

	t.Run("Failed", func(t *testing.T) {
		s, err := common.NewJWSSigner(
			&models.VerificationMethod{
				ID:   "testKeyID",
				Type: "Ed25519VerificationKey2018",
			},
			&cryptoMock{Err: errors.New("test error")})
		require.NoError(t, err)

		_, err = s.Sign([]byte("test data"))
		require.Error(t, err)
	})
}

type cryptoMock struct {
	Signature []byte
	Err       error
}

func (c *cryptoMock) Sign(msg []byte, keyID string) ([]byte, error) {
	return c.Signature, c.Err
}

// Verify is not yet defined.
func (c *cryptoMock) Verify(signature, msg []byte, keyID string) error {
	return nil
}
