/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did_test

import (
	"crypto/ed25519"
	"errors"
	"testing"

	"github.com/go-jose/go-jose/v3"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-crypto-go/doc/jose/jwk"
	"github.com/trustbloc/kms-crypto-go/doc/util/jwkkid"
	"github.com/trustbloc/kms-crypto-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
)

type mockKeyHandleReader struct {
	getKeyReturn    *jwk.JWK
	errGetKeyHandle error
}

func (m *mockKeyHandleReader) ExportPubKey(string) (*api.JSONWebKey, error) {
	return &api.JSONWebKey{JWK: m.getKeyReturn}, m.errGetKeyHandle
}

type mockKeyWriter struct {
	getKeyVal *api.JSONWebKey
	getKeyErr error
}

func (m *mockKeyWriter) Create(string) (*api.JSONWebKey, error) {
	return m.getKeyVal, m.getKeyErr
}

func TestNewCreatorWithKeyWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS := createTestKMS(t)

		didCreator, err := did.NewCreator(localKMS)
		require.NoError(t, err)
		require.NotNil(t, didCreator)
	})
	t.Run("Failure - no KeyWriter specified", func(t *testing.T) {
		didCreator, err := did.NewCreator(nil)
		require.EqualError(t, err, "a KeyWriter must be specified")
		require.Nil(t, didCreator)
	})
}

func TestNewCreatorWithKeyReader(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS := createTestKMS(t)

		didCreator, err := did.NewCreatorWithKeyReader(localKMS)
		require.NoError(t, err)
		require.NotNil(t, didCreator)
	})
	t.Run("Failure - no KeyReader specified", func(t *testing.T) {
		didCreator, err := did.NewCreatorWithKeyReader(nil)
		require.EqualError(t, err, "a KeyReader must be specified")
		require.Nil(t, didCreator)
	})
}

func TestCreator_Create(t *testing.T) {
	t.Run("Using KeyWriter (automatic key generation)", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			localKMS := createTestKMS(t)

			creator, err := did.NewCreator(localKMS)
			require.NoError(t, err)

			didDocResolution, err := creator.Create(did.DIDMethodKey, nil)
			require.NoError(t, err)
			require.NotEmpty(t, didDocResolution)
		})

		t.Run("fail to create key", func(t *testing.T) {
			kw := &mockKeyWriter{
				getKeyErr: errors.New("expected error"),
			}

			creator, err := did.NewCreator(kw)
			require.NoError(t, err)

			opts := did.NewCreateOpts()
			opts.SetKeyType("")
			// Calling SetMetricsLogger here to increase code coverage - actual functionality tests are in the
			// integration tests.
			opts.SetMetricsLogger(nil)

			didDocResolution, err := creator.Create(did.DIDMethodKey, nil)
			requireErrorContains(t, err, "CREATE_DID_KEY_FAILED")
			require.Empty(t, didDocResolution)
		})
	})

	t.Run("Using KeyReader (caller specified options)", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			key, _, err := ed25519.GenerateKey(nil)
			require.NoError(t, err)

			pkJWK, err := jwkkid.BuildJWK(key, kms.ED25519Type)
			require.NoError(t, err)

			mockKHR := &mockKeyHandleReader{
				getKeyReturn: pkJWK,
			}

			creator, err := did.NewCreatorWithKeyReader(mockKHR)
			require.NoError(t, err)

			opts := did.NewCreateOpts()
			opts.SetVerificationType(did.JSONWebKey2020)

			didDocResolution, err := creator.Create(did.DIDMethodKey, "SomeKeyID", opts)
			require.NoError(t, err)
			require.NotEmpty(t, didDocResolution)
		})
		t.Run("Fail to get key", func(t *testing.T) {
			mockKHR := &mockKeyHandleReader{
				errGetKeyHandle: errors.New("test failure"),
			}

			creator, err := did.NewCreatorWithKeyReader(mockKHR)
			require.NoError(t, err)

			opts := did.NewCreateOpts()
			opts.SetVerificationType(did.Ed25519VerificationKey2018)

			didDocResolution, err := creator.Create(did.DIDMethodKey, "SomeKeyID", opts)
			requireErrorContains(t, err, "CREATE_DID_KEY_FAILED")
			require.Empty(t, didDocResolution)
		})

		t.Run("key contains garbage data", func(t *testing.T) {
			mockKHR := &mockKeyHandleReader{
				getKeyReturn: &jwk.JWK{JSONWebKey: jose.JSONWebKey{
					Algorithm: "invalid",
					Key:       new(chan int),
				}},
			}

			creator, err := did.NewCreatorWithKeyReader(mockKHR)
			require.NoError(t, err)

			opts := did.NewCreateOpts()
			opts.SetVerificationType(did.Ed25519VerificationKey2018)

			didDocResolution, err := creator.Create(did.DIDMethodKey, "SomeKeyID", opts)
			requireErrorContains(t, err, "CREATE_DID_KEY_FAILED")
			require.Empty(t, didDocResolution)
		})
	})
}

func createTestKMS(t *testing.T) *localkms.KMS {
	t.Helper()

	kmsStore := localkms.NewMemKMSStore()

	localKMS, err := localkms.NewKMS(kmsStore)
	require.NoError(t, err)

	return localKMS
}
