/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/attestation"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/oauth2"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/trustregistry"
	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/helpers"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4ci"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/testenv"
)

// Run these lines to make tests work locally
// echo '127.0.0.1 testnet.orb.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 file-server.trustbloc.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 did-resolver.trustbloc.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 vc-rest-echo.trustbloc.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 api-gateway.trustbloc.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 cognito-mock.trustbloc.local' | sudo tee -a /etc/hosts

var (
	//go:embed expecteddisplaydata/bank_issuer.json
	expectedDisplayDataBankIssuer string

	//go:embed expecteddisplaydata/did_ion_issuer.json
	expectedDisplayDataDIDION string

	//go:embed expecteddisplaydata/drivers_license_issuer.json
	expectedDisplayDataDriversLicenseIssuer string

	//go:embed expecteddisplaydata/university_degree_issuer.json
	expectedUniversityDegreeIssuer string
)

const (
	queryTraceURL = "http://localhost:16686/api/traces/"

	organizationID = "f13d1va9lp403pb9lyj89vk55"
)

type test struct {
	issuerProfileID     string
	issuerDIDMethod     string
	walletDIDMethod     string
	walletKeyType       string
	expectedIssuerURI   string
	expectedDisplayData *display.Data
	claimData           map[string]interface{}
	acknowledgeReject   bool
	trustInfo           bool
}

func TestOpenID4CIFullFlow(t *testing.T) {
	println("!!! Ensure the test certificate is imported into your keychain, and make sure the following" +
		" entries are in your hosts file:")
	println(`127.0.0.1 file-server.trustbloc.local
          127.0.0.1 did-resolver.trustbloc.local
          127.0.0.1 vc-rest-echo.trustbloc.local
          127.0.0.1 api-gateway.trustbloc.local
          127.0.0.1 cognito-mock.trustbloc.local`)

	println("Beginning pre-auth code flow tests.")
	doPreAuthCodeFlowTest(t)
	println("Completed pre-auth code flow tests.")
	println("Beginning auth code flow test.")
	doAuthCodeFlowTest(t, false)
	println("Completed auth code flow test.")
	println("Beginning auth code flow with dynamic client registration test.")
	doAuthCodeFlowTest(t, true)
	println("Completed auth code flow with dynamic client registration test.")
}

