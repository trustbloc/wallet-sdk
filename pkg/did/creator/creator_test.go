/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package creator_test

import (
	"crypto/ed25519"
	"errors"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/kms"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/did/creator"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
)

type mockKeyHandleReader struct {
	getKeyReturn    []byte
	errGetKeyHandle error
}

func (m *mockKeyHandleReader) ExportPubKey(string) ([]byte, error) {
	return m.getKeyReturn, m.errGetKeyHandle
}

func TestNewCreator(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
		require.NoError(t, err)

		didCreator, err := creator.NewCreator(localKMS, nil)
		require.NoError(t, err)
		require.NotNil(t, didCreator)
	})
	t.Run("Failure - no KeyWriter specified", func(t *testing.T) {
		didCreator, err := creator.NewCreator(nil, nil)
		testutil.RequireErrorContains(t, err, "a KeyWriter must be specified")
		require.Nil(t, didCreator)
	})
}

func TestNewCreatorWithKeyWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
		require.NoError(t, err)

		didCreator, err := creator.NewCreatorWithKeyWriter(localKMS)
		require.NoError(t, err)
		require.NotNil(t, didCreator)
	})
	t.Run("Failure - no KeyWriter specified", func(t *testing.T) {
		didCreator, err := creator.NewCreatorWithKeyWriter(nil)
		testutil.RequireErrorContains(t, err, "a KeyWriter must be specified")
		require.Nil(t, didCreator)
	})
}

func TestNewCreatorWithKeyReader(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
		require.NoError(t, err)

		didCreator, err := creator.NewCreatorWithKeyReader(localKMS)
		require.NoError(t, err)
		require.NotNil(t, didCreator)
	})
	t.Run("Failure - no KeyReader specified", func(t *testing.T) {
		didCreator, err := creator.NewCreatorWithKeyReader(nil)
		testutil.RequireErrorContains(t, err, "a KeyReader must be specified")
		require.Nil(t, didCreator)
	})
}

func TestCreator_Create(t *testing.T) {
	t.Run("Using KeyWriter (automatic key generation) - success", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
		require.NoError(t, err)

		didCreator, err := creator.NewCreatorWithKeyWriter(localKMS)
		require.NoError(t, err)

		createDIDOpts := &api.CreateDIDOpts{}

		didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
		require.NoError(t, err)
		require.NotEmpty(t, didDocResolution)

		didDocResolution, err = didCreator.Create(creator.DIDMethodIon, createDIDOpts)
		require.NoError(t, err)
		require.NotEmpty(t, didDocResolution)

		didDocResolution, err = didCreator.Create(creator.DIDMethodJWK, createDIDOpts)
		require.NoError(t, err)
		require.NotEmpty(t, didDocResolution)
	})
	t.Run("Using KeyWriter (automatic key generation) - failure", func(t *testing.T) {
		expectErr := errors.New("expected error")

		badKMS := mockKeyWriter(func(keyType kms.KeyType) (string, []byte, error) {
			return "", nil, expectErr
		})

		didCreator, err := creator.NewCreatorWithKeyWriter(badKMS)
		require.NoError(t, err)

		createDIDOpts := &api.CreateDIDOpts{}

		didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
		testutil.RequireErrorContains(t, err, "CREATE_DID_KEY_FAILED(DID1-0000):expected error")
		require.Empty(t, didDocResolution)

		didDocResolution, err = didCreator.Create(creator.DIDMethodIon, createDIDOpts)
		testutil.RequireErrorContains(t, err, "CREATE_DID_ION_FAILED(DID1-0001):expected error")
		require.Empty(t, didDocResolution)

		didDocResolution, err = didCreator.Create(creator.DIDMethodJWK, createDIDOpts)
		testutil.RequireErrorContains(t, err, "CREATE_DID_JWK_FAILED(DID1-0002):expected error")
		require.Empty(t, didDocResolution)
	})
	t.Run("Using KeyReader (caller specified options)", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			key, _, err := ed25519.GenerateKey(nil)
			require.NoError(t, err)

			mockKHR := &mockKeyHandleReader{
				getKeyReturn: key,
			}

			localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
			require.NoError(t, err)

			didCreator, err := creator.NewCreator(localKMS, mockKHR)
			require.NoError(t, err)

			createDIDOpts := &api.CreateDIDOpts{
				KeyID:            "SomeKeyID",
				VerificationType: creator.Ed25519VerificationKey2018,
				KeyType:          kms.ED25519Type,
			}

			didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
			require.NoError(t, err)
			require.NotEmpty(t, didDocResolution)

			didDocResolution, err = didCreator.Create(creator.DIDMethodIon, createDIDOpts)
			require.NoError(t, err)
			require.NotEmpty(t, didDocResolution)
		})
		t.Run("No verification type specified", func(t *testing.T) {
			mockKHR := &mockKeyHandleReader{}

			didCreator, err := creator.NewCreatorWithKeyReader(mockKHR)
			require.NoError(t, err)

			createDIDOpts := &api.CreateDIDOpts{
				KeyID: "SomeKeyID",
			}

			didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
			testutil.RequireErrorContains(t, err, "no verification type specified")
			require.Empty(t, didDocResolution)

			didDocResolution, err = didCreator.Create(creator.DIDMethodIon, createDIDOpts)
			testutil.RequireErrorContains(t, err, "no verification type specified")
			require.Empty(t, didDocResolution)
		})
		t.Run("Fail to get key handle", func(t *testing.T) {
			mockKHR := &mockKeyHandleReader{
				errGetKeyHandle: errors.New("test failure"),
			}

			didCreator, err := creator.NewCreatorWithKeyReader(mockKHR)
			require.NoError(t, err)

			createDIDOpts := &api.CreateDIDOpts{
				KeyID:            "SomeKeyID",
				VerificationType: creator.Ed25519VerificationKey2018,
			}

			didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
			testutil.RequireErrorContains(t, err, "failed to get key handle: test failure")
			require.Empty(t, didDocResolution)

			didDocResolution, err = didCreator.Create(creator.DIDMethodIon, createDIDOpts)
			testutil.RequireErrorContains(t, err, "failed to get key handle: test failure")
			require.Empty(t, didDocResolution)
		})
		t.Run("Fail to create jwk", func(t *testing.T) {
			mockKHR := &mockKeyHandleReader{
				getKeyReturn: nil,
			}

			didCreator, err := creator.NewCreatorWithKeyReader(mockKHR)
			require.NoError(t, err)

			createDIDOpts := &api.CreateDIDOpts{
				KeyType:          kms.ED25519Type,
				KeyID:            "SomeKeyID",
				VerificationType: creator.JSONWebKey2020,
			}

			didDocResolution, err := didCreator.Create(creator.DIDMethodIon, createDIDOpts)
			require.Error(t, err)
			require.Contains(t, err.Error(), "failed to create JWK from public key")
			require.Empty(t, didDocResolution)
		})
	})
	t.Run("Unsupported DID method", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS(&localkms.Config{})
		require.NoError(t, err)

		didCreator, err := creator.NewCreatorWithKeyWriter(localKMS)
		require.NoError(t, err)

		didDocResolution, err := didCreator.Create("NotAValidDIDMethod", nil)
		testutil.RequireErrorContains(t, err, "DID method NotAValidDIDMethod not supported")
		require.Empty(t, didDocResolution)
	})
}

type mockKeyWriter func(keyType kms.KeyType) (string, []byte, error)

func (kw mockKeyWriter) Create(keyType kms.KeyType) (string, []byte, error) {
	return kw(keyType)
}
