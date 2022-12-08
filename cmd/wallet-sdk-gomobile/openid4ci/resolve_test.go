/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	_ "embed"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

var (
	//go:embed testdata/issuer_metadata.json
	sampleIssuerMetadata []byte

	//go:embed testdata/credential_university_degree.jsonld
	credentialUniversityDegree string
)

func TestResolve(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Directly passing in VC", func(t *testing.T) {
			vcs := fmt.Sprintf("[%s]", credentialUniversityDegree)

			credentials := openid4ci.Credentials{
				VCs: &api.JSONArray{Data: []byte(vcs)},
			}

			issuerMetadata := openid4ci.IssuerMetadata{
				Metadata: &api.JSONObject{Data: sampleIssuerMetadata},
			}

			resolvedDisplayData, err := openid4ci.ResolveDisplay(&credentials, &issuerMetadata, "en-US")
			require.NoError(t, err)
			require.NotEmpty(t, resolvedDisplayData)
		})
		t.Run("With credential reader", func(t *testing.T) {
			memProvider := credential.NewInMemoryDB()

			vc := &api.JSONObject{Data: []byte(credentialUniversityDegree)}

			err := memProvider.Add(vc)
			require.NoError(t, err)

			credentials := openid4ci.Credentials{
				Reader: memProvider,
				IDs:    &api.JSONArray{Data: []byte(`["http://example.edu/credentials/1872"]`)},
			}

			issuerMetadata := openid4ci.IssuerMetadata{
				Metadata: &api.JSONObject{Data: sampleIssuerMetadata},
			}

			resolvedDisplayData, err := openid4ci.ResolveDisplay(&credentials, &issuerMetadata, "en-US")
			require.NoError(t, err)
			require.NotEmpty(t, resolvedDisplayData)
		})
		t.Run("Using issuer URI", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: string(sampleIssuerMetadata)}
			server := httptest.NewServer(issuerServerHandler)

			defer server.Close()

			vcs := fmt.Sprintf("[%s]", credentialUniversityDegree)

			credentials := openid4ci.Credentials{
				VCs: &api.JSONArray{Data: []byte(vcs)},
			}

			issuerMetadata := openid4ci.IssuerMetadata{
				IssuerURI: server.URL,
			}

			resolvedDisplayData, err := openid4ci.ResolveDisplay(&credentials, &issuerMetadata, "en-US")
			require.NoError(t, err)
			require.NotEmpty(t, resolvedDisplayData)
		})
	})
	t.Run("No credentials specified", func(t *testing.T) {
		t.Run("Error from wallet-sdk-gomobile layer", func(t *testing.T) {
			resolvedDisplayData, err := openid4ci.ResolveDisplay(nil, nil, "")
			require.EqualError(t, err, "no credentials specified")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Error from Go SDK layer", func(t *testing.T) {
			resolvedDisplayData, err := openid4ci.ResolveDisplay(&openid4ci.Credentials{},
				&openid4ci.IssuerMetadata{}, "")
			require.EqualError(t, err, "no credentials specified")
			require.Nil(t, resolvedDisplayData)
		})
	})
	t.Run("No issuer metadata source specified", func(t *testing.T) {
		resolvedDisplayData, err := openid4ci.ResolveDisplay(&openid4ci.Credentials{}, nil, "")
		require.EqualError(t, err, "no issuer metadata source specified")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("VCs are not provided as a JSON array", func(t *testing.T) {
		resolvedDisplayData, err := openid4ci.ResolveDisplay(&openid4ci.Credentials{
			VCs: &api.JSONArray{Data: []byte("NotAJSONArray")},
		}, &openid4ci.IssuerMetadata{}, "")
		require.EqualError(t, err, "failed to unmarshal VCs into an array: invalid character 'N' "+
			"looking for beginning of value")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("Failed to parse VC", func(t *testing.T) {
		resolvedDisplayData, err := openid4ci.ResolveDisplay(&openid4ci.Credentials{
			VCs: &api.JSONArray{Data: []byte(`[{}]`)},
		}, &openid4ci.IssuerMetadata{}, "")
		require.EqualError(t, err, "failed to parse credential: build new credential: "+
			"fill credential types from raw: credential type of unknown structure")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("Credential IDs are nil", func(t *testing.T) {
		memProvider := credential.NewInMemoryDB()

		vc := &api.JSONObject{Data: []byte(credentialUniversityDegree)}

		err := memProvider.Add(vc)
		require.NoError(t, err)

		credentials := openid4ci.Credentials{
			Reader: memProvider,
			IDs:    &api.JSONArray{},
		}

		issuerMetadata := openid4ci.IssuerMetadata{
			Metadata: &api.JSONObject{Data: sampleIssuerMetadata},
		}

		resolvedDisplayData, err := openid4ci.ResolveDisplay(&credentials, &issuerMetadata, "en-US")
		require.EqualError(t, err, "failed to unmarshal credential IDs into a []string: "+
			"unexpected end of JSON input")
		require.Nil(t, resolvedDisplayData)
	})
}
