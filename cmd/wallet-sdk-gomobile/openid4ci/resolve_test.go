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

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api/vcparse"

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
			checkResolvedDisplayData(t, resolvedDisplayData)
		})
		t.Run("With credential reader", func(t *testing.T) {
			memProvider := credential.NewInMemoryDB()

			vc, err := vcparse.Parse(credentialUniversityDegree, nil)
			require.NoError(t, err)

			err = memProvider.Add(vc)
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
			checkResolvedDisplayData(t, resolvedDisplayData)
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
			checkResolvedDisplayData(t, resolvedDisplayData)
		})
	})
	t.Run("No credentials specified", func(t *testing.T) {
		t.Run("Category from wallet-sdk-gomobile layer", func(t *testing.T) {
			resolvedDisplayData, err := openid4ci.ResolveDisplay(nil, nil, "")
			require.EqualError(t, err, "no credentials specified")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Category from Go SDK layer", func(t *testing.T) {
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

		vc, err := vcparse.Parse(credentialUniversityDegree, nil)
		require.NoError(t, err)

		err = memProvider.Add(vc)
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

func checkResolvedDisplayData(t *testing.T, resolvedDisplayData *openid4ci.DisplayData) {
	t.Helper()

	checkIssuerDisplay(t, resolvedDisplayData.IssuerDisplay())

	require.Equal(t, 1, resolvedDisplayData.CredentialDisplaysLength())

	credentialDisplay := resolvedDisplayData.CredentialDisplayAtIndex(0)
	checkCredentialDisplay(t, credentialDisplay)
}

func checkIssuerDisplay(t *testing.T, issuerDisplay *openid4ci.IssuerDisplay) {
	t.Helper()

	require.Equal(t, "Example University", issuerDisplay.Name())
	require.Equal(t, "en-US", issuerDisplay.Locale())
}

func checkCredentialDisplay(t *testing.T, credentialDisplay *openid4ci.CredentialDisplay) {
	t.Helper()

	credentialOverview := credentialDisplay.Overview()
	require.Equal(t, "University Credential", credentialOverview.Name())
	require.Equal(t, "en-US", credentialOverview.Locale())
	require.Equal(t, "https://exampleuniversity.com/public/logo.png", credentialOverview.Logo().URL())
	require.Empty(t, credentialOverview.Logo().AltText())
	require.Equal(t, "#12107c", credentialOverview.BackgroundColor())
	require.Equal(t, "#FFFFFF", credentialOverview.TextColor())

	require.Equal(t, 4, credentialDisplay.ClaimsLength())

	// Since the claims object in the supported_credentials object from the issuer is a map which effectively gets
	// converted to an array of resolved claims, the order of resolved claims can differ from run-to-run. The code
	// below checks to ensure we have the expected claims in any order.
	type expectedClaim struct {
		Label  string
		Value  string
		Locale string
	}

	expectedClaimsChecklist := struct {
		Claims []expectedClaim
		Found  []bool
	}{
		Claims: []expectedClaim{
			{
				Label:  "ID",
				Value:  "1234",
				Locale: "en-US",
			},
			{
				Label:  "Given Name",
				Value:  "Alice",
				Locale: "en-US",
			},
			{
				Label:  "Surname",
				Value:  "Bowman",
				Locale: "en-US",
			},
			{
				Label:  "GPA",
				Value:  "4.0",
				Locale: "en-US",
			},
		},
	}
	expectedClaimsChecklist.Found = make([]bool, len(expectedClaimsChecklist.Claims))

	for i := 0; i < credentialDisplay.ClaimsLength(); i++ {
		claim := credentialDisplay.ClaimAtIndex(i)

		for j := 0; j < len(expectedClaimsChecklist.Claims); j++ {
			expectedClaim := expectedClaimsChecklist.Claims[j]
			if claim.Label() == expectedClaim.Label &&
				claim.Value() == expectedClaim.Value &&
				claim.Locale() == expectedClaim.Locale {
				if expectedClaimsChecklist.Found[j] {
					require.FailNow(t, "duplicate claim found: ",
						"[Label: %s] [Value: %s] [Locale: %s]",
						claim.Label(), claim.Value(), claim.Locale())
				}

				expectedClaimsChecklist.Found[j] = true

				break
			}

			if j == len(expectedClaimsChecklist.Claims)-1 {
				require.FailNow(t, "received unexpected claim: ",
					"[Label: %s] [Value: %s] [Locale: %s]",
					claim.Label(), claim.Value(), claim.Locale())
			}
		}
	}

	for i := 0; i < len(expectedClaimsChecklist.Claims); i++ {
		if !expectedClaimsChecklist.Found[i] {
			expectedClaim := expectedClaimsChecklist.Claims[i]
			require.FailNow(t, "the following claim was expected but wasn't received: ",
				"[Label: %s] [Value: %s] [Locale: %s]",
				expectedClaim.Label, expectedClaim.Value, expectedClaim.Locale)
		}
	}
}
