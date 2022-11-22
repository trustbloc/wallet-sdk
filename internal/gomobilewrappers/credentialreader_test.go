/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gomobilewrappers_test

import (
	_ "embed"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

	"github.com/trustbloc/wallet-sdk/internal/gomobilewrappers"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
)

var (
	//go:embed test_data/university_degree.jwt
	universityDegreeVC string

	//go:embed test_data/permanent_resident_card.jwt
	permanentResidentCardVC string
)

func TestCredentialReaderWrapper_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		reader := gomobilewrappers.CredentialReaderWrapper{
			CredentialReader: &readerMock{
				content: []byte(universityDegreeVC),
			},
			DocumentLoader: testutil.DocumentLoader(t),
		}

		cred, err := reader.Get("test")
		require.NoError(t, err)
		require.NotNil(t, cred)
	})

	t.Run("Reader error", func(t *testing.T) {
		reader := gomobilewrappers.CredentialReaderWrapper{
			CredentialReader: &readerMock{
				err: errors.New("reader error"),
			},
			DocumentLoader: testutil.DocumentLoader(t),
		}

		_, err := reader.Get("test")
		require.Contains(t, err.Error(), "reader error")
	})

	t.Run("Parse credential error", func(t *testing.T) {
		reader := gomobilewrappers.CredentialReaderWrapper{
			CredentialReader: &readerMock{
				content: []byte("[[["),
			},
			DocumentLoader: testutil.DocumentLoader(t),
		}

		_, err := reader.Get("test")
		require.Contains(t, err.Error(), "verifiable credential parse failed")
	})
}

func TestCredentialReaderWrapper_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		reader := gomobilewrappers.CredentialReaderWrapper{
			CredentialReader: &readerMock{
				content: createCredJSONArray(t, []string{permanentResidentCardVC, universityDegreeVC}),
			},
			DocumentLoader: testutil.DocumentLoader(t),
		}

		creds, err := reader.GetAll()
		require.NoError(t, err)
		require.NotNil(t, creds)
	})

	t.Run("Reader error", func(t *testing.T) {
		reader := gomobilewrappers.CredentialReaderWrapper{
			CredentialReader: &readerMock{
				err: errors.New("reader error"),
			},
			DocumentLoader: testutil.DocumentLoader(t),
		}

		_, err := reader.GetAll()
		require.Contains(t, err.Error(), "reader error")
	})

	t.Run("Parse credential error", func(t *testing.T) {
		reader := gomobilewrappers.CredentialReaderWrapper{
			CredentialReader: &readerMock{
				content: []byte("[]"),
			},
			DocumentLoader: testutil.DocumentLoader(t),
		}

		_, err := reader.GetAll()
		require.Contains(t, err.Error(), "unmarshal of credentials array failed")
	})
}

type readerMock struct {
	content []byte
	err     error
}

func (r *readerMock) Get(id string) (*api.JSONObject, error) {
	return &api.JSONObject{Data: r.content}, r.err
}

func (r *readerMock) GetAll() (*api.JSONArray, error) {
	return &api.JSONArray{Data: r.content}, r.err
}

func createCredJSONArray(t *testing.T, creds []string) []byte {
	t.Helper()

	arr, err := json.Marshal(creds)
	require.NoError(t, err)

	return arr
}
