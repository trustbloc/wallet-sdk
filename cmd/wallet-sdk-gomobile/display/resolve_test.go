/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display_test

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api/vcparse"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

var (
	//go:embed testdata/issuer_metadata.json
	sampleIssuerMetadata []byte

	//go:embed testdata/credential_university_degree.jsonld
	credentialUniversityDegree string
)

type mockIssuerServerHandler struct {
	t              *testing.T
	issuerMetadata string
}

func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	_, err := writer.Write([]byte(m.issuerMetadata))
	require.NoError(m.t, err)
}

func TestResolve(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:              t,
			issuerMetadata: string(sampleIssuerMetadata),
		}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		opts := vcparse.NewOpts(true, nil)

		vc, err := vcparse.Parse(credentialUniversityDegree, opts)
		require.NoError(t, err)

		vcs := api.NewVerifiableCredentialsArray()
		vcs.Add(vc)

		resolveOpts := &display.ResolveOpts{
			VCs:       vcs,
			IssuerURI: server.URL,
		}

		t.Run("Without a preferred locale specified", func(t *testing.T) {
			resolvedDisplayData, err := display.Resolve(resolveOpts)
			require.NoError(t, err)
			checkResolvedDisplayData(t, resolvedDisplayData)
		})
		t.Run("With a preferred locale specified", func(t *testing.T) {
			resolveOpts.PreferredLocale = "en-us"

			resolvedDisplayData, err := display.Resolve(resolveOpts)
			require.NoError(t, err)
			checkResolvedDisplayData(t, resolvedDisplayData)
		})
	})
	t.Run("No credentials specified", func(t *testing.T) {
		resolvedDisplayData, err := display.Resolve(&display.ResolveOpts{})
		require.EqualError(t, err, "no credentials specified")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("No issuer URI specified", func(t *testing.T) {
		resolveOpts := &display.ResolveOpts{
			VCs: api.NewVerifiableCredentialsArray(),
		}

		resolvedDisplayData, err := display.Resolve(resolveOpts)
		require.EqualError(t, err, "no issuer URI specified")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("Malformed issuer URI", func(t *testing.T) {
		resolveOpts := &display.ResolveOpts{
			VCs:       api.NewVerifiableCredentialsArray(),
			IssuerURI: "badURL",
		}

		resolvedDisplayData, err := display.Resolve(resolveOpts)
		require.EqualError(t, err,
			`Get "badURL/.well-known/openid-credential-issuer": unsupported protocol scheme ""`)
		require.Nil(t, resolvedDisplayData)
	})
}

func checkResolvedDisplayData(t *testing.T, resolvedDisplayData *display.Data) {
	t.Helper()

	checkIssuerDisplay(t, resolvedDisplayData.IssuerDisplay())

	require.Equal(t, 1, resolvedDisplayData.CredentialDisplaysLength())

	credentialDisplay := resolvedDisplayData.CredentialDisplayAtIndex(0)
	checkCredentialDisplay(t, credentialDisplay)
}

func checkIssuerDisplay(t *testing.T, issuerDisplay *display.IssuerDisplay) {
	t.Helper()

	require.Equal(t, "Example University", issuerDisplay.Name())
	require.Equal(t, "en-US", issuerDisplay.Locale())
}

func checkCredentialDisplay(t *testing.T, credentialDisplay *display.CredentialDisplay) {
	t.Helper()

	credentialOverview := credentialDisplay.Overview()
	require.Equal(t, "University Credential", credentialOverview.Name())
	require.Equal(t, "en-US", credentialOverview.Locale())
	require.Equal(t, "https://exampleuniversity.com/public/logo.png", credentialOverview.Logo().URL())
	require.Equal(t, "a square logo of a university", credentialOverview.Logo().AltText())
	require.Equal(t, "#12107c", credentialOverview.BackgroundColor())
	require.Equal(t, "#FFFFFF", credentialOverview.TextColor())

	require.Equal(t, 4, credentialDisplay.ClaimsLength())

	checkClaims(t, credentialDisplay)
}

func checkClaims(t *testing.T, credentialDisplay *display.CredentialDisplay) { //nolint:gocyclo // Test file
	t.Helper()

	// Since the claims object in the supported_credentials object from the issuer is a map which effectively gets
	// converted to an array of resolved claims, the order of resolved claims can differ from run-to-run. The code
	// below checks to ensure we have the expected claims in any order.
	type expectedClaim struct {
		Label     string
		ValueType string
		Value     string
		Locale    string
	}

	expectedClaimsChecklist := struct {
		Claims []expectedClaim
		Found  []bool
	}{
		Claims: []expectedClaim{
			{
				Label:     "ID",
				ValueType: "string",
				Value:     "1234",
				Locale:    "en-US",
			},
			{
				Label:     "Given Name",
				ValueType: "string",
				Value:     "Alice",
				Locale:    "en-US",
			},
			{
				Label:     "Surname",
				ValueType: "string",
				Value:     "Bowman",
				Locale:    "en-US",
			},
			{
				Label:     "GPA",
				ValueType: "number",
				Value:     "4.0",
				Locale:    "en-US",
			},
		},
	}
	expectedClaimsChecklist.Found = make([]bool, len(expectedClaimsChecklist.Claims))

	for i := 0; i < credentialDisplay.ClaimsLength(); i++ {
		claim := credentialDisplay.ClaimAtIndex(i)

		for j := 0; j < len(expectedClaimsChecklist.Claims); j++ {
			expectedClaim := expectedClaimsChecklist.Claims[j]
			if claim.Label() == expectedClaim.Label &&
				claim.ValueType() == expectedClaim.ValueType &&
				claim.Value() == expectedClaim.Value &&
				claim.Locale() == expectedClaim.Locale {
				if expectedClaimsChecklist.Found[j] {
					require.FailNow(t, "duplicate claim found: ",
						"[Label: %s] [Value Type: %s] [Value: %s] [Locale: %s]",
						claim.Label(), claim.ValueType(), claim.Value(), claim.Locale())
				}

				expectedClaimsChecklist.Found[j] = true

				break
			}

			if j == len(expectedClaimsChecklist.Claims)-1 {
				require.FailNow(t, "received unexpected claim: ",
					"[Label: %s] [Value Type: %s] [Value: %s] [Locale: %s]",
					claim.Label(), claim.ValueType(), claim.Value(), claim.Locale())
			}
		}
	}

	for i := 0; i < len(expectedClaimsChecklist.Claims); i++ {
		if !expectedClaimsChecklist.Found[i] {
			expectedClaim := expectedClaimsChecklist.Claims[i]
			require.FailNow(t, "the following claim was expected but wasn't received: ",
				"[Label: %s] [Value Type: %s] [Value: %s] [Locale: %s]",
				expectedClaim.Label, expectedClaim.ValueType, expectedClaim.Value, expectedClaim.Locale)
		}
	}
}
