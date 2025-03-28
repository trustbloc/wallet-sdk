/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package verifiable_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
)

//go:embed testdata/credential_university_degree.jsonld
var universityDegreeCredential string

//go:embed testdata/credential_university_degree_without_name.jsonld
var universityDegreeCredentialWithoutName string

func TestParse(t *testing.T) {
	t.Run("Success - default options", func(t *testing.T) {
		opts := verifiable.NewOpts()
		opts.DisableProofCheck()

		universityDegreeVC, err := verifiable.ParseCredential(universityDegreeCredential, opts)
		require.NoError(t, err)
		require.Equal(t, "http://example.edu/credentials/1872", universityDegreeVC.VC.Contents().ID)
	})
	t.Run("Success - proof check disabled", func(t *testing.T) {
		opts := verifiable.NewOpts()
		opts.DisableProofCheck()
		opts.SetHTTPTimeoutNanoseconds(0)

		universityDegreeVC, err := verifiable.ParseCredential(universityDegreeCredential, opts)
		require.NoError(t, err)
		require.Equal(t, "http://example.edu/credentials/1872", universityDegreeVC.VC.Contents().ID)
	})
	t.Run("Success - additional headers", func(t *testing.T) {
		opts := verifiable.NewOpts()
		opts.DisableProofCheck()
		opts.SetHTTPTimeoutNanoseconds(0)
		opts.AddHeader(api.NewHeader("request_id", "123"))

		universityDegreeVC, err := verifiable.ParseCredential(universityDegreeCredential, opts)
		require.NoError(t, err)
		require.Equal(t, "http://example.edu/credentials/1872", universityDegreeVC.VC.Contents().ID)
	})
	t.Run("Success - skip client verification", func(t *testing.T) {
		opts := verifiable.NewOpts()
		opts.DisableProofCheck()
		opts.DisableHTTPClientTLSVerify()

		universityDegreeVC, err := verifiable.ParseCredential(universityDegreeCredential, opts)
		require.NoError(t, err)
		require.Equal(t, "http://example.edu/credentials/1872", universityDegreeVC.VC.Contents().ID)
	})
	t.Run("Failure - blank VC", func(t *testing.T) {
		opts := &verifiable.Opts{}
		opts.SetDocumentLoader(&documentLoaderMock{})

		universityDegreeVC, err := verifiable.ParseCredential("", opts)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unmarshal cbor credential: EOF")
		require.Nil(t, universityDegreeVC)
	})
}

type documentLoaderMock struct {
	LoadResult *api.LDDocument
	LoadErr    error
}

func (d *documentLoaderMock) LoadDocument(string) (*api.LDDocument, error) {
	return d.LoadResult, d.LoadErr
}
