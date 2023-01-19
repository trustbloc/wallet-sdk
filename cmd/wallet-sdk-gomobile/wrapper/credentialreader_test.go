/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper_test

import (
	_ "embed"
	"errors"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api/vcparse"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
)

var (
	//go:embed test_data/university_degree.jwt
	universityDegreeVC string

	//go:embed test_data/permanent_resident_card.jwt
	permanentResidentCardVC string
)

func TestCredentialReaderWrapper_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		vc, err := vcparse.Parse(universityDegreeVC, &vcparse.Opts{DisableProofCheck: true})
		require.NoError(t, err)

		reader := wrapper.CredentialReaderWrapper{
			CredentialReader: &readerMock{
				getReturn: vc,
			},
		}

		cred, err := reader.Get("test")
		require.NoError(t, err)
		require.NotNil(t, cred)
	})

	t.Run("Reader error", func(t *testing.T) {
		reader := wrapper.CredentialReaderWrapper{
			CredentialReader: &readerMock{
				err: errors.New("reader error"),
			},
		}

		_, err := reader.Get("test")
		require.Contains(t, err.Error(), "reader error")
	})
}

func TestCredentialReaderWrapper_GetAll(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		prCardVC, err := vcparse.Parse(permanentResidentCardVC, &vcparse.Opts{DisableProofCheck: true})
		require.NoError(t, err)

		uniDegreeVC, err := vcparse.Parse(universityDegreeVC, &vcparse.Opts{DisableProofCheck: true})
		require.NoError(t, err)

		vcArray := api.NewVerifiableCredentialsArray()

		vcArray.Add(prCardVC)
		vcArray.Add(uniDegreeVC)

		reader := wrapper.CredentialReaderWrapper{
			CredentialReader: &readerMock{
				getAllReturn: vcArray,
			},
		}

		creds, err := reader.GetAll()
		require.NoError(t, err)
		require.NotNil(t, creds)
	})

	t.Run("Reader error", func(t *testing.T) {
		reader := wrapper.CredentialReaderWrapper{
			CredentialReader: &readerMock{
				err: errors.New("reader error"),
			},
		}

		_, err := reader.GetAll()
		require.Contains(t, err.Error(), "reader error")
	})
}

type readerMock struct {
	getReturn    *api.VerifiableCredential
	getAllReturn *api.VerifiableCredentialsArray
	err          error
}

func (r *readerMock) Get(id string) (*api.VerifiableCredential, error) {
	return r.getReturn, r.err
}

func (r *readerMock) GetAll() (*api.VerifiableCredentialsArray, error) {
	return r.getAllReturn, r.err
}