func doPreAuthCodeFlowTest(t *testing.T) {
	trustRegistryAPI := trustregistry.NewRegistry(&trustregistry.RegistryConfig{
		EvaluateIssuanceURL:        "https://localhost:8100/wallet/interactions/issuance",
		EvaluatePresentationURL:    "https://localhost:8100/wallet/interactions/presentation",
		DisableHTTPClientTLSVerify: true,
	})

	driverLicenseClaims := map[string]interface{}{
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

	verifiableEmployeeClaims := map[string]interface{}{
		"displayName":       "John Doe",
		"givenName":         "John",
		"jobTitle":          "Software Developer",
		"surname":           "Doe",
		"preferredLanguage": "English",
		"mail":              "john.doe@foo.bar",
		"photo":             "data-URL-encoded image",
		"sensitiveID":       "123456789",
		"reallySensitiveID": "abcdefg",
	}

	universityDegreeClaims := map[string]interface{}{
		"familyName":   "John Doe",
		"givenName":    "John",
		"degree":       "MIT",
		"degreeSchool": "MIT school",
		"photo":        "binary data",
	}

	preAuthTests := []test{
		{
			issuerProfileID:     "university_degree_issuer_bbs",
			issuerDIDMethod:     "key",
			walletDIDMethod:     "ion",
			claimData:           universityDegreeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedUniversityDegreeIssuer),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/university_degree_issuer_bbs/v1.0",
		},
		{
			issuerProfileID:     "bank_issuer_jwtsd",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "jwk",
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/bank_issuer_jwtsd/v1.0",
			claimData:           verifiableEmployeeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataBankIssuer),
			walletKeyType:       localkms.KeyTypeP384,
		},
		{
			issuerProfileID:     "bank_issuer_attest",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "ion",
			claimData:           verifiableEmployeeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataBankIssuer),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/bank_issuer_attest/v1.0",
			acknowledgeReject:   true,
			trustInfo:           true,
		},
		{
			issuerProfileID:     "did_ion_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "key",
			claimData:           verifiableEmployeeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataDIDION),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/did_ion_issuer/v1.0",
			acknowledgeReject:   true,
		},
		{
			issuerProfileID:     "drivers_license_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "ion",
			claimData:           driverLicenseClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataDriversLicenseIssuer),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/drivers_license_issuer/v1.0",
		},
		{
			issuerProfileID:     "university_degree_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "ion",
			claimData:           universityDegreeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedUniversityDegreeIssuer),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/university_degree_issuer/v1.0",
		},
	}

	oidc4ciSetup, err := oidc4ci.NewSetup(testenv.NewHttpRequest())
	require.NoError(t, err)

	err = oidc4ciSetup.AuthorizeIssuerBypassAuth(organizationID, vcsAPIDirectURL)
	require.NoError(t, err)

	vcStatusVerifier, err := credential.NewStatusVerifier(credential.NewStatusVerifierOpts())
	require.NoError(t, err)

	var traceIDs []string

	for _, tc := range preAuthTests {
		fmt.Println(fmt.Sprintf("running tests with issuerProfileID=%s issuerDIDMethod=%s walletDIDMethod=%s",
			tc.issuerProfileID, tc.issuerDIDMethod, tc.walletDIDMethod))

		offerCredentialURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance(tc.issuerProfileID, tc.claimData)
		require.NoError(t, err)

		println(offerCredentialURL)

		testHelper := helpers.NewCITestHelper(t, tc.walletDIDMethod, tc.walletKeyType)

		opts := did.NewResolverOpts()
		opts.SetResolverServerURI(didResolverURL)

		didResolver, err := did.NewResolver(opts)
		require.NoError(t, err)

		didID, err := testHelper.DIDDoc.ID()
		require.NoError(t, err)

		interactionRequiredArgs := openid4ci.NewIssuerInitiatedInteractionArgs(offerCredentialURL, testHelper.KMS.GetCrypto(), didResolver)

		interactionOptionalArgs := openid4ci.NewInteractionOpts()
		interactionOptionalArgs.SetDocumentLoader(&documentLoaderReverseWrapper{DocumentLoader: testutil.DocumentLoader(t)})
		interactionOptionalArgs.SetActivityLogger(testHelper.ActivityLogger)
		interactionOptionalArgs.SetMetricsLogger(testHelper.MetricsLogger)
		interactionOptionalArgs.DisableHTTPClientTLSVerify()
		interactionOptionalArgs.EnableDIProofChecks(testHelper.KMS)

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(interactionRequiredArgs, interactionOptionalArgs)
		require.NoError(t, err)

		traceIDs = append(traceIDs, interaction.OTelTraceID())

		require.True(t, interaction.PreAuthorizedCodeGrantTypeSupported())

		preAuthorizedCodeGrantParams, err := interaction.PreAuthorizedCodeGrantParams()
		require.NoError(t, err)

		require.False(t, preAuthorizedCodeGrantParams.PINRequired())

		issuerMetadata, err := interaction.IssuerMetadata()
		require.NoError(t, err)

		offeringDisplayData := display.ResolveCredentialOffer(
			issuerMetadata,
			interaction.OfferedCredentialsTypes(),
			"",
		)

		helpers.CheckResolvedDisplayData(t, offeringDisplayData, tc.expectedDisplayData, false)

		vm, err := testHelper.DIDDoc.AssertionMethod()
		require.NoError(t, err)

		opt := openid4ci.NewRequestCredentialWithPreAuthOpts()

		if tc.trustInfo {
			trustInfo, trErr := interaction.IssuerTrustInfo()
			require.NoError(t, trErr)
			require.NotNil(t, trustInfo)

			require.Contains(t, trustInfo.Domain, "trustbloc.local:8078")

			req := &trustregistry.IssuanceRequest{
				IssuerDID:    trustInfo.DID,
				IssuerDomain: trustInfo.Domain,
			}

			for _, offer := range trustInfo.CredentialOffers {
				req.AddCredentialOffers(&trustregistry.CredentialOffer{
					CredentialType:             offer.CredentialType,
					CredentialFormat:           offer.CredentialFormat,
					ClientAttestationRequested: false,
				})
			}

			result, trustErr := trustRegistryAPI.EvaluateIssuance(req)

			require.NoError(t, trustErr)
			require.Equal(t, true, result.Allowed)

			for i := 0; i < result.RequestedAttestationLength(); i++ {
				if result.RequestedAttestationAtIndex(i) == "wallet_authentication" {
					attClient, err := attestation.NewClient(
						attestation.NewCreateClientArgs(attestationURL, testHelper.KMS.GetCrypto()).
							DisableHTTPClientTLSVerify())
					require.NoError(t, err)

					attestationVC, err := attClient.GetAttestationVC(vm, attestation.NewAttestRequest().
						AddAssertion("wallet_authentication").
						AddWalletAuthentication("wallet_id", didID).
						AddWalletMetadata("wallet_name", "int-test"),
					)
					require.NoError(t, err)

					attestationVCString, err := attestationVC.Serialize()
					require.NoError(t, err)

					opt.SetAttestationVC(vm, attestationVCString)

					require.NoError(t, err)
					require.NotNil(t, attestationVC)

					println("attestationVC=", attestationVC)
				}
			}
		}

		credentials, err := interaction.RequestCredentialWithPreAuth(vm, opt)
		require.NoError(t, err)
		require.NotNil(t, credentials)

		requestedAcknowledgment, err := interaction.Acknowledgment()
		require.NotNil(t, requestedAcknowledgment)

		requestedAcknowledgmentData, err := requestedAcknowledgment.Serialize()
		require.NoError(t, err)

		requestedAcknowledgmentRestored, err := openid4ci.NewAcknowledgment(requestedAcknowledgmentData)
		require.NotNil(t, requestedAcknowledgmentRestored)

		if tc.acknowledgeReject {
			require.NoError(t, requestedAcknowledgmentRestored.Reject())
		} else {
			require.NoError(t, requestedAcknowledgmentRestored.Success())
		}

		vc := credentials.AtIndex(0)

		serializedVC, err := vc.Serialize()
		require.NoError(t, err)

		println("credential:", serializedVC)
		require.NoError(t, err)
		require.Contains(t, vc.VC.Contents().Issuer.ID, tc.issuerDIDMethod)

		helpers.ResolveDisplayData(t, credentials, tc.expectedDisplayData, interaction.IssuerURI(), tc.issuerProfileID,
			didResolver)

		issuerURI := interaction.IssuerURI()
		require.Equal(t, tc.expectedIssuerURI, issuerURI)

		subID, err := verifiable.SubjectID(vc.VC.Contents().Subject)
		require.NoError(t, err)
		require.Contains(t, subID, didID)

		require.NoError(t, vcStatusVerifier.Verify(vc))

		testHelper.CheckActivityLogAfterOpenID4CIFlow(t, vcsAPIDirectURL,
			tc.issuerProfileID, subID)
	}

	require.Len(t, traceIDs, len(preAuthTests))

	time.Sleep(5 * time.Second)
	for _, traceID := range traceIDs {
		_, err = testenv.NewHttpRequest().Send(http.MethodGet,
			queryTraceURL+traceID,
			"",
			nil,
			nil,
			nil,
		)
		require.NoError(t, err)
	}
}

