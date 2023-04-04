/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential_test

import (
	_ "embed"
	"errors"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
)

func TestCredentialReaderWrapper_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		opts := verifiable.NewOpts()
		opts.DisableProofCheck()

		vc, err := verifiable.ParseCredential(universityDegreeVC, opts)
		require.NoError(t, err)

		reader := credential.ReaderWrapper{
			CredentialReader: &readerMock{
				getReturn: vc,
			},
		}

		cred, err := reader.Get("test")
		require.NoError(t, err)
		require.NotNil(t, cred)
	})

	t.Run("Reader error", func(t *testing.T) {
		reader := credential.ReaderWrapper{
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
		opts := verifiable.NewOpts()
		opts.DisableProofCheck()

		prCardVC, err := verifiable.ParseCredential(string(permanentResidentCardVC), opts)
		require.NoError(t, err)

		uniDegreeVC, err := verifiable.ParseCredential(universityDegreeVC, opts)
		require.NoError(t, err)

		vcArray := verifiable.NewCredentialsArray()

		vcArray.Add(prCardVC)
		vcArray.Add(uniDegreeVC)

		reader := credential.ReaderWrapper{
			CredentialReader: &readerMock{
				getAllReturn: vcArray,
			},
		}

		creds, err := reader.GetAll()
		require.NoError(t, err)
		require.NotNil(t, creds)
	})

	t.Run("Reader error", func(t *testing.T) {
		reader := credential.ReaderWrapper{
			CredentialReader: &readerMock{
				err: errors.New("reader error"),
			},
		}

		_, err := reader.GetAll()
		require.Contains(t, err.Error(), "reader error")
	})
}

type readerMock struct {
	getReturn    *verifiable.Credential
	getAllReturn *verifiable.CredentialsArray
	err          error
}

func (r *readerMock) Get(string) (*verifiable.Credential, error) {
	return r.getReturn, r.err
}

func (r *readerMock) GetAll() (*verifiable.CredentialsArray, error) {
	return r.getAllReturn, r.err
}
