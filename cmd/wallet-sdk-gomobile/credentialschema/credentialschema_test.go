/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema_test

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/memstorage"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credentialschema"
)

var (
	//go:embed testdata/issuer_metadata.json
	sampleIssuerMetadata []byte

	//go:embed testdata/credential_university_degree.jsonld
	credentialUniversityDegree string
)

type mockIssuerServerHandler struct{}

func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	_, err := writer.Write(sampleIssuerMetadata)
	if err != nil {
		println(err.Error())
	}
}

func TestResolve(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Directly passing in VC", func(t *testing.T) {
			vcs := fmt.Sprintf("[%s]", credentialUniversityDegree)

			credentials := credentialschema.Credentials{
				VCs: &api.JSONArray{Data: []byte(vcs)},
			}

			issuerMetadata := credentialschema.IssuerMetadata{
				Metadata: &api.JSONObject{Data: sampleIssuerMetadata},
			}

			resolvedDisplayData, err := credentialschema.Resolve(&credentials, &issuerMetadata, "en-US")
			require.NoError(t, err)
			require.NotEmpty(t, resolvedDisplayData)
		})
		t.Run("With credential reader", func(t *testing.T) {
			memProvider := memstorage.NewProvider()

			vc := &api.JSONObject{Data: []byte(credentialUniversityDegree)}

			err := memProvider.Add(vc)
			require.NoError(t, err)

			credentials := credentialschema.Credentials{
				Reader: memProvider,
				IDs:    &api.JSONArray{Data: []byte(`["http://example.edu/credentials/1872"]`)},
			}

			issuerMetadata := credentialschema.IssuerMetadata{
				Metadata: &api.JSONObject{Data: sampleIssuerMetadata},
			}

			resolvedDisplayData, err := credentialschema.Resolve(&credentials, &issuerMetadata, "en-US")
			require.NoError(t, err)
			require.NotEmpty(t, resolvedDisplayData)
		})
		t.Run("Using issuer URI", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{}
			server := httptest.NewServer(issuerServerHandler)

			defer server.Close()

			vcs := fmt.Sprintf("[%s]", credentialUniversityDegree)

			credentials := credentialschema.Credentials{
				VCs: &api.JSONArray{Data: []byte(vcs)},
			}

			issuerMetadata := credentialschema.IssuerMetadata{
				IssuerURI: server.URL,
			}

			resolvedDisplayData, err := credentialschema.Resolve(&credentials, &issuerMetadata, "en-US")
			require.NoError(t, err)
			require.NotEmpty(t, resolvedDisplayData)
		})
	})
	t.Run("No credentials specified", func(t *testing.T) {
		t.Run("Error from wallet-sdk-gomobile layer", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(nil, nil, "")
			require.EqualError(t, err, "no credentials specified")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Error from Go SDK layer", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(&credentialschema.Credentials{},
				&credentialschema.IssuerMetadata{}, "")
			require.EqualError(t, err, "no credentials specified")
			require.Nil(t, resolvedDisplayData)
		})
	})
	t.Run("No issuer metadata source specified", func(t *testing.T) {
		resolvedDisplayData, err := credentialschema.Resolve(&credentialschema.Credentials{}, nil, "")
		require.EqualError(t, err, "no issuer metadata source specified")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("VCs are not provided as a JSON array", func(t *testing.T) {
		resolvedDisplayData, err := credentialschema.Resolve(&credentialschema.Credentials{
			VCs: &api.JSONArray{Data: []byte("NotAJSONArray")},
		}, &credentialschema.IssuerMetadata{}, "")
		require.EqualError(t, err, "failed to unmarshal VCs into an array: invalid character 'N' "+
			"looking for beginning of value")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("Failed to parse VC", func(t *testing.T) {
		resolvedDisplayData, err := credentialschema.Resolve(&credentialschema.Credentials{
			VCs: &api.JSONArray{Data: []byte(`[{}]`)},
		}, &credentialschema.IssuerMetadata{}, "")
		require.EqualError(t, err, "failed to parse credential: build new credential: "+
			"fill credential types from raw: credential type of unknown structure")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("Credential IDs are nil", func(t *testing.T) {
		memProvider := memstorage.NewProvider()

		vc := &api.JSONObject{Data: []byte(credentialUniversityDegree)}

		err := memProvider.Add(vc)
		require.NoError(t, err)

		credentials := credentialschema.Credentials{
			Reader: memProvider,
			IDs:    &api.JSONArray{},
		}

		issuerMetadata := credentialschema.IssuerMetadata{
			Metadata: &api.JSONObject{Data: sampleIssuerMetadata},
		}

		resolvedDisplayData, err := credentialschema.Resolve(&credentials, &issuerMetadata, "en-US")
		require.EqualError(t, err, "failed to unmarshal credential IDs into a []string: "+
			"unexpected end of JSON input")
		require.Nil(t, resolvedDisplayData)
	})
}