func doAuthCodeFlowTest(t *testing.T, useDynamicClientRegistration bool) {
	credentialOfferURL, err := oidc4ci.InitiateAuthCodeIssuance()
	require.NoError(t, err)

	testHelper := helpers.NewCITestHelper(t, "ion", "")

	opts := did.NewResolverOpts()
	opts.SetResolverServerURI(didResolverURL)

	didResolver, err := did.NewResolver(opts)
	require.NoError(t, err)

	interactionRequiredArgs := openid4ci.NewIssuerInitiatedInteractionArgs(credentialOfferURL, testHelper.KMS.GetCrypto(), didResolver)
	interactionOptionalArgs := openid4ci.NewInteractionOpts().DisableHTTPClientTLSVerify()

	interaction, err := openid4ci.NewIssuerInitiatedInteraction(interactionRequiredArgs, interactionOptionalArgs)
	require.NoError(t, err)

	// If dynamic client registration is used, then the client ID and scopes below will be overwritten by the
	// values from the registration response.
	clientID := "oidc4vc_client"

	scopes := api.NewStringArray()
	scopes.Append("openid")
	scopes.Append("profile")

	if useDynamicClientRegistration {
		supported, err := interaction.DynamicClientRegistrationSupported()
		require.NoError(t, err)
		require.True(t, supported)

		registrationEndpoint, err := interaction.DynamicClientRegistrationEndpoint()
		require.NoError(t, err)
		require.NotEmpty(t, registrationEndpoint)

		clientMetadata := oauth2.NewClientMetadata()

		grantTypes := api.NewStringArray().Append("authorization_code")
		clientMetadata.SetGrantTypes(grantTypes)

		redirectURIs := api.NewStringArray().Append("http://127.0.0.1/callback")
		clientMetadata.SetRedirectURIs(redirectURIs)

		clientMetadata.SetScopes(scopes)

		clientMetadata.SetTokenEndpointAuthMethod("none")

		authorizationCodeGrantParams, err := interaction.AuthorizationCodeGrantParams()
		require.NoError(t, err)

		if authorizationCodeGrantParams.HasIssuerState() {
			issuerState, err := authorizationCodeGrantParams.IssuerState()
			require.NoError(t, err)

			clientMetadata.SetIssuerState(issuerState)
		}

		registerClientResponse, err := oauth2.RegisterClient(registrationEndpoint, clientMetadata, nil)
		require.NoError(t, err)

		clientID = registerClientResponse.ClientID()

		registeredMetadata := registerClientResponse.RegisteredMetadata()

		// Use the actual scopes registered by the authorization server, which may differ from the scopes
		// we specified in the metadata in our request.
		scopes = registeredMetadata.Scopes()
	}

	redirectURIWithAuthCode := getRedirectURIWithAuthCode(t, clientID, interaction, scopes)

	vm, err := testHelper.DIDDoc.AssertionMethod()
	require.NoError(t, err)

	credentials, err := interaction.RequestCredentialWithAuth(vm, redirectURIWithAuthCode, nil)
	require.NoError(t, err)
	require.NotNil(t, credentials)
	require.Equal(t, 1, credentials.Length())

	requestedAcknowledgment, err := interaction.Acknowledgment()
	require.NotNil(t, requestedAcknowledgment)
	require.NoError(t, requestedAcknowledgment.Success())
}

func getRedirectURIWithAuthCode(t *testing.T, clientID string, interaction *openid4ci.IssuerInitiatedInteraction,
	scopes *api.StringArray,
) string {
	authURL, err := interaction.CreateAuthorizationURL(clientID,
		"http://127.0.0.1/callback", openid4ci.NewCreateAuthorizationURLOpts().SetScopes(scopes))
	require.NoError(t, err)

	var redirectURIWithAuthCode string

	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		}},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// intercept auth code
			if strings.HasPrefix(req.URL.String(), "http://127.0.0.1/callback") {
				redirectURIWithAuthCode = req.URL.String()

				return http.ErrUseLastResponse
			}

			return nil
		},
	}

	resp, err := httpClient.Get(authURL)
	require.NoError(t, err)
	require.Equal(t, http.StatusSeeOther, resp.StatusCode)

	return redirectURIWithAuthCode
}
