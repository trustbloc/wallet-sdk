/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/attestation"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4vp"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/trustregistry"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/helpers"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/metricslogger"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4vp"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/testenv"
)

type claimData = map[string]interface{}

//go:embed expecteddisplaydata/university_degree_sd.json
var expectedUniversityDegreeSD string

func TestOpenID4VPFullFlow(t *testing.T) {
	trustRegistryAPI := trustregistry.NewRegistry(&trustregistry.RegistryConfig{
		EvaluateIssuanceURL:        "https://localhost:8100/wallet/interactions/issuance",
		EvaluatePresentationURL:    "https://localhost:8100/wallet/interactions/presentation",
		DisableHTTPClientTLSVerify: true,
	})

	driverLicenseClaims := claimData{
		"birthdate":            "1990-01-01",
		"document_number":      "123-456-789",
		"driving_privileges":   "G2",
		"expiry_date":          "2025-05-26",
		"family_name":          "Smith",
		"given_name":           "John",
		"issue_date":           "2020-05-27",
		"issuing_authority":    "Ministry of Transport Ontario",
		"issuing_country":      "Canada",
		"resident_address":     "4726 Pine Street",
		"resident_city":        "Toronto",
		"resident_postal_code": "A1B 2C3",
		"resident_province":    "Ontario",
	}

	verifiableEmployeeClaims := claimData{
		"displayName":       "John Doe",
		"givenName":         "John",
		"jobTitle":          "Software Developer",
		"surname":           "Doe",
		"preferredLanguage": "English",
		"mail":              "john.doe@foo.bar",
		"photo":             "data-URL-encoded image",
	}

	universityDegreeClaims := map[string]interface{}{
		"familyName":   "John Doe",
		"givenName":    "John",
		"degree":       "MIT",
		"degreeSchool": "MIT school",
	}

	type customScope struct {
		name         string
		customClaims string
	}

	type test struct {
		issuerProfileIDs     []string
		claimData            []claimData
		walletDIDMethod      string
		verifierProfileID    string
		signingKeyType       string
		matchedDisplayData   *display.Data
		customScopes         []customScope
		trustInfo            bool
		shouldBeForbidden    bool
		acknowledgeNoConsent bool
	}

	tests := []test{
		{
			issuerProfileIDs:   []string{"university_degree_issuer_bbs"},
			claimData:          []claimData{universityDegreeClaims},
			walletDIDMethod:    "ion",
			verifierProfileID:  "v_ldp_university_degree_sd_bbs",
			matchedDisplayData: helpers.ParseDisplayData(t, expectedUniversityDegreeSD),
			signingKeyType:     "ECDSAP256IEEEP1363", // Will result in a DI proof being added to the presentation
			trustInfo:          true,
		},
		{
			issuerProfileIDs:  []string{"university_degree_issuer"},
			claimData:         []claimData{universityDegreeClaims},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_ldp_university_degree",
		},
		{
			issuerProfileIDs:  []string{"university_degree_issuer_jwt"},
			claimData:         []claimData{universityDegreeClaims},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_jwt_university_degree",
		},
		{
			issuerProfileIDs:  []string{"bank_issuer"},
			claimData:         []claimData{verifiableEmployeeClaims},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_myprofile_jwt_verified_employee",
			trustInfo:         true,
			shouldBeForbidden: true,
		},
		{
			issuerProfileIDs:  []string{"bank_issuer"},
			claimData:         []claimData{verifiableEmployeeClaims},
			walletDIDMethod:   "key",
			verifierProfileID: "v_myprofile_jwt_verified_employee",
		},
		{
			issuerProfileIDs:  []string{"bank_issuer"},
			claimData:         []claimData{verifiableEmployeeClaims},
			walletDIDMethod:   "jwk",
			verifierProfileID: "v_myprofile_jwt_verified_employee",
		},
		{
			issuerProfileIDs:  []string{"bank_issuer_jwtsd"},
			claimData:         []claimData{verifiableEmployeeClaims},
			walletDIDMethod:   "jwk",
			verifierProfileID: "v_myprofile_sdjwt",
			signingKeyType:    localkms.KeyTypeP384,
		},
		{
			issuerProfileIDs:  []string{"drivers_license_issuer"},
			claimData:         []claimData{driverLicenseClaims},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_myprofile_jwt_attestation",
			trustInfo:         true,
		},
		{
			issuerProfileIDs:  []string{"bank_issuer", "drivers_license_issuer"},
			claimData:         []claimData{verifiableEmployeeClaims, driverLicenseClaims},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_myprofile_jwt_verified_employee",
		},
		{
			issuerProfileIDs:  []string{"bank_issuer", "bank_issuer"},
			claimData:         []claimData{verifiableEmployeeClaims, verifiableEmployeeClaims},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_myprofile_jwt_verified_employee",
			customScopes: []customScope{
				{
					name:         "registration",
					customClaims: `{"email": "test@example.com"}`,
				},
				{
					name: "profile",
					customClaims: `{
								"name": "Json chow",
								"email":"profile@example.com"
								}`,
				},
			},
		},
		{
			issuerProfileIDs:     []string{"bank_issuer"},
			claimData:            []claimData{verifiableEmployeeClaims},
			walletDIDMethod:      "ion",
			verifierProfileID:    "v_myprofile_jwt_verified_employee",
			acknowledgeNoConsent: true,
		},
		{
			issuerProfileIDs:  []string{"university_degree_issuer_v2"},
			claimData:         []claimData{universityDegreeClaims},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_ldp_university_degree",
		},
		{
			issuerProfileIDs:  []string{"university_degree_issuer_v2"},
			claimData:         []claimData{universityDegreeClaims},
			walletDIDMethod:   "jwk",
			verifierProfileID: "v_ldp_university_degree_v2",
		},
	}

	var traceIDs []string

	for i, tc := range tests {
		fmt.Printf("running test %d: issuerProfileIDs=%s verifierProfileID=%s "+
			"walletDIDMethod=%s\n", i,
			tc.issuerProfileIDs, tc.verifierProfileID, tc.walletDIDMethod)

		testHelper := helpers.NewVPTestHelper(t, tc.walletDIDMethod, tc.signingKeyType)

		issuedCredentials, issuersInfo := testHelper.IssueCredentials(t,
			vcsAPIDirectURL,
			tc.issuerProfileIDs,
			tc.claimData, nil)
		println("Issued", issuedCredentials.Length(), "credentials")
		for k := 0; k < issuedCredentials.Length(); k++ {
			cred, _ := issuedCredentials.AtIndex(k).Serialize()
			println("Issued VC[", k, "]: ", cred)
		}

		setup := oidc4vp.NewSetup(testenv.NewHttpRequest())

		err := setup.AuthorizeVerifierBypassAuth("f13d1va9lp403pb9lyj89vk55", vcsAPIDirectURL)
		require.NoError(t, err)

		var customScopes []string

		for _, scope := range tc.customScopes {
			customScopes = append(customScopes, scope.name)
		}

		initiateURL, err := setup.InitiateInteraction(tc.verifierProfileID, "test purpose.", customScopes)
		require.NoError(t, err)

		opts := did.NewResolverOpts()
		opts.SetResolverServerURI(didResolverURL)

		didResolver, err := did.NewResolver(opts)
		require.NoError(t, err)

		activityLogger := mem.NewActivityLogger()

		metricsLogger := metricslogger.NewMetricsLogger()

		interactionRequiredArgs := openid4vp.NewArgs(initiateURL, testHelper.KMS.GetCrypto(), didResolver)

		interactionOptionalArgs := openid4vp.NewOpts()

		interactionOptionalArgs.SetActivityLogger(activityLogger)
		interactionOptionalArgs.SetMetricsLogger(metricsLogger)
		interactionOptionalArgs.DisableHTTPClientTLSVerify()
		// DI proofs are only supported for certain key types, so if we're not using a compatible one then we'll need to
		// skip adding one.
		if tc.signingKeyType == "ECDSAP256IEEEP1363" {
			interactionOptionalArgs.EnableAddingDIProofs(testHelper.KMS)
		}

		interaction, err := openid4vp.NewInteraction(interactionRequiredArgs, interactionOptionalArgs)
		require.NoError(t, err)

		traceIDs = append(traceIDs, interaction.OTelTraceID())

		query, err := interaction.GetQuery()
		require.NoError(t, err)
		println("query", string(query))

		displayData := interaction.VerifierDisplayData()
		require.NoError(t, err)
		require.NotEmpty(t, displayData.DID)
		require.Equal(t, tc.verifierProfileID, displayData.Name())
		require.Equal(t, "test purpose.", displayData.Purpose())
		require.Equal(t, "", displayData.LogoURI())

		inquirerOpts := credential.NewInquirerOpts().SetDIDResolver(didResolver)

		inquirer, err := credential.NewInquirer(inquirerOpts)
		require.NoError(t, err)

		requirements, err := inquirer.GetSubmissionRequirements(query, issuedCredentials)
		require.NoError(t, err)
		require.GreaterOrEqual(t, requirements.Len(), 1)
		require.GreaterOrEqual(t, requirements.AtIndex(0).DescriptorLen(), 1)

		requirementDescriptor := requirements.AtIndex(0).DescriptorAtIndex(0)
		require.GreaterOrEqual(t, requirementDescriptor.MatchedVCs.Length(), 1)

		matchedVCs := requirementDescriptor.MatchedVCs

		if tc.matchedDisplayData != nil {
			vc := matchedVCs.AtIndex(0)
			issuer := issuersInfo[vc.ID()]
			helpers.ResolveDisplayData(t, toCredArray(vc), tc.matchedDisplayData, issuer.IssuerURI, issuer.ProfileID,
				didResolver)
		}

		requestedAcknowledgment := interaction.Acknowledgment()

		requestedAcknowledgmentData, err := requestedAcknowledgment.Serialize()
		require.NoError(t, err)

		requestedAcknowledgmentRestored, err := openid4vp.NewAcknowledgment(requestedAcknowledgmentData)
		require.NotNil(t, requestedAcknowledgmentRestored)

		if tc.acknowledgeNoConsent {
			require.NoError(t, requestedAcknowledgmentRestored.NoConsent())
			continue
		}

		selectedCreds := verifiable.NewCredentialsArray()
		for ind := 0; ind < matchedVCs.Length(); ind++ {
			vcID := matchedVCs.AtIndex(ind).ID()

			for j := 0; j < issuedCredentials.Length(); j++ {
				if issuedCredentials.AtIndex(j).ID() == vcID {
					selectedCreds.Add(issuedCredentials.AtIndex(ind))
				}
			}
		}

		presentOps := openid4vp.NewPresentCredentialOpts()

		if tc.trustInfo {
			info, trustErr := interaction.TrustInfo()
			require.NoError(t, trustErr)
			require.NotNil(t, info)
			require.Contains(t, info.Domain, "vcs.webhook.example.com")
			presTrustInfo := &trustregistry.PresentationRequest{
				VerifierDID:    info.DID,
				VerifierDomain: info.Domain,
			}

			for ind := 0; ind < selectedCreds.Length(); ind++ {
				cred := selectedCreds.AtIndex(ind)

				claims, err2 := interaction.PresentedClaims(cred)
				require.NoError(t, err2)

				presTrustInfo.AddCredentialClaims(&trustregistry.CredentialClaimsToCheck{
					CredentialID:        cred.ID(),
					CredentialTypes:     cred.Types(),
					IssuerID:            cred.IssuerID(),
					CredentialClaimKeys: claims,
				})
			}

			result, trustErr := trustRegistryAPI.EvaluatePresentation(presTrustInfo)
			require.NoError(t, trustErr)
			require.Equal(t, !tc.shouldBeForbidden, result.Allowed, result.ErrorMessage)

			for i := 0; i < result.RequestedAttestationLength(); i++ {
				if result.RequestedAttestationAtIndex(i) == "wallet_authentication" {
					vm, err := testHelper.DIDDoc.AssertionMethod()
					require.NoError(t, err)

					attClient, err := attestation.NewClient(
						attestation.NewCreateClientArgs(attestationURL, testHelper.KMS.GetCrypto()).
							DisableHTTPClientTLSVerify().AddHeader(&api.Header{
							Name:  "Authorization",
							Value: "Bearer token",
						}))
					require.NoError(t, err)

					attestationVC, err := attClient.GetAttestationVC(vm, `{
							"type": "urn:attestation:application:trustbloc",
							"application": {
								"type":    "wallet-cli",
								"name":    "wallet-cli",
								"version": "1.0"
							},
							"compliance": {
								"type": "fcra"				
							}
						}`,
					)
					require.NoError(t, err)

					attestationVCString, err := attestationVC.Serialize()
					require.NoError(t, err)

					presentOps.SetAttestationVC(vm, attestationVCString)

					require.NoError(t, err)
					require.NotNil(t, attestationVC)

					println("attestationVC=", attestationVC)
				}
			}
		}

		serializedIssuedVC, err := issuedCredentials.AtIndex(0).Serialize()
		require.NoError(t, err)

		serializedMatchedVC, err := selectedCreds.AtIndex(0).Serialize()
		require.NoError(t, err)
		println(serializedMatchedVC)

		require.Equal(t, serializedIssuedVC, serializedMatchedVC)

		for _, scope := range tc.customScopes {
			presentOps.AddScopeClaim(scope.name, scope.customClaims)
		}

		presentOps.SetInteractionDetails(fmt.Sprintf(`{"profile": %q}`, tc.verifierProfileID))

		err = interaction.PresentCredentialOpts(selectedCreds, presentOps)
		require.NoError(t, err)

		testHelper.CheckActivityLogAfterOpenID4VPFlow(t, activityLogger, tc.verifierProfileID)
		testHelper.CheckMetricsLoggerAfterOpenID4VPFlow(t, metricsLogger)

		fmt.Printf("done test %d\n", i)
	}

	require.Len(t, traceIDs, len(tests))

	time.Sleep(5 * time.Second)
	for _, traceID := range traceIDs {
		_, err := testenv.NewHttpRequest().Send(http.MethodGet,
			queryTraceURL+traceID,
			"",
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)
	}
}

func toCredArray(cred *verifiable.Credential) *verifiable.CredentialsArray {
	credsArr := verifiable.NewCredentialsArray()
	credsArr.Add(cred)
	return credsArr
}
