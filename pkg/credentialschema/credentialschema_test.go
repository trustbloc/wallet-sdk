/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema_test

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-go/doc/jose"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/credentialschema"
	"github.com/trustbloc/wallet-sdk/pkg/memstorage"
	"github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

const (
	sensitiveIDLabel       = "Sensitive ID"
	reallySensitiveIDLabel = "Really Sensitive ID"
)

var (
	//go:embed testdata/issuer_metadata.json
	sampleIssuerMetadata []byte

	//go:embed testdata/issuer_metadata_without_claims_display.json
	issuerMetadataWithoutClaimsDisplay []byte

	//go:embed testdata/credential_university_degree.jsonld
	credentialUniversityDegree []byte

	//go:embed testdata/open_badge_vc.jsonld
	openBadgeVC []byte

	//go:embed testdata/open_badge_issuer_metadata.json
	openBadgeMetadata []byte

	//go:embed testdata/unsupported_credential_multiple_subjects.jsonld
	unsupportedCredentialMultipleSubjects []byte

	//go:embed testdata/verified_employee_sd.jwt
	credentialVerifiedEmployeeSD []byte

	//go:embed testdata/bank_issuer_metadata.json
	bankIssuerMetadata []byte
)

type mockIssuerServerHandler struct{}

func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	_, err := writer.Write(sampleIssuerMetadata)
	if err != nil {
		println(err.Error())
	}
}

