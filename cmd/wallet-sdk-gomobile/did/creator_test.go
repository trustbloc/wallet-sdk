/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did_test

import (
	"crypto/ed25519"
	"errors"
	"testing"

	"github.com/go-jose/go-jose/v3"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk"
	"github.com/hyperledger/aries-framework-go/pkg/doc/util/jwkkid"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/stretchr/testify/require"
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

		didCreator, err := did.NewCreatorWithKeyWriter(localKMS)
		require.NoError(t, err)
		require.NotNil(t, didCreator)
	})
	t.Run("Failure - no KeyWriter specified", func(t *testing.T) {
		didCreator, err := did.NewCreatorWithKeyWriter(nil)
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

			creator, err := did.NewCreatorWithKeyWriter(localKMS)
			require.NoError(t, err)

			didDocResolution, err := creator.Create(did.DIDMethodKey, nil)
			require.NoError(t, err)
			require.NotEmpty(t, didDocResolution)
		})

		t.Run("fail to create key", func(t *testing.T) {
			kw := &mockKeyWriter{
				getKeyErr: errors.New("expected error"),
			}

			creator, err := did.NewCreatorWithKeyWriter(kw)
			require.NoError(t, err)

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

			createDIDOpts := &api.CreateDIDOpts{
				KeyID:            "SomeKeyID",
				VerificationType: did.JSONWebKey2020,
			}

			didDocResolution, err := creator.Create(did.DIDMethodKey, createDIDOpts)
			require.NoError(t, err)
			require.NotEmpty(t, didDocResolution)
		})
		t.Run("Fail to get key", func(t *testing.T) {
			mockKHR := &mockKeyHandleReader{
				errGetKeyHandle: errors.New("test failure"),
			}

			creator, err := did.NewCreatorWithKeyReader(mockKHR)
			require.NoError(t, err)

			createDIDOpts := &api.CreateDIDOpts{
				KeyID:            "SomeKeyID",
				VerificationType: did.Ed25519VerificationKey2018,
			}

			didDocResolution, err := creator.Create(did.DIDMethodKey, createDIDOpts)
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

			createDIDOpts := &api.CreateDIDOpts{
				KeyID:            "SomeKeyID",
				VerificationType: did.Ed25519VerificationKey2018,
			}

			didDocResolution, err := creator.Create(did.DIDMethodKey, createDIDOpts)
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
