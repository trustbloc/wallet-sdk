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

	"github.com/trustbloc/wallet-sdk/pkg/memstorage"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/credentialschema"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

var (
	//go:embed testdata/issuer_metadata.json
	sampleIssuerMetadata []byte

	//go:embed testdata/issuer_metadata_without_claims_display.json
	issuerMetadataWithoutClaimsDisplay []byte

	//go:embed testdata/credential_university_degree.jsonld
	credentialUniversityDegree []byte

	//go:embed testdata/unsupported_credential_string_subject.jsonld
	unsupportedCredentialStringSubject []byte

	//go:embed testdata/unsupported_credential_multiple_subjects.jsonld
	unsupportedCredentialMultipleSubjects []byte
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
		t.Run("Credentials supported object contains display info for the given VC", func(t *testing.T) {
			credential, err := verifiable.ParseCredential(credentialUniversityDegree,
				verifiable.WithJSONLDDocumentLoader(ld.NewDefaultDocumentLoader(http.DefaultClient)),
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
						credentialschema.WithPreferredLocale("en-US"))
					require.NoError(t, errResolve)

					checkSuccessCaseMatchedDisplayData(t, resolvedDisplayData)
				})
				t.Run("Issuer metadata does not have a localization for the given locale", func(t *testing.T) {
					resolvedDisplayData, errResolve := credentialschema.Resolve(
						credentialschema.WithCredentials([]*verifiable.Credential{credential}),
						credentialschema.WithIssuerMetadata(&issuerMetadata),
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
					credentialschema.WithIssuerURI(server.URL))
				require.NoError(t, errResolve)

				checkSuccessCaseMatchedDisplayData(t, resolvedDisplayData)
			})
			t.Run("Credentials supported object does not contain display info for the given VC, "+
				"resulting in the default display being used", func(t *testing.T) {
				var issuerMetadata issuer.Metadata

				err = json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
				require.NoError(t, err)

				issuerMetadata.CredentialsSupported[0].Types[1] = "SomeOtherType"

				resolvedDisplayData, errResolve := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithIssuerMetadata(&issuerMetadata),
					credentialschema.WithPreferredLocale("en-US"))
				require.NoError(t, errResolve)
				require.Equal(t, "Example University", resolvedDisplayData.IssuerDisplay.Name)
				require.Equal(t, "en-US", resolvedDisplayData.IssuerDisplay.Locale)
				require.Len(t, resolvedDisplayData.CredentialDisplays, 1)
				require.Equal(t, "http://example.edu/credentials/1872",
					resolvedDisplayData.CredentialDisplays[0].Overview.Name)
				require.Equal(t, "N/A",
					resolvedDisplayData.CredentialDisplays[0].Overview.Locale)
				require.Nil(t, resolvedDisplayData.CredentialDisplays[0].Overview.Logo)
				require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
				require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)

				expectedClaims := []credentialschema.ResolvedClaim{
					{Label: "ID", Value: "1234", Locale: "N/A"},
					{Label: "given_name", Value: "Alice", Locale: "N/A"},
					{Label: "surname", Value: "Bowman", Locale: "N/A"},
					{Label: "gpa", Value: "4.0", Locale: "N/A"},
				}

				verifyClaimsAnyOrder(t, resolvedDisplayData.CredentialDisplays[0].Claims, expectedClaims)
			})
			t.Run("Credentials supported object does not have claim display info", func(t *testing.T) {
				var issuerMetadata issuer.Metadata

				err = json.Unmarshal(issuerMetadataWithoutClaimsDisplay, &issuerMetadata)
				require.NoError(t, err)

				resolvedDisplayData, errResolve := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithIssuerMetadata(&issuerMetadata),
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
				require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Overview.Logo.AltText)
				require.Equal(t, "#12107c", resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
				require.Equal(t, "#FFFFFF", resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)
				require.Len(t, resolvedDisplayData.CredentialDisplays[0].Claims, 1)
				require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Claims[0])
			})
			t.Run("Issuer metadata does not have the optional issuer display info", func(t *testing.T) {
				var issuerMetadata issuer.Metadata

				err = json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
				require.NoError(t, err)

				issuerMetadata.CredentialIssuer = nil

				resolvedDisplayData, err := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithIssuerMetadata(&issuerMetadata))
				require.NoError(t, err)

				require.Nil(t, resolvedDisplayData.IssuerDisplay)
				require.Len(t, resolvedDisplayData.CredentialDisplays, 1)
				require.Equal(t, "University Credential",
					resolvedDisplayData.CredentialDisplays[0].Overview.Name)
				require.Equal(t, "en-US",
					resolvedDisplayData.CredentialDisplays[0].Overview.Locale)
				require.Equal(t, "https://exampleuniversity.com/public/logo.png",
					resolvedDisplayData.CredentialDisplays[0].Overview.Logo.URL)
				require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Overview.Logo.AltText)
				require.Equal(t, "#12107c", resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
				require.Equal(t, "#FFFFFF", resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)

				expectedClaims := []credentialschema.ResolvedClaim{
					{Label: "ID", Value: "1234", Locale: "en-US"},
					{Label: "Given Name", Value: "Alice", Locale: "en-US"},
					{Label: "Surname", Value: "Bowman", Locale: "en-US"},
					{Label: "GPA", Value: "4.0", Locale: "en-US"},
				}

				verifyClaimsAnyOrder(t, resolvedDisplayData.CredentialDisplays[0].Claims, expectedClaims)
			})
			t.Run("VC does not have the subject fields specified by the claim display info", func(t *testing.T) {
				credential, err := verifiable.ParseCredential(credentialUniversityDegree,
					verifiable.WithJSONLDDocumentLoader(ld.NewDefaultDocumentLoader(http.DefaultClient)),
					verifiable.WithDisabledProofCheck())
				require.NoError(t, err)

				credential.Subject = []verifiable.Subject{
					{
						ID:           "",
						CustomFields: nil,
					},
				}

				var issuerMetadata issuer.Metadata

				err = json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
				require.NoError(t, err)

				resolvedDisplayData, err := credentialschema.Resolve(
					credentialschema.WithCredentials([]*verifiable.Credential{credential}),
					credentialschema.WithIssuerMetadata(&issuerMetadata))
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
				require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Overview.Logo.AltText)
				require.Equal(t, "#12107c", resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
				require.Equal(t, "#FFFFFF", resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)

				expectedClaims := []credentialschema.ResolvedClaim{
					{Label: "ID", Value: "", Locale: "en-US"},
					{Label: "Given Name", Value: "", Locale: "en-US"},
					{Label: "Surname", Value: "", Locale: "en-US"},
					{Label: "GPA", Value: "", Locale: "en-US"},
				}

				verifyClaimsAnyOrder(t, resolvedDisplayData.CredentialDisplays[0].Claims, expectedClaims)
			})
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
				credentialschema.WithCredentialReader(memstorage.NewProvider(), []string{}))
			require.EqualError(t, err, "cannot have multiple credential sources specified - "+
				"must use either WithCredentials or WithCredentialReader, but not both")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Using credential reader, but no IDs specified", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithCredentialReader(memstorage.NewProvider(), []string{}))
			require.EqualError(t, err, "credential IDs must be provided when using a credential reader")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Using credential reader, but credential could not be found", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithCredentialReader(memstorage.NewProvider(), []string{"SomeID"}),
				credentialschema.WithIssuerMetadata(&issuer.Metadata{}))
			require.EqualError(t, err, "no credential with an id of SomeID was found")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("No issuer metadata source specified", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithCredentials([]*verifiable.Credential{{}}))
			require.EqualError(t, err, "no issuer metadata source specified")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Using issuer URI option, but failed to fetch issuer metadata", func(t *testing.T) {
			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithCredentials([]*verifiable.Credential{{}}),
				credentialschema.WithIssuerURI("http://BadURL"))
			require.Contains(t, err.Error(), `Get "http://BadURL/.well-known/openid-configuration":`+
				` dial tcp: lookup BadURL:`)
			require.Nil(t, resolvedDisplayData)
		})
	})
	t.Run("Invalid supported credentials object", func(t *testing.T) {
		metadata := issuer.Metadata{
			CredentialsSupported: []issuer.SupportedCredential{{ID: "SomeID"}, {ID: "SomeID"}},
		}

		resolvedDisplayData, err := credentialschema.Resolve(
			credentialschema.WithCredentials([]*verifiable.Credential{{}}),
			credentialschema.WithIssuerMetadata(&metadata))
		require.EqualError(t, err, "issuer metadata's supported credentials object is invalid: "+
			"the ID SomeID appears in multiple supported credential objects")
		require.Nil(t, resolvedDisplayData)
	})
	t.Run("Unsupported VC", func(t *testing.T) {
		t.Run("Unsupported subject type", func(t *testing.T) {
			credential, err := verifiable.ParseCredential(unsupportedCredentialStringSubject,
				verifiable.WithJSONLDDocumentLoader(ld.NewDefaultDocumentLoader(http.DefaultClient)),
				verifiable.WithDisabledProofCheck())
			require.NoError(t, err)

			var issuerMetadata issuer.Metadata

			err = json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
			require.NoError(t, err)

			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithCredentials([]*verifiable.Credential{credential}),
				credentialschema.WithIssuerMetadata(&issuerMetadata))
			require.EqualError(t, err, "unsupported vc subject type")
			require.Nil(t, resolvedDisplayData)
		})
		t.Run("Multiple subjects", func(t *testing.T) {
			credential, err := verifiable.ParseCredential(unsupportedCredentialMultipleSubjects,
				verifiable.WithJSONLDDocumentLoader(ld.NewDefaultDocumentLoader(http.DefaultClient)),
				verifiable.WithDisabledProofCheck())
			require.NoError(t, err)

			var issuerMetadata issuer.Metadata

			err = json.Unmarshal(sampleIssuerMetadata, &issuerMetadata)
			require.NoError(t, err)

			resolvedDisplayData, err := credentialschema.Resolve(
				credentialschema.WithCredentials([]*verifiable.Credential{credential}),
				credentialschema.WithIssuerMetadata(&issuerMetadata))
			require.EqualError(t, err, "only VCs with one credential subject are supported")
			require.Nil(t, resolvedDisplayData)
		})
	})
}