func TestResolve(t *testing.T) { //nolint:gocognit
	t.Run("Success", func(t *testing.T) {
		t.Run("Credentials supported object contains display info for the given VC", func(t *testing.T) {
			credential, err := verifiable.ParseCredential(credentialUniversityDegree,
				verifiable.WithCredDisableValidation(),
				verifiable.WithDisabledProofCheck())
			require.NoError(t, err)

			var issuerMetadata issuer.Metadata

			err = json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
			require.NoError(t, err)

			t.Run("Without preferred locale specified", func(t *testing.T) {
				resolvedDisplayData, errResolve := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithIssuerMetadata(&issuerMetadata))
				require.NoError(t, errResolve)

				checkSuccessCaseMatchedDisplayData(t, resolvedDisplayData)
			})
			t.Run("With preferred locale specified", func(t *testing.T) {
				t.Run("Issuer metadata has a localization for the given locale", func(t *testing.T) {
					resolvedDisplayData, errResolve := credentialschema.Resolve(
						credentialschema.WithCredentials([]*verifiable.Credential{credential}),
						credentialschema.WithIssuerMetadata(&issuerMetadata),
						credentialschema.WithHTTPClient(http.DefaultClient),
						credentialschema.WithPreferredLocale("en-US"))
					require.NoError(t, errResolve)

					checkSuccessCaseMatchedDisplayData(t, resolvedDisplayData)
				})
				t.Run("Issuer metadata does not have a localization for the given locale", func(t *testing.T) {
					resolvedDisplayData, errResolve := credentialschema.Resolve(
						credentialschema.WithCredentials([]*verifiable.Credential{credential}),
						credentialschema.WithIssuerMetadata(&issuerMetadata),
						credentialschema.WithHTTPClient(http.DefaultClient),
						credentialschema.WithPreferredLocale("UnknownLocale"))
					require.NoError(t, errResolve)

					checkSuccessCaseMatchedDisplayData(t, resolvedDisplayData)
				})
			})
			t.Run("With credential reader instead of directly passing in VC", func(t *testing.T) {
				memStorageProvider := memstorage.NewProvider()

				errAdd := memStorageProvider.Add(credential)
				require.NoError(t, errAdd)

				resolvedDisplayData, errResolve := credentialschema.Resolve(
					credentialschema.WithCredentialReader(memStorageProvider,
						[]string{"http://example.edu/credentials/1872"}),
					credentialschema.WithIssuerMetadata(&issuerMetadata),
					credentialschema.WithHTTPClient(http.DefaultClient),
					credentialschema.WithPreferredLocale("en-US"))
				require.NoError(t, errResolve)

				checkSuccessCaseMatchedDisplayData(t, resolvedDisplayData)
			})
			t.Run("With issuer URI instead of directly passing in issuer metadata", func(t *testing.T) {
				issuerServerHandler := &mockIssuerServerHandler{}
				server := httptest.NewServer(issuerServerHandler)

				defer server.Close()

				resolvedDisplayData, errResolve := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithHTTPClient(http.DefaultClient),
					credentialschema.WithIssuerURI(server.URL),
					credentialschema.WithJWTSignatureVerifier(&mockSignatureVerifier{}))
				require.NoError(t, errResolve)

				checkSuccessCaseMatchedDisplayData(t, resolvedDisplayData)
			})

			t.Run("Skip non claim data", func(t *testing.T) {
				resolvedDisplayData, errResolve := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithIssuerMetadata(&issuerMetadata),
					credentialschema.WithSkipNonClaimData())
				require.NoError(t, errResolve)

				checkSuccessCaseMatchedDisplayData(t, resolvedDisplayData)
			})

			t.Run("Use metrics logger", func(t *testing.T) {
				resolvedDisplayData, errResolve := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithIssuerMetadata(&issuerMetadata),
					credentialschema.WithMetricsLogger(noop.NewMetricsLogger()))
				require.NoError(t, errResolve)

				checkSuccessCaseMatchedDisplayData(t, resolvedDisplayData)
			})

			t.Run("Credentials supported object does not contain display info for the given VC, "+
				"resulting in the default display being used", func(t *testing.T) {
				var metadata issuer.Metadata

				err = json.Unmarshal(sampleIssuerMetadata, &metadata)
				require.NoError(t, err)

				metadata.CredentialConfigurationsSupported["UniversityDegreeCredential_jwt_vc_json-ld_v1"].
					CredentialDefinition.Type[1] = "SomeOtherType"

				resolvedDisplayData, errResolve := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithIssuerMetadata(&metadata),
					credentialschema.WithHTTPClient(http.DefaultClient),
					credentialschema.WithPreferredLocale("en-US"))
				require.NoError(t, errResolve)
				checkForDefaultDisplayData(t, resolvedDisplayData)
			})
			t.Run("With custom masking string", func(t *testing.T) {
				t.Run(`"*"`, func(t *testing.T) {
					resolvedDisplayData, errResolve := credentialschema.Resolve(
						credentialschema.WithCredentials([]*verifiable.Credential{credential}),
						credentialschema.WithIssuerMetadata(&issuerMetadata),
						credentialschema.WithMaskingString("*"))
					require.NoError(t, errResolve)

					claims := resolvedDisplayData.CredentialDisplays[0].Claims

					for _, claim := range claims {
						if claim.Label == sensitiveIDLabel {
							require.Equal(t, "*****6789", *claim.Value)
						} else if claim.Label == reallySensitiveIDLabel {
							require.Equal(t, "*******", *claim.Value)
						}
					}
				})
				t.Run(`"+++"`, func(t *testing.T) {
					resolvedDisplayData, errResolve := credentialschema.Resolve(
						credentialschema.WithCredentials([]*verifiable.Credential{credential}),
						credentialschema.WithIssuerMetadata(&issuerMetadata),
						credentialschema.WithMaskingString("+++"))
					require.NoError(t, errResolve)

					claims := resolvedDisplayData.CredentialDisplays[0].Claims

					for _, claim := range claims {
						if claim.Label == sensitiveIDLabel {
							require.Equal(t, "+++++++++++++++6789", *claim.Value)
						} else if claim.Label == reallySensitiveIDLabel {
							require.Equal(t, "+++++++++++++++++++++", *claim.Value)
						}
					}
				})
				t.Run(`"" (empty string)`, func(t *testing.T) {
					resolvedDisplayData, errResolve := credentialschema.Resolve(
						credentialschema.WithCredentials([]*verifiable.Credential{credential}),
						credentialschema.WithIssuerMetadata(&issuerMetadata),
						credentialschema.WithMaskingString(""))
					require.NoError(t, errResolve)

					claims := resolvedDisplayData.CredentialDisplays[0].Claims

					for _, claim := range claims {
						if claim.Label == sensitiveIDLabel {
							require.Equal(t, "6789", *claim.Value)
						} else if claim.Label == reallySensitiveIDLabel {
							require.Equal(t, "", *claim.Value)
						}
					}
				})
			})

			t.Run("Credentials supported object does not have claim display info", func(t *testing.T) {
				var localIssuerMetadata issuer.Metadata

				err = json.Unmarshal(issuerMetadataWithoutClaimsDisplay, &localIssuerMetadata)
				require.NoError(t, err)

				resolvedDisplayData, errResolve := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithIssuerMetadata(&localIssuerMetadata),
					credentialschema.WithHTTPClient(http.DefaultClient),
					credentialschema.WithPreferredLocale("en-US"))
				require.NoError(t, errResolve)
				require.Equal(t, "Example University", resolvedDisplayData.IssuerDisplay.Name)
				require.Equal(t, "en-US", resolvedDisplayData.IssuerDisplay.Locale)
				require.Len(t, resolvedDisplayData.CredentialDisplays, 1)
				require.Equal(t, "University Credential",
					resolvedDisplayData.CredentialDisplays[0].Overview.Name)
				require.Equal(t, "en-US",
					resolvedDisplayData.CredentialDisplays[0].Overview.Locale)
				require.Equal(t, "https://exampleuniversity.com/public/logo.png",
					resolvedDisplayData.CredentialDisplays[0].Overview.Logo.URL)
				require.Equal(t, "a square logo of a university",
					resolvedDisplayData.CredentialDisplays[0].Overview.Logo.AltText)
				require.Equal(t, "#12107c", resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
				require.Equal(t, "#FFFFFF", resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)
				require.Nil(t, resolvedDisplayData.CredentialDisplays[0].Claims)
			})
			t.Run("Issuer metadata does not have the optional issuer display info", func(t *testing.T) {
				var localIssuerMetadata issuer.Metadata

				err = json.Unmarshal(sampleIssuerMetadata, &localIssuerMetadata)
				require.NoError(t, err)

				localIssuerMetadata.LocalizedIssuerDisplays = nil

				resolvedDisplayData, err := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithHTTPClient(http.DefaultClient),
					credentialschema.WithIssuerMetadata(&localIssuerMetadata))
				require.NoError(t, err)

				require.Nil(t, resolvedDisplayData.IssuerDisplay)
				require.Len(t, resolvedDisplayData.CredentialDisplays, 1)
				require.Equal(t, "University Credential",
					resolvedDisplayData.CredentialDisplays[0].Overview.Name)
				require.Equal(t, "en-US",
					resolvedDisplayData.CredentialDisplays[0].Overview.Locale)
				require.Equal(t, "https://exampleuniversity.com/public/logo.png",
					resolvedDisplayData.CredentialDisplays[0].Overview.Logo.URL)
				require.Equal(t, "a square logo of a university",
					resolvedDisplayData.CredentialDisplays[0].Overview.Logo.AltText)
				require.Equal(t, "#12107c", resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
				require.Equal(t, "#FFFFFF", resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)

				expectedIDOrder := 0
				expectedGivenNameOrder := 1
				expectedSurnameOrder := 2
				sensitiveIDValue := "•••••6789"
				reallySensitiveIDValue := "•••••••"
				expectedClaims := []credentialschema.ResolvedClaim{
					{RawID: "id", Label: "ID", RawValue: "1234", ValueType: "string", Locale: "en-US", Order: &expectedIDOrder},
					{
						RawID: "given_name", Label: "Given Name", RawValue: "Alice", ValueType: "string", Locale: "en-US",
						Order: &expectedGivenNameOrder,
					},
					{
						RawID: "surname", Label: "Surname", RawValue: "Bowman", ValueType: "string", Locale: "en-US",
						Order: &expectedSurnameOrder,
					},
					{RawID: "gpa", Label: "GPA", RawValue: "4.0", ValueType: "number", Locale: "en-US"},
					{
						RawID: "sensitive_id", Label: "Sensitive ID", RawValue: "123456789",
						Value: &sensitiveIDValue, ValueType: "string", Mask: "regex(^(.*).{4}$)", Locale: "en-US",
					},
					{
						RawID: "really_sensitive_id", Label: "Really Sensitive ID", RawValue: "abcdefg",
						Value: &reallySensitiveIDValue, ValueType: "string", Mask: "regex((.*))", Locale: "en-US",
					},
					{
						RawID: "chemistry", Label: "Chemistry Final Grade", RawValue: "78",
						ValueType: "number", Locale: "en-US",
					},
				}

				verifyClaimsAnyOrder(t, resolvedDisplayData.CredentialDisplays[0].Claims, expectedClaims)
			})
			t.Run("VC does not have the subject fields specified by the claim display info", func(t *testing.T) {
				var rawCred verifiable.JSONObject

				require.NoError(t, json.Unmarshal(credentialUniversityDegree, &rawCred))
				// TODO: it not works in case of nil credentialSubject, but works with empty subject id.
				// Is empty subject id has sense at all?
				rawCred["credentialSubject"] = map[string]interface{}{
					"id": "",
				}

				verifiableCredential, err := verifiable.ParseCredentialJSON(rawCred,
					verifiable.WithCredDisableValidation(),
					verifiable.WithDisabledProofCheck())
				require.NoError(t, err)

				var localIssuerMetadata issuer.Metadata

				err = json.Unmarshal(sampleIssuerMetadata, &localIssuerMetadata)
				require.NoError(t, err)

				resolvedDisplayData, err := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{verifiableCredential}),
					credentialschema.WithHTTPClient(http.DefaultClient),
					credentialschema.WithIssuerMetadata(&localIssuerMetadata))
				require.NoError(t, err)

				require.Equal(t, "Example University", resolvedDisplayData.IssuerDisplay.Name)
				require.Equal(t, "en-US", resolvedDisplayData.IssuerDisplay.Locale)
				require.Len(t, resolvedDisplayData.CredentialDisplays, 1)
				require.Equal(t, "University Credential",
					resolvedDisplayData.CredentialDisplays[0].Overview.Name)
				require.Equal(t, "en-US",
					resolvedDisplayData.CredentialDisplays[0].Overview.Locale)
				require.Equal(t, "https://exampleuniversity.com/public/logo.png",
					resolvedDisplayData.CredentialDisplays[0].Overview.Logo.URL)
				require.Equal(t, "a square logo of a university",
					resolvedDisplayData.CredentialDisplays[0].Overview.Logo.AltText)
				require.Equal(t, "#12107c", resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
				require.Equal(t, "#FFFFFF", resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)

				require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Claims)
			})

			t.Run("Credentials_without_JWT_envelop", func(t *testing.T) {
				credential := *credential
				resolvedDisplayData, errResolve := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{&credential}),
					credentialschema.WithIssuerMetadata(&issuerMetadata))
				require.NoError(t, errResolve)

				checkSuccessCaseMatchedDisplayData(t, resolvedDisplayData)
			})
		})

		t.Run("Correctly shown display info for selective disclosure JWT", func(t *testing.T) {
			credential, err := verifiable.ParseCredential(credentialVerifiedEmployeeSD,
				verifiable.WithCredDisableValidation(),
				verifiable.WithDisabledProofCheck())
			require.NoError(t, err)

			var issuerMetadata issuer.Metadata

			err = json.Unmarshal(bankIssuerMetadata, &issuerMetadata)
			require.NoError(t, err)

			resolvedDisplayData, errResolve := credentialschema.Resolve(
				credentialschema.WithCredentials([]*verifiable.Credential{credential}),
				credentialschema.WithHTTPClient(http.DefaultClient),
				credentialschema.WithIssuerMetadata(&issuerMetadata))
			require.NoError(t, errResolve)

			checkSDVCMatchedDisplayData(t, resolvedDisplayData)
		})
	})
	t.Run("Invalid options:", func(t *testing.T) {
		t.Run("No credentials specified", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve()
			require.EqualError(t, err, "no credentials specified")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Multiple credential sources", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithCredentials([]*verifiable.Credential{{}}),
				credentialschema.WithHTTPClient(http.DefaultClient),
				credentialschema.WithCredentialReader(memstorage.NewProvider(), []string{}))
			require.EqualError(t, err, "cannot have multiple credential sources specified - "+
				"must use either WithCredentials or WithCredentialReader, but not both")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Using credential reader, but no IDs specified", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithHTTPClient(http.DefaultClient),
				credentialschema.WithCredentialReader(memstorage.NewProvider(), []string{}))
			require.EqualError(t, err, "credential IDs must be provided when using a credential reader")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Using credential reader, but credential could not be found", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithCredentialReader(memstorage.NewProvider(), []string{"SomeID"}),
				credentialschema.WithHTTPClient(http.DefaultClient),
				credentialschema.WithIssuerMetadata(&issuer.Metadata{}))
			require.EqualError(t, err, "no credential with an id of SomeID was found")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("No issuer metadata source specified", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithHTTPClient(http.DefaultClient),
				credentialschema.WithCredentials([]*verifiable.Credential{{}}))
			require.EqualError(t, err, "no issuer metadata source specified")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Using issuer URI option, but failed to fetch issuer metadata", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithCredentials([]*verifiable.Credential{{}}),
				credentialschema.WithHTTPClient(http.DefaultClient),
				credentialschema.WithIssuerURI("http://BadURL"))
			require.Contains(t, err.Error(), `Get "http://BadURL/.well-known/openid-credential-issuer":`+
				` dial tcp: lookup BadURL`)
			require.Nil(t, resolvedDisplayData)
		})
	})
	t.Run("Unsupported VC", func(t *testing.T) {
		t.Run("Multiple subjects", func(t *testing.T) {
			credential, err := verifiable.ParseCredential(unsupportedCredentialMultipleSubjects,
				verifiable.WithCredDisableValidation(),
				verifiable.WithDisabledProofCheck())
			require.NoError(t, err)

			var issuerMetadata issuer.Metadata

			err = json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
			require.NoError(t, err)

			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithCredentials([]*verifiable.Credential{credential}),
				credentialschema.WithHTTPClient(http.DefaultClient),
				credentialschema.WithIssuerMetadata(&issuerMetadata))
			require.EqualError(t, err, "only VCs with one credential subject are supported")
			require.Nil(t, resolvedDisplayData)
		})
	})
	t.Run("Fail to compile regex", func(t *testing.T) {
		credential, err := verifiable.ParseCredential(credentialUniversityDegree,
			verifiable.WithCredDisableValidation(),
			verifiable.WithDisabledProofCheck())
		require.NoError(t, err)

		var issuerMetadata issuer.Metadata

		err = json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
		require.NoError(t, err)

		credentialConf := issuerMetadata.CredentialConfigurationsSupported["UniversityDegreeCredential_jwt_vc_json-ld_v1"]
		credentialConf.CredentialDefinition.CredentialSubject["sensitive_id"] = &issuer.Claim{
			LocalizedClaimDisplays: []issuer.LocalizedClaimDisplay{{}},
			Mask:                   "regex(()",
		}

		resolvedDisplayData, errResolve := credentialschema.Resolve(
			credentialschema.WithCredentials([]*verifiable.Credential{credential}),
			credentialschema.WithHTTPClient(http.DefaultClient),
			credentialschema.WithIssuerMetadata(&issuerMetadata))
		require.EqualError(t, errResolve, "error parsing regexp: missing closing ): `(`")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("Option fails", func(t *testing.T) {
		credential, err := verifiable.ParseCredential(credentialUniversityDegree,
			verifiable.WithCredDisableValidation(),
			verifiable.WithDisabledProofCheck())
		require.NoError(t, err)

		var issuerMetadata issuer.Metadata

		err = json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
		require.NoError(t, err)

		t.Run("Credential config not found", func(t *testing.T) {
			resolvedDisplayData, errResolve := credentialschema.Resolve(
				credentialschema.WithCredentials([]*verifiable.Credential{credential}, "12234"),
				credentialschema.WithIssuerMetadata(&issuerMetadata),
				credentialschema.WithSkipNonClaimData())
			require.Error(t, errResolve)

			require.ErrorContains(t, errResolve, "credential configuration with ID 12234 not found")
			require.Nil(t, resolvedDisplayData)
		})

		t.Run("Mismatch credential configs", func(t *testing.T) {
			resolvedDisplayData, errResolve := credentialschema.Resolve(
				credentialschema.WithCredentials([]*verifiable.Credential{credential}, "12234", "1245"),
				credentialschema.WithIssuerMetadata(&issuerMetadata),
				credentialschema.WithSkipNonClaimData())
			require.Error(t, errResolve)

			require.ErrorContains(t, errResolve, "mismatch between the number of credentials")
			require.Nil(t, resolvedDisplayData)
		})
	})
}

func TestResolveMetadataWithJsonPath(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		credential, err := verifiable.ParseCredential(openBadgeVC,
			verifiable.WithCredDisableValidation(),
			verifiable.WithDisabledProofCheck())
		require.NoError(t, err)

		var issuerMetadata issuer.Metadata

		err = json.Unmarshal(openBadgeMetadata, &issuerMetadata)
		require.NoError(t, err)

		resolvedDisplayData, errResolve := credentialschema.Resolve(
			credentialschema.WithCredentials([]*verifiable.Credential{credential}),
			credentialschema.WithIssuerMetadata(&issuerMetadata))
		require.NoError(t, errResolve)
		require.Len(t, resolvedDisplayData.CredentialDisplays[0].Claims, 7)
	})
}

