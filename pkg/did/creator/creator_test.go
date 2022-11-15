/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package creator_test

import (
	"crypto/ed25519"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

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

func (m *mockKeyHandleReader) GetSignAlgorithm(keyID string) (string, error) {
	return "", nil
}

func TestNewCreatorWithKeyWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS()
		require.NoError(t, err)

		didCreator, err := creator.NewCreatorWithKeyWriter(localKMS)
		require.NoError(t, err)
		require.NotNil(t, didCreator)
	})
	t.Run("Failure - no KeyWriter specified", func(t *testing.T) {
		didCreator, err := creator.NewCreatorWithKeyWriter(nil)
		require.EqualError(t, err, "a KeyWriter must be specified")
		require.Nil(t, didCreator)
	})
}

func TestNewCreatorWithKeyReader(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS()
		require.NoError(t, err)

		didCreator, err := creator.NewCreatorWithKeyReader(localKMS)
		require.NoError(t, err)
		require.NotNil(t, didCreator)
	})
	t.Run("Failure - no KeyReader specified", func(t *testing.T) {
		didCreator, err := creator.NewCreatorWithKeyReader(nil)
		require.EqualError(t, err, "a KeyReader must be specified")
		require.Nil(t, didCreator)
	})
}

func TestCreator_Create(t *testing.T) {
	t.Run("Using KeyWriter (automatic key generation) - success", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS()
		require.NoError(t, err)

		didCreator, err := creator.NewCreatorWithKeyWriter(localKMS)
		require.NoError(t, err)

		createDIDOpts := &api.CreateDIDOpts{}

		didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
		require.NoError(t, err)
		require.NotEmpty(t, didDocResolution)
	})
	t.Run("Using KeyReader (caller specified options)", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			key, _, err := ed25519.GenerateKey(nil)
			require.NoError(t, err)

			mockKHR := &mockKeyHandleReader{
				getKeyReturn: key,
			}

			didCreator, err := creator.NewCreatorWithKeyReader(mockKHR)
			require.NoError(t, err)

			createDIDOpts := &api.CreateDIDOpts{
				KeyID:            "SomeKeyID",
				VerificationType: creator.Ed25519VerificationKey2018,
			}

			didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
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
			require.EqualError(t, err, "no verification type specified")
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
			require.EqualError(t, err, "failed to get key handle: test failure")
			require.Empty(t, didDocResolution)
		})
	})
	t.Run("Unsupported DID method", func(t *testing.T) {
		localKMS, err := localkms.NewLocalKMS()
		require.NoError(t, err)

		didCreator, err := creator.NewCreatorWithKeyWriter(localKMS)
		require.NoError(t, err)

		didDocResolution, err := didCreator.Create("NotAValidDIDMethod", nil)
		require.EqualError(t, err, "DID method NotAValidDIDMethod not supported")
		require.Empty(t, didDocResolution)
	})
}
