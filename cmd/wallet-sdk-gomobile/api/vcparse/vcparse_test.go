/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package vcparse_test

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api/vcparse"
)

//go:embed testdata/credential_university_degree.jsonld
var universityDegreeCredential string

func TestParse(t *testing.T) {
	t.Run("Success - default options", func(t *testing.T) {
		universityDegreeVC, err := vcparse.Parse(universityDegreeCredential, nil)
		require.NoError(t, err)
		require.Equal(t, "http://example.edu/credentials/1872", universityDegreeVC.VC.ID)
	})
	t.Run("Success - proof check disabled", func(t *testing.T) {
		parseOpts := vcparse.NewOpts(true, nil)

		universityDegreeVC, err := vcparse.Parse(universityDegreeCredential, parseOpts)
		require.NoError(t, err)
		require.Equal(t, "http://example.edu/credentials/1872", universityDegreeVC.VC.ID)
	})
	t.Run("Failure - blank VC", func(t *testing.T) {
		parseOpts := &vcparse.Opts{DocumentLoader: &documentLoaderMock{}}

		universityDegreeVC, err := vcparse.Parse("", parseOpts)
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
