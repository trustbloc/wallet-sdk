/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package display_test

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

const (
	sensitiveIDLabel       = "Sensitive ID"
	reallySensitiveIDLabel = "Really Sensitive ID"
)

var (
	//go:embed testdata/issuer_metadata.json
	sampleIssuerMetadata []byte

	//go:embed testdata/credential_university_degree.jsonld
	credentialUniversityDegree string

	//go:embed testdata/university_degree_resolved_data.json
	universityDegreeResolvedData string
)

type mockIssuerServerHandler struct {
	t              *testing.T
	issuerMetadata string
	headersToCheck *api.Headers
}

func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if m.headersToCheck != nil {
		for _, headerToCheck := range m.headersToCheck.GetAll() {
			// Note: for these tests, we're assuming that there aren't multiple values under a single name/key.
			value := request.Header.Get(headerToCheck.Name)
			assert.Equal(m.t, headerToCheck.Value, value)
		}
	}

	_, err := writer.Write([]byte(m.issuerMetadata))
	assert.NoError(m.t, err)
}

func TestResolve(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:              t,
			issuerMetadata: string(sampleIssuerMetadata),
		}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		parseVCOptionalArgs := verifiable.NewOpts()
		parseVCOptionalArgs.DisableProofCheck()

		vc, err := verifiable.ParseCredential(credentialUniversityDegree, parseVCOptionalArgs)
		require.NoError(t, err)

		vcs := verifiable.NewCredentialsArray()
		vcs.Add(vc)

		t.Run("With custom masking string", func(t *testing.T) {
			t.Run(`"*"`, func(t *testing.T) {
				opts := display.NewOpts().SetMaskingString("*")

				resolvedDisplayData, err := display.Resolve(vcs, server.URL, opts)
				require.NoError(t, err)

				credentialDisplay := resolvedDisplayData.CredentialDisplayAtIndex(0)

				for i := range credentialDisplay.ClaimsLength() {
					claim := credentialDisplay.ClaimAtIndex(i)

					if claim.Label() == sensitiveIDLabel {
						require.Equal(t, "*****6789", claim.Value())
					} else if claim.Label() == reallySensitiveIDLabel {
						require.Equal(t, "*******", claim.Value())
					}
				}
			})
			t.Run(`"+++"`, func(t *testing.T) {
				opts := display.NewOpts().SetMaskingString("+++")

				resolvedDisplayData, err := display.Resolve(vcs, server.URL, opts)
				require.NoError(t, err)

				credentialDisplay := resolvedDisplayData.CredentialDisplayAtIndex(0)

				for i := range credentialDisplay.ClaimsLength() {
					claim := credentialDisplay.ClaimAtIndex(i)

					if claim.Label() == sensitiveIDLabel {
						require.Equal(t, "+++++++++++++++6789", claim.Value())
					} else if claim.Label() == reallySensitiveIDLabel {
						require.Equal(t, "+++++++++++++++++++++", claim.Value())
					}
				}
			})
			t.Run(`"" (empty string`, func(t *testing.T) {
				opts := display.NewOpts().SetMaskingString("")

				resolvedDisplayData, err := display.Resolve(vcs, server.URL, opts)
				require.NoError(t, err)

				credentialDisplay := resolvedDisplayData.CredentialDisplayAtIndex(0)

				for i := range credentialDisplay.ClaimsLength() {
					claim := credentialDisplay.ClaimAtIndex(i)

					if claim.Label() == sensitiveIDLabel {
						require.Equal(t, "6789", claim.Value())
					} else if claim.Label() == reallySensitiveIDLabel {
						require.Equal(t, "", claim.Value())
					}
				}
			})
		})
		t.Run("Without additional headers", func(t *testing.T) {
			t.Run("Without a preferred locale specified", func(t *testing.T) {
				opts := display.NewOpts().SetHTTPTimeoutNanoseconds(0)
				opts.DisableHTTPClientTLSVerify()

				resolvedDisplayData, err := display.Resolve(vcs, server.URL, opts)
				require.NoError(t, err)
				checkResolvedDisplayData(t, resolvedDisplayData)
			})
			t.Run("With a preferred locale specified", func(t *testing.T) {
				opts := display.NewOpts()
				opts.SetPreferredLocale("en-us")

				resolvedDisplayData, err := display.Resolve(vcs, server.URL, opts)
				require.NoError(t, err)
				checkResolvedDisplayData(t, resolvedDisplayData)
			})
		})
		t.Run("With DID resolver", func(t *testing.T) {
			resolver, err := did.NewResolver(nil)
			require.NoError(t, err)

			opts := display.NewOpts()
			opts.SetDIDResolver(resolver)

			resolvedDisplayData, err := display.Resolve(vcs, server.URL, opts)
			require.NoError(t, err)
			checkResolvedDisplayData(t, resolvedDisplayData)
		})
		t.Run("With skip non claim data", func(t *testing.T) {
			opts := display.NewOpts()
			opts.SkipNonClaimData()

			resolvedDisplayData, err := display.Resolve(vcs, server.URL, opts)
			require.NoError(t, err)
			checkResolvedDisplayData(t, resolvedDisplayData)
		})
		t.Run("With additional headers", func(t *testing.T) {
			additionalHeaders := api.NewHeaders()
			additionalHeaders.Add(api.NewHeader("header-name-1", "header-value-1"))
			additionalHeaders.Add(api.NewHeader("header-name-2", "header-value-2"))

			opts := display.NewOpts()
			opts.AddHeaders(additionalHeaders)

			issuerServerHandler.headersToCheck = additionalHeaders

			resolvedDisplayData, err := display.Resolve(vcs, server.URL, opts)
			require.NoError(t, err)
			checkResolvedDisplayData(t, resolvedDisplayData)
		})
	})

	t.Run("No credentials specified", func(t *testing.T) {
		resolvedDisplayData, err := display.Resolve(nil, "", nil)
		require.EqualError(t, err, "no credentials specified")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("No issuer URI specified", func(t *testing.T) {
		resolvedDisplayData, err := display.Resolve(verifiable.NewCredentialsArray(), "", nil)
		require.EqualError(t, err, "no issuer URI specified")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("Malformed issuer URI", func(t *testing.T) {
		opts := display.NewOpts()

		// Setting this for test coverage purposes. Actual testing of metrics logger functionality is handled
		// in the integration tests.
		opts.SetMetricsLogger(nil)

		resolvedDisplayData, err := display.Resolve(verifiable.NewCredentialsArray(), "badURL", opts)
		require.EqualError(t, err,
			"failed to get response from the issuer's metadata endpoint: "+
				`Get "badURL/.well-known/openid-credential-issuer": unsupported protocol scheme ""`)
		require.Nil(t, resolvedDisplayData)
	})
}

func TestResolveCredential(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:              t,
		issuerMetadata: string(sampleIssuerMetadata),
	}
	server := httptest.NewServer(issuerServerHandler)

	defer server.Close()

	parseVCOptionalArgs := verifiable.NewOpts()
	parseVCOptionalArgs.DisableProofCheck()

	t.Run("Success: CredentialArray v1", func(t *testing.T) {
		vc, err := verifiable.ParseCredential(credentialUniversityDegree, parseVCOptionalArgs)
		require.NoError(t, err)

		vcs := verifiable.NewCredentialsArray()
		vcs.Add(vc)

		opts := display.NewOpts().SetMaskingString("*")

		resolvedDisplayData, err := display.ResolveCredential(vcs, server.URL, opts)
		require.NoError(t, err)

		require.Equal(t, 2, resolvedDisplayData.LocalizedIssuersLength())
		require.Equal(t, 1, resolvedDisplayData.CredentialsLength())
		require.Equal(t, 1, resolvedDisplayData.CredentialAtIndex(0).LocalizedOverviewsLength())
		require.Equal(t, 6, resolvedDisplayData.CredentialAtIndex(0).SubjectsLength())

		credentialDisplay := resolvedDisplayData.CredentialAtIndex(0)

		for i := range credentialDisplay.SubjectsLength() {
			claim := credentialDisplay.SubjectAtIndex(i)

			if claim.LocalizedLabelAtIndex(0).Name() == sensitiveIDLabel {
				require.Equal(t, "*****6789", claim.Value())
			} else if claim.LocalizedLabelAtIndex(0).Name() == reallySensitiveIDLabel {
				require.Equal(t, "*******", claim.Value())
			}
		}
	})

	t.Run("Success: CredentialArray v2", func(t *testing.T) {
		vc, err := verifiable.ParseCredential(credentialUniversityDegree, parseVCOptionalArgs)
		require.NoError(t, err)

		vcs := verifiable.NewCredentialsArrayV2()
		vcs.Add(vc, "UniversityDegreeCredential_jwt_vc_json_v1")

		opts := display.NewOpts().SetMaskingString("*")

		resolvedDisplayData, err := display.ResolveCredentialV2(vcs, server.URL, opts)
		require.NoError(t, err)

		require.Equal(t, 2, resolvedDisplayData.LocalizedIssuersLength())
		require.Equal(t, 1, resolvedDisplayData.CredentialsLength())
		require.Equal(t, 1, resolvedDisplayData.CredentialAtIndex(0).LocalizedOverviewsLength())
		require.Equal(t, 6, resolvedDisplayData.CredentialAtIndex(0).SubjectsLength())

		credentialDisplay := resolvedDisplayData.CredentialAtIndex(0)

		for i := range credentialDisplay.SubjectsLength() {
			claim := credentialDisplay.SubjectAtIndex(i)

			if claim.LocalizedLabelAtIndex(0).Name() == sensitiveIDLabel {
				require.Equal(t, "*****6789", claim.Value())
			} else if claim.LocalizedLabelAtIndex(0).Name() == reallySensitiveIDLabel {
				require.Equal(t, "*******", claim.Value())
			}
		}
	})
}

func TestResolveCredentialOffer(t *testing.T) {
	metadata := &issuer.Metadata{}

	require.NoError(t, json.Unmarshal(sampleIssuerMetadata, metadata))

	resolvedDisplayData := display.ResolveCredentialOffer(
		openid4ci.IssuerMetadataFromGoImpl(metadata), api.NewStringArrayArray().Add(
			api.NewStringArray().Append("UniversityDegreeCredential")), "")

	checkIssuerDisplay(t, resolvedDisplayData.IssuerDisplay())

	require.Equal(t, 1, resolvedDisplayData.CredentialDisplaysLength())

	credentialDisplay := resolvedDisplayData.CredentialDisplayAtIndex(0)
	checkCredentialOverview(t, credentialDisplay)
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
	require.Equal(t, "en-US", issuerDisplay.Locale())
	require.Equal(t, "https://server.example.com", issuerDisplay.URL())
	require.Equal(t, "https://exampleuniversity.com/public/logo.png", issuerDisplay.Logo().URL())
	require.Equal(t, "#12107c", issuerDisplay.BackgroundColor())
	require.Equal(t, "#FFFFFF", issuerDisplay.TextColor())
}

func checkCredentialDisplay(t *testing.T, credentialDisplay *display.CredentialDisplay) {
	t.Helper()

	checkCredentialOverview(t, credentialDisplay)

	require.Equal(t, 6, credentialDisplay.ClaimsLength())

	checkClaims(t, credentialDisplay)
}

func checkCredentialOverview(t *testing.T, credentialDisplay *display.CredentialDisplay) {
	t.Helper()

	credentialOverview := credentialDisplay.Overview()
	require.Equal(t, "University Credential", credentialOverview.Name())
	require.Equal(t, "en-US", credentialOverview.Locale())
	require.Equal(t, "https://exampleuniversity.com/public/logo.png", credentialOverview.Logo().URL())
	require.Equal(t, "a square logo of a university", credentialOverview.Logo().AltText())
	require.Equal(t, "#12107c", credentialOverview.BackgroundColor())
	require.Equal(t, "#FFFFFF", credentialOverview.TextColor())
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
		IsMasked  bool
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
				Value:     "•••••6789",
				RawValue:  "123456789",
				Locale:    "en-US",
				IsMasked:  true,
			},
			{
				RawID:     "really_sensitive_id",
				Label:     "Really Sensitive ID",
				ValueType: "string",
				Value:     "•••••••",
				RawValue:  "abcdefg",
				Locale:    "en-US",
				IsMasked:  true,
			},
		},
	}
	expectedClaimsChecklist.Found = make([]bool, len(expectedClaimsChecklist.Claims))

	for i := range credentialDisplay.ClaimsLength() {
		claim := credentialDisplay.ClaimAtIndex(i)

		for j := range len(expectedClaimsChecklist.Claims) {
			expectedClaim := expectedClaimsChecklist.Claims[j]
			if claim.Label() == expectedClaim.Label &&
				claim.ValueType() == expectedClaim.ValueType &&
				claim.Value() == expectedClaim.Value &&
				claim.Locale() == expectedClaim.Locale &&
				claim.RawID() == expectedClaim.RawID &&
				claim.Pattern() == expectedClaim.Pattern &&
				claim.IsMasked() == expectedClaim.IsMasked &&
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

	for i := range len(expectedClaimsChecklist.Claims) {
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