func TestResolveCredentialOffer(t *testing.T) {
	var issuerMetadata issuer.Metadata
	err := json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		resolvedDisplayData := credentialschema.ResolveCredentialOffer(
			&issuerMetadata, [][]string{{"UniversityDegreeCredential"}}, "")

		checkSuccessCaseMatchedOverviewData(t, resolvedDisplayData)
	})
}

func TestResolveCredential(t *testing.T) {
	credential, err := verifiable.ParseCredential(credentialUniversityDegree,
		verifiable.WithCredDisableValidation(),
		verifiable.WithDisabledProofCheck())
	require.NoError(t, err)

	var issuerMetadata issuer.Metadata

	err = json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
	require.NoError(t, err)

	resolvedDisplayData, errResolve := credentialschema.ResolveCredential(
		credentialschema.WithCredentials([]*verifiable.Credential{credential}),
		credentialschema.WithIssuerMetadata(&issuerMetadata))
	require.NoError(t, errResolve)

	require.Len(t, resolvedDisplayData.Credential, 1)
}

func checkSuccessCaseMatchedOverviewData(t *testing.T, resolvedDisplayData *credentialschema.ResolvedDisplayData) {
	t.Helper()

	require.Equal(t, "Example University", resolvedDisplayData.IssuerDisplay.Name)
	require.Equal(t, "en-US", resolvedDisplayData.IssuerDisplay.Locale)
	require.Len(t, resolvedDisplayData.CredentialDisplays, 1)
	require.Equal(t, "University Credential",
		resolvedDisplayData.CredentialDisplays[0].Overview.Name)
	require.Equal(t, "en-US",
		resolvedDisplayData.CredentialDisplays[0].Overview.Locale)
	require.Equal(t, "https://exampleuniversity.com/public/logo.png",
		resolvedDisplayData.CredentialDisplays[0].Overview.Logo.URL)
	require.Equal(t, "a square logo of a university",
		resolvedDisplayData.CredentialDisplays[0].Overview.Logo.AltText)
	require.Equal(t, "#12107c", resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
	require.Equal(t, "#FFFFFF", resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)
}

func checkSuccessCaseMatchedDisplayData(t *testing.T, resolvedDisplayData *credentialschema.ResolvedDisplayData) {
	t.Helper()

	checkSuccessCaseMatchedOverviewData(t, resolvedDisplayData)

	expectedIDOrder := 0
	expectedGivenNameOrder := 1
	expectedSurnameOrder := 2
	sensitiveIDValue := "•••••6789"
	reallySensitiveIDValue := "•••••••"
	expectedClaims := []credentialschema.ResolvedClaim{
		{RawID: "id", Label: "ID", RawValue: "1234", ValueType: "string", Locale: "en-US", Order: &expectedIDOrder},
		{
			RawID: "given_name", Label: "Given Name", RawValue: "Alice", ValueType: "string",
			Locale: "en-US", Order: &expectedGivenNameOrder,
		},
		{
			RawID: "surname", Label: "Surname", RawValue: "Bowman", ValueType: "string",
			Locale: "en-US", Order: &expectedSurnameOrder,
		},
		{RawID: "gpa", Label: "GPA", RawValue: "4.0", ValueType: "number", Locale: "en-US"},
		{
			RawID: "sensitive_id", Label: "Sensitive ID", RawValue: "123456789", Value: &sensitiveIDValue,
			ValueType: "string", Mask: "regex(^(.*).{4}$)", Locale: "en-US",
		},
		{
			RawID: "really_sensitive_id", Label: "Really Sensitive ID", RawValue: "abcdefg",
			Value: &reallySensitiveIDValue, ValueType: "string", Mask: "regex((.*))", Locale: "en-US",
		},
		{
			RawID: "chemistry", Label: "Chemistry Final Grade", RawValue: "78",
			ValueType: "number", Locale: "en-US",
		},
	}

	verifyClaimsAnyOrder(t, resolvedDisplayData.CredentialDisplays[0].Claims, expectedClaims)
}

func checkForDefaultDisplayData(t *testing.T, resolvedDisplayData *credentialschema.ResolvedDisplayData) {
	t.Helper()

	require.Equal(t, "Example University", resolvedDisplayData.IssuerDisplay.Name)
	require.Equal(t, "en-US", resolvedDisplayData.IssuerDisplay.Locale)
	require.Len(t, resolvedDisplayData.CredentialDisplays, 1)
	require.Equal(t, "http://example.edu/credentials/1872",
		resolvedDisplayData.CredentialDisplays[0].Overview.Name)
	require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Overview.Locale)
	require.Nil(t, resolvedDisplayData.CredentialDisplays[0].Overview.Logo)
	require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
	require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)

	expectedClaims := []credentialschema.ResolvedClaim{
		{RawID: "id", RawValue: "1234", ValueType: "string"},
		{RawID: "given_name", RawValue: "Alice"},
		{RawID: "surname", RawValue: "Bowman"},
		{RawID: "gpa", RawValue: "4.0"},
		{RawID: "sensitive_id", RawValue: "123456789"},
		{RawID: "really_sensitive_id", RawValue: "abcdefg"},
		{RawID: "course_grades", RawValue: "map[chemistry:78 physics:85]"},
	}

	verifyClaimsAnyOrder(t, resolvedDisplayData.CredentialDisplays[0].Claims, expectedClaims)
}

