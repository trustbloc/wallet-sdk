/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display_test

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"strconv"
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

	require.Equal(t, 5, credentialDisplay.ClaimsLength())

	checkClaims(t, credentialDisplay)
}

func checkClaims(t *testing.T, credentialDisplay *display.CredentialDisplay) { //nolint:gocyclo // Test file
	t.Helper()

	// Since the claims object in the supported_credentials object from the issuer is a map which effectively gets
	// converted to an array of resolved claims, the order of resolved claims can differ from run-to-run. The code
	// below checks to ensure we have the expected claims in any order.
	type expectedClaim struct {
		RawID     string
		Label     string
		ValueType string
		Value     string
		RawValue  string
		Locale    string
		Pattern   string
		Order     *int
	}

	expectedIDOrder := 0
	expectedGivenNameOrder := 1
	expectedSurnameOrder := 2

	expectedClaimsChecklist := struct {
		Claims []expectedClaim
		Found  []bool
	}{
		Claims: []expectedClaim{
			{
				RawID:     "id",
				Label:     "ID",
				ValueType: "string",
				Value:     "1234",
				RawValue:  "1234",
				Locale:    "en-US",
				Order:     &expectedIDOrder,
			},
			{
				RawID:     "given_name",
				Label:     "Given Name",
				ValueType: "string",
				Value:     "Alice",
				RawValue:  "Alice",
				Locale:    "en-US",
				Order:     &expectedGivenNameOrder,
			},
			{
				RawID:     "surname",
				Label:     "Surname",
				ValueType: "string",
				Value:     "Bowman",
				RawValue:  "Bowman",
				Locale:    "en-US",
				Order:     &expectedSurnameOrder,
			},
			{
				RawID:     "gpa",
				Label:     "GPA",
				ValueType: "number",
				Value:     "4.0",
				RawValue:  "4.0",
				Locale:    "en-US",
			},
			{
				RawID:     "sensitive_id",
				Label:     "Sensitive ID",
				ValueType: "string",
				Value:     "******789",
				RawValue:  "123456789",
				Locale:    "en-US",
				Pattern:   "mask:regex(^.{6})",
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
				claim.Locale() == expectedClaim.Locale &&
				claim.RawID() == expectedClaim.RawID &&
				claim.Pattern() == expectedClaim.Pattern &&
				claim.RawValue() == expectedClaim.RawValue &&
				claimOrderMatches(t, claim, expectedClaim.Order) {
				if expectedClaimsChecklist.Found[j] {
					require.FailNow(t, "duplicate claim found",
						"[Claim ID: %s] [Pattern: %s] [Raw value: %s] [Label: %s] "+
							"[Value Type: %s] [Value: %s] [Order: %s] [Locale: %s]",
						claim.RawID(), claim.Pattern(), claim.RawValue(), claim.Label(),
						claim.ValueType(), claim.Value(), getOrderAsString(t, claim), claim.Locale())
				}

				expectedClaimsChecklist.Found[j] = true

				break
			}

			if j == len(expectedClaimsChecklist.Claims)-1 {
				require.FailNow(t, "received unexpected claim",
					"[Claim ID: %s] [Pattern: %s] [Raw value: %s] [Label: %s] "+
						"[Value Type: %s] [Value: %s] [Order: %s] [Locale: %s]",
					claim.RawID(), claim.Pattern(), claim.RawValue(), claim.Label(),
					claim.ValueType(), claim.Value(), getOrderAsString(t, claim), claim.Locale())
			}
		}
	}

	for i := 0; i < len(expectedClaimsChecklist.Claims); i++ {
		if !expectedClaimsChecklist.Found[i] {
			expectedClaim := expectedClaimsChecklist.Claims[i]
			require.FailNow(t, "claim was expected but wasn't received",
				"[Claim ID: %s] [Pattern: %s] [Raw value: %s] [Label: %s] "+
					"[Value Type: %s] [Value: %s] [Order: %s] [Locale: %s]",
				expectedClaim.RawID, expectedClaim.Pattern, expectedClaim.RawValue, expectedClaim.Label,
				expectedClaim.ValueType, expectedClaim.Value, getOrderIntPointerPointerAsString(expectedClaim.Order),
				expectedClaim.Locale)
		}
	}
}

func claimOrderMatches(t *testing.T, claim *display.Claim, expectedOrder *int) bool {
	t.Helper()

	if claim.HasOrder() {
		if expectedOrder == nil {
			return false
		}

		claimOrder, err := claim.Order()
		require.NoError(t, err)

		if claimOrder != *expectedOrder {
			return false
		}
	} else if expectedOrder != nil {
		return false
	}

	return true
}

func getOrderAsString(t *testing.T, claim *display.Claim) string {
	t.Helper()

	if claim.HasOrder() {
		order, err := claim.Order()
		require.NoError(t, err)

		return strconv.Itoa(order)
	}

	return "none"
}

func getOrderIntPointerPointerAsString(order *int) string {
	if order != nil {
		return strconv.Itoa(*order)
	}

	return "none"
}