func checkSuccessCaseMatchedDisplayData(t *testing.T, resolvedDisplayData *credentialschema.ResolvedDisplayData) {
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
	require.Empty(t, resolvedDisplayData.CredentialDisplays[0].Overview.Logo.AltText)
	require.Equal(t, "#12107c", resolvedDisplayData.CredentialDisplays[0].Overview.BackgroundColor)
	require.Equal(t, "#FFFFFF", resolvedDisplayData.CredentialDisplays[0].Overview.TextColor)

	expectedClaims := []credentialschema.ResolvedClaim{
		{Label: "ID", Value: "1234", Locale: "en-US"},
		{Label: "Given Name", Value: "Alice", Locale: "en-US"},
		{Label: "Surname", Value: "Bowman", Locale: "en-US"},
		{Label: "GPA", Value: "4.0", Locale: "en-US"},
	}

	verifyClaimsAnyOrder(t, resolvedDisplayData.CredentialDisplays[0].Claims, expectedClaims)
}

func verifyClaimsAnyOrder(t *testing.T, actualClaims []credentialschema.ResolvedClaim,
	expectedClaims []credentialschema.ResolvedClaim,
) {
	t.Helper()

	require.Len(t, actualClaims, len(expectedClaims), "expected (in any order): [%v]", expectedClaims)

	claimsMatched := make([]bool, len(expectedClaims))

	for _, actualClaim := range actualClaims {
		for j, expectedClaim := range expectedClaims {
			if claimsMatched[j] {
				continue
			}

			if actualClaim.Label == expectedClaim.Label &&
				actualClaim.Value == expectedClaim.Value &&
				actualClaim.Locale == expectedClaim.Locale {
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