func checkSDVCMatchedDisplayData(t *testing.T, resolvedDisplayData *credentialschema.ResolvedDisplayData) {
	t.Helper()

	require.Equal(t, "Bank Issuer", resolvedDisplayData.IssuerDisplay.Name)
	require.Equal(t, "en-US", resolvedDisplayData.IssuerDisplay.Locale)
	require.Len(t, resolvedDisplayData.CredentialDisplays, 1)
	require.Equal(t, "Verified Employee",
		resolvedDisplayData.CredentialDisplays[0].Overview.Name)
	require.Equal(t, "en-US",
		resolvedDisplayData.CredentialDisplays[0].Overview.Locale)
	require.Equal(t, "https://example.com/public/logo.png",
		resolvedDisplayData.CredentialDisplays[0].Overview.Logo.URL)
	require.Equal(t, "a square logo of an employee verification",
		resolvedDisplayData.CredentialDisplays[0].Overview.Logo.AltText)
	require.Equal(t, "#12107c", resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
	require.Equal(t, "#FFFFFF", resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)

	expectedClaims := []credentialschema.ResolvedClaim{
		{RawID: "displayName", Label: "Employee", RawValue: "John Doe", ValueType: "string", Locale: "en-US"},
		{RawID: "givenName", Label: "Given Name", RawValue: "John", ValueType: "string", Locale: "en-US"},
		{RawID: "surname", Label: "Surname", RawValue: "Doe", ValueType: "string", Locale: "en-US"},
		{RawID: "jobTitle", Label: "Job Title", RawValue: "Software Developer", ValueType: "string", Locale: "en-US"},
		{RawID: "mail", Label: "Mail", RawValue: "john.doe@foo.bar", ValueType: "string", Locale: "en-US"},
		{RawID: "photo", Label: "Photo", RawValue: "base64photo", ValueType: "image", Locale: ""},
		{RawID: "preferredLanguage", Label: "Preferred Language", RawValue: "English", ValueType: "string", Locale: "en-US"},
	}

	verifyClaimsAnyOrder(t, resolvedDisplayData.CredentialDisplays[0].Claims, expectedClaims)
}

func verifyClaimsAnyOrder(t *testing.T, actualClaims []credentialschema.ResolvedClaim,
	expectedClaims []credentialschema.ResolvedClaim,
) {
	t.Helper()

	require.Len(t, actualClaims, len(expectedClaims), "expected (in any order): [%v]", expectedClaims)

	claimsMatched := make([]bool, len(expectedClaims))

	for i := range actualClaims {
		for j := range expectedClaims {
			if claimsMatched[j] {
				continue
			}

			if claimsMatch(&actualClaims[i], &expectedClaims[j]) {
				claimsMatched[j] = true

				break
			}
		}
	}

	for _, claimMatched := range claimsMatched {
		if !claimMatched {
			require.FailNow(t, "received unexpected claims",
				"actual: [%v] expected (in any order): [%v]", actualClaims, expectedClaims)
		}
	}
}

func claimsMatch(claim1, claim2 *credentialschema.ResolvedClaim) bool { //nolint: gocyclo // Test file
	// Check the value pointer fields first.
	if claim1.Value != nil {
		if claim2.Value == nil {
			return false
		}

		if *claim1.Value != *claim2.Value {
			return false
		}
	} else if claim2.Value != nil {
		return false
	}

	if claim1.Label == claim2.Label &&
		claim1.Locale == claim2.Locale &&
		claim1.ValueType == claim2.ValueType &&
		claim1.RawID == claim2.RawID &&
		claim1.Pattern == claim2.Pattern &&
		claim1.Mask == claim2.Mask &&
		claim1.RawValue == claim2.RawValue {
		return ordersMatch(claim1.Order, claim2.Order)
	}

	return false
}

func ordersMatch(order1, order2 *int) bool {
	if order1 != nil {
		if order2 == nil {
			return false
		}

		if *order1 != *order2 {
			return false
		}
	} else if order2 != nil {
		return false
	}

	return true
}

type mockSignatureVerifier struct{}

func (*mockSignatureVerifier) CheckJWTProof(jose.Headers, string, []byte, []byte) error {
	return nil
}
