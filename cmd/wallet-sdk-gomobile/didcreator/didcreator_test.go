/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didcreator_test

import (
	"crypto/ed25519"
	"errors"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didcreator"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
)

type mockKeyHandleReader struct {
	getKeyReturn    []byte
	errGetKeyHandle error
}

func (m *mockKeyHandleReader) GetKey(string) ([]byte, error) {
	return m.getKeyReturn, m.errGetKeyHandle
}

func TestNewCreatorWithKeyWriter(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS, err := localkms.NewKMS()
		require.NoError(t, err)

		didCreator, err := didcreator.NewCreatorWithKeyWriter(localKMS)
		require.NoError(t, err)
		require.NotNil(t, didCreator)
	})
	t.Run("Failure - no KeyWriter specified", func(t *testing.T) {
		didCreator, err := didcreator.NewCreatorWithKeyWriter(nil)
		require.EqualError(t, err, "a KeyWriter must be specified")
		require.Nil(t, didCreator)
	})
}

func TestNewCreatorWithKeyReader(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		localKMS, err := localkms.NewKMS()
		require.NoError(t, err)

		didCreator, err := didcreator.NewCreatorWithKeyReader(localKMS)
		require.NoError(t, err)
		require.NotNil(t, didCreator)
	})
	t.Run("Failure - no KeyReader specified", func(t *testing.T) {
		didCreator, err := didcreator.NewCreatorWithKeyReader(nil)
		require.EqualError(t, err, "a KeyReader must be specified")
		require.Nil(t, didCreator)
	})
}

func TestCreator_Create(t *testing.T) {
	t.Run("Using KeyWriter (automatic key generation) - success", func(t *testing.T) {
		localKMS, err := localkms.NewKMS()
		require.NoError(t, err)

		creator, err := didcreator.NewCreatorWithKeyWriter(localKMS)
		require.NoError(t, err)

		createDIDOpts := &api.CreateDIDOpts{}

		didDocResolution, err := creator.Create(didcreator.DIDMethodKey, createDIDOpts)
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

			creator, err := didcreator.NewCreatorWithKeyReader(mockKHR)
			require.NoError(t, err)

			createDIDOpts := &api.CreateDIDOpts{
				KeyID:            "SomeKeyID",
				VerificationType: didcreator.Ed25519VerificationKey2018,
			}

			didDocResolution, err := creator.Create(didcreator.DIDMethodKey, createDIDOpts)
			require.NoError(t, err)
			require.NotEmpty(t, didDocResolution)
		})
		t.Run("Fail to get key handle", func(t *testing.T) {
			mockKHR := &mockKeyHandleReader{
				errGetKeyHandle: errors.New("test failure"),
			}

			creator, err := didcreator.NewCreatorWithKeyReader(mockKHR)
			require.NoError(t, err)

			createDIDOpts := &api.CreateDIDOpts{
				KeyID:            "SomeKeyID",
				VerificationType: didcreator.Ed25519VerificationKey2018,
			}

			didDocResolution, err := creator.Create(didcreator.DIDMethodKey, createDIDOpts)
			require.EqualError(t, err, "failed to get key handle: test failure")
			require.Empty(t, didDocResolution)
		})
	})
}
