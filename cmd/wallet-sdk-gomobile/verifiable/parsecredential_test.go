/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package verifiable_test

import (
	_ "embed"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

//go:embed testdata/credential_university_degree.jsonld
var universityDegreeCredential string

func TestParse(t *testing.T) {
	t.Run("Success - default options", func(t *testing.T) {
		universityDegreeVC, err := verifiable.ParseCredential(universityDegreeCredential, nil)
		require.NoError(t, err)
		require.Equal(t, "http://example.edu/credentials/1872", universityDegreeVC.VC.ID)
	})
	t.Run("Success - proof check disabled", func(t *testing.T) {
		opts := verifiable.NewOpts()
		opts.DisableProofCheck()
		opts.SetHTTPTimeoutNanoseconds(0)

		universityDegreeVC, err := verifiable.ParseCredential(universityDegreeCredential, opts)
		require.NoError(t, err)
		require.Equal(t, "http://example.edu/credentials/1872", universityDegreeVC.VC.ID)
	})
	t.Run("Failure - blank VC", func(t *testing.T) {
		opts := &verifiable.Opts{}
		opts.SetDocumentLoader(&documentLoaderMock{})

		universityDegreeVC, err := verifiable.ParseCredential("", opts)
		require.EqualError(t, err, "decode new credential: embedded proof is not JSON: "+
			"unexpected end of JSON input")
		require.Nil(t, universityDegreeVC)
	})
}

type documentLoaderMock struct {
	LoadResult *api.LDDocument
	LoadErr    error
}

func (d *documentLoaderMock) LoadDocument(u string) (*api.LDDocument, error) {
	return d.LoadResult, d.LoadErr
}
