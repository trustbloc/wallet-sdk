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
)

const ed25519VerificationKey2018 = "Ed25519VerificationKey2018"

type mockKeyHandleReader struct {
	getKeyHandleReturn *api.KeyHandle
	errGetKeyHandle    error
}

func (m *mockKeyHandleReader) GetKeyHandle(string) (*api.KeyHandle, error) {
	return m.getKeyHandleReturn, m.errGetKeyHandle
}

func (m *mockKeyHandleReader) Export(string) ([]byte, error) {
	return nil, nil
}

func TestDIDCreator_Create(t *testing.T) {
	t.Run("Without a key ID specified (key gets generated for caller)", func(t *testing.T) {
		didCreator := creator.NewDIDCreator(nil)

		createDIDOpts := &api.CreateDIDOpts{}

		didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
		require.NoError(t, err)
		require.NotEmpty(t, didDocResolution)
	})
	t.Run("With a key ID specified", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			key, _, err := ed25519.GenerateKey(nil)
			require.NoError(t, err)

			mockKHR := &mockKeyHandleReader{
				getKeyHandleReturn: &api.KeyHandle{
					Key:     key,
					KeyType: ed25519VerificationKey2018,
				},
			}

			didCreator := creator.NewDIDCreator(mockKHR)

			createDIDOpts := &api.CreateDIDOpts{
				KeyID: "SomeKeyID",
			}

			didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
			require.NoError(t, err)
			require.NotEmpty(t, didDocResolution)
		})
		t.Run("Key handle reader not set up", func(t *testing.T) {
			didCreator := creator.NewDIDCreator(nil)

			createDIDOpts := &api.CreateDIDOpts{
				KeyID: "SomeKeyID",
			}

			didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
			require.EqualError(t, err, "key ID specified but no key handle reader set up")
			require.Empty(t, didDocResolution)
		})
		t.Run("Fail to get key handle", func(t *testing.T) {
			mockKHR := &mockKeyHandleReader{
				errGetKeyHandle: errors.New("test failure"),
			}

			didCreator := creator.NewDIDCreator(mockKHR)

			createDIDOpts := &api.CreateDIDOpts{
				KeyID: "SomeKeyID",
			}

			didDocResolution, err := didCreator.Create(creator.DIDMethodKey, createDIDOpts)
			require.EqualError(t, err, "failed to get key handle: test failure")
			require.Empty(t, didDocResolution)
		})
	})
}
