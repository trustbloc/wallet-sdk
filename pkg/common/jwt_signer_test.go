/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/common"
)

func TestJWSSigner(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		signer, err := common.NewJWSSigner(
			"testKeyID",
			"Ed25519VerificationKey2018",
			&cryptoMock{Signature: []byte("mock sig")})
		require.NoError(t, err)

		sig, err := signer.Sign([]byte("test data"))

		require.NoError(t, err)
		require.Equal(t, sig, []byte("mock sig"))

		require.Equal(t, signer.Headers()["kid"], "testKeyID")
		require.Equal(t, signer.Headers()["alg"], "EdDSA")
	})

	t.Run("Invalid verificationType", func(t *testing.T) {
		_, err := common.NewJWSSigner(
			"testKeyID",
			"Invalid",
			&cryptoMock{Err: errors.New("test error")})
		require.Error(t, err)
	})

	t.Run("Failed", func(t *testing.T) {
		s, err := common.NewJWSSigner(
			"testKeyID",
			"Ed25519VerificationKey2018",
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
