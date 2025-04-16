/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

	//go:embed expecteddisplaydata/university_degree_multi.json
	expectedUniversityDegreeMulti string

	//go:embed expecteddisplaydata/university_degree_v2.json
	expectedUniversityDegreeV2 string
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
	claims              []*claimEntry
	acknowledgeReject   bool
	trustInfo           bool
	displayAPIv2        bool
}

type claimEntry struct {
	Data map[string]interface{}
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

	universityDegreeClaims2 := map[string]interface{}{
		"familyName":   "John Doe",
		"givenName":    "John",
		"degree":       "MS",
		"degreeSchool": "Stanford",
		"photo":        "binary data",
	}

	preAuthTests := []test{
		{
			issuerProfileID: "university_degree_issuer_bbs",
			issuerDIDMethod: "key",
			walletDIDMethod: "ion",
			claims: []*claimEntry{
				{
					Data: universityDegreeClaims,
				},
			},
			expectedDisplayData: helpers.ParseDisplayData(t, expectedUniversityDegreeIssuer),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/university_degree_issuer_bbs/v1.0",
		},
		{
			issuerProfileID:   "bank_issuer_jwtsd",
			issuerDIDMethod:   "ion",
			walletDIDMethod:   "jwk",
			expectedIssuerURI: "http://localhost:8075/oidc/idp/bank_issuer_jwtsd/v1.0",
			claims: []*claimEntry{
				{
					Data: verifiableEmployeeClaims,
				},
			},
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataBankIssuer),
			walletKeyType:       localkms.KeyTypeP384,
		},
		{
			issuerProfileID: "bank_issuer_attest",
			issuerDIDMethod: "ion",
			walletDIDMethod: "ion",
			claims: []*claimEntry{
				{
					Data: verifiableEmployeeClaims,
				},
			},
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataBankIssuer),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/bank_issuer_attest/v1.0",
			acknowledgeReject:   true,
			trustInfo:           true,
		},
		{
			issuerProfileID: "did_ion_issuer",
			issuerDIDMethod: "ion",
			walletDIDMethod: "key",
			claims: []*claimEntry{
				{
					Data: verifiableEmployeeClaims,
				},
			},
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataDIDION),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/did_ion_issuer/v1.0",
			acknowledgeReject:   true,
		},
		{
			issuerProfileID: "drivers_license_issuer",
			issuerDIDMethod: "ion",
			walletDIDMethod: "ion",
			claims: []*claimEntry{
				{
					Data: driverLicenseClaims,
				},
			},
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataDriversLicenseIssuer),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/drivers_license_issuer/v1.0",
		},
		{
			issuerProfileID: "university_degree_issuer",
			issuerDIDMethod: "ion",
			walletDIDMethod: "ion",
			claims: []*claimEntry{
				{
					Data: universityDegreeClaims,
				},
				{
					Data: universityDegreeClaims2,
				},
			},
			expectedDisplayData: helpers.ParseDisplayData(t, expectedUniversityDegreeMulti),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/university_degree_issuer/v1.0",
		},
		{
			issuerProfileID: "university_degree_issuer_v2",
			issuerDIDMethod: "ion",
			walletDIDMethod: "ion",
			claims: []*claimEntry{
				{
					Data: universityDegreeClaims,
				},
				{
					Data: universityDegreeClaims2,
				},
			},
			expectedDisplayData: helpers.ParseDisplayData(t, expectedUniversityDegreeV2),
			expectedIssuerURI:   "http://localhost:8075/oidc/idp/university_degree_issuer_v2/v1.0",
			displayAPIv2:        true,
		},
	}

	oidc4ciSetup, setupErr := oidc4ci.NewSetup(testenv.NewHttpRequest())
	require.NoError(t, setupErr)

	setupErr = oidc4ciSetup.AuthorizeIssuerBypassAuth(organizationID, vcsAPIDirectURL)
	require.NoError(t, setupErr)

	var traceIDs []string

	for _, tc := range preAuthTests {
		fmt.Printf("Running tests with issuerProfileID=%s issuerDIDMethod=%s walletDIDMethod=%s\n",
			tc.issuerProfileID, tc.issuerDIDMethod, tc.walletDIDMethod)

		credentialConfigs := make([]oidc4ci.CredentialConfiguration, 0)

		for _, c := range tc.claims {
			credentialConfigs = append(credentialConfigs,
				oidc4ci.CredentialConfiguration{
					ClaimData: c.Data,
				},
			)
		}

		offerCredentialURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance(tc.issuerProfileID, credentialConfigs)
		require.NoError(t, err)

		if tc.displayAPIv2 {
			// In the current implementation, VCS determines the configuration ID by matching the credential type.
			// To test Display API v2, we intentionally switch to the second configuration ID, which shares the same
			// credential type as the first configuration but has a different set of display fields.
			setCredentialConfigurationIDs(t, &offerCredentialURL,
				[]string{"UniversityDegreeCredential_ldp_vc_v2", "UniversityDegreeCredential_ldp_vc_v1"})
		}

		fmt.Printf("offerCredentialURL=%s\n", offerCredentialURL)

		testHelper := helpers.NewCITestHelper(t, tc.walletDIDMethod, tc.walletKeyType)

		resolverOpts := did.NewResolverOpts()
		resolverOpts.SetResolverServerURI(didResolverURL)

		didResolver, err := did.NewResolver(resolverOpts)
		require.NoError(t, err)

		didID, err := testHelper.DIDDoc.ID()
		require.NoError(t, err)

		interactionRequiredArgs := openid4ci.NewIssuerInitiatedInteractionArgs(offerCredentialURL,
			testHelper.KMS.GetCrypto(), didResolver)

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

		issuerURI := interaction.IssuerURI()
		require.Equal(t, tc.expectedIssuerURI, issuerURI)

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

			require.Contains(t, trustInfo.Domain, "trustbloc.local:8075")

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
							DisableHTTPClientTLSVerify().AddHeader(&api.Header{
							Name:  "Authorization",
							Value: "Bearer token",
						}))
					require.NoError(t, err)

					attestationVC, err := attClient.GetAttestationVC(vm,
						`{"type":"urn:attestation:application:midy","application":{"type":"MidyWallet","name":"Midy Wallet","version":"2.0"},"compliance":{"type":"fcra"}}`,
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

		var subjectID string

		if tc.displayAPIv2 {
			subjectID = requestCredentialWithPreAuthV2(t, interaction, vm, didResolver, opt, tc.issuerDIDMethod,
				tc.issuerProfileID, tc.expectedDisplayData)
		} else {
			subjectID = requestCredentialWithPreAuth(t, interaction, vm, didResolver, opt, tc.issuerDIDMethod,
				tc.issuerProfileID, tc.expectedDisplayData)
		}

		require.Contains(t, subjectID, didID)

		requestedAcknowledgment, err := interaction.Acknowledgment()
		require.NotNil(t, requestedAcknowledgment)
		require.NoError(t, err)

		requestedAcknowledgmentData, err := requestedAcknowledgment.Serialize()
		require.NoError(t, err)

		requestedAcknowledgmentRestored, err := openid4ci.NewAcknowledgment(requestedAcknowledgmentData)
		require.NotNil(t, requestedAcknowledgmentRestored)
		require.NoError(t, err)

		err = requestedAcknowledgmentRestored.SetInteractionDetails(fmt.Sprintf(`{"profile": %q}`, tc.issuerProfileID))
		require.NoError(t, err)

		if tc.acknowledgeReject {
			require.NoError(t, requestedAcknowledgmentRestored.Reject())
		} else {
			require.NoError(t, requestedAcknowledgmentRestored.Success())
		}

		testHelper.CheckActivityLogAfterOpenID4CIFlow(t, vcsAPIDirectURL,
			tc.issuerProfileID, subjectID)
	}

	require.Len(t, traceIDs, len(preAuthTests))

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

func setCredentialConfigurationIDs(t *testing.T, credentialOfferURL *string, configIDs []string) {
	t.Helper()

	u, err := url.Parse(*credentialOfferURL)
	require.NoError(t, err)

	offerParam := u.Query().Get("credential_offer")
	require.NotEmpty(t, offerParam)

	var offer map[string]interface{}
	err = json.Unmarshal([]byte(offerParam), &offer)
	require.NoError(t, err)

	offer["credential_configuration_ids"] = configIDs

	modifiedOffer, err := json.Marshal(offer)
	require.NoError(t, err)

	query := u.Query()
	query.Set("credential_offer", string(modifiedOffer))
	u.RawQuery = query.Encode()

	*credentialOfferURL = u.String()
}

func requestCredentialWithPreAuth(
	t *testing.T,
	interaction *openid4ci.IssuerInitiatedInteraction,
	vm *api.VerificationMethod,
	didResolver *did.Resolver,
	opts *openid4ci.RequestCredentialWithPreAuthOpts,
	issuerDIDMethod, issuerProfileID string,
	expectedDisplayData *display.Data,
) string {
	credentials, err := interaction.RequestCredentialWithPreAuth(vm, opts)
	require.NoError(t, err)
	require.NotNil(t, credentials)

	for i := 0; i < credentials.Length(); i++ {
		cred := credentials.AtIndex(i)

		require.NotEmpty(t, cred.ID())
		require.NotEmpty(t, cred.IssuerID())
		require.NotEmpty(t, cred.Types())
		require.True(t, cred.IssuanceDate() > 0)
		require.True(t, cred.ExpirationDate() > 0)
	}

	vc := credentials.AtIndex(0)

	serializedVC, err := vc.Serialize()
	require.NoError(t, err)

	fmt.Printf("credential: %s\n", serializedVC)

	require.NoError(t, err)
	require.Contains(t, vc.VC.Contents().Issuer.ID, issuerDIDMethod)

	helpers.ResolveDisplayData(t, credentials, expectedDisplayData, interaction.IssuerURI(), issuerProfileID,
		didResolver)

	subjectID, err := verifiable.SubjectID(vc.VC.Contents().Subject)
	require.NoError(t, err)

	statusVerifier, err := credential.NewStatusVerifierWithDIDResolver(didResolver, credential.NewStatusVerifierOpts())
	require.NoError(t, err)

	err = statusVerifier.Verify(vc)
	require.NoError(t, err)

	return subjectID
}

func requestCredentialWithPreAuthV2(
	t *testing.T,
	interaction *openid4ci.IssuerInitiatedInteraction,
	vm *api.VerificationMethod,
	didResolver *did.Resolver,
	opts *openid4ci.RequestCredentialWithPreAuthOpts,
	issuerDIDMethod, issuerProfileID string,
	expectedDisplayData *display.Data,
) string {
	credentials, err := interaction.RequestCredentialWithPreAuthV2(vm, opts)
	require.NoError(t, err)
	require.NotNil(t, credentials)

	for i := 0; i < credentials.Length(); i++ {
		cred := credentials.AtIndex(i)

		require.NotEmpty(t, cred.ID())
		require.NotEmpty(t, cred.IssuerID())
		require.NotEmpty(t, cred.Types())
		require.True(t, cred.IssuanceDate() > 0)
		require.True(t, cred.ExpirationDate() > 0)
	}

	vc := credentials.AtIndex(0)

	serializedVC, err := vc.Serialize()
	require.NoError(t, err)

	fmt.Printf("credential: %s\n", serializedVC)

	require.NoError(t, err)
	require.Contains(t, vc.VC.Contents().Issuer.ID, issuerDIDMethod)

	helpers.ResolveDisplayDataV2(t, credentials, expectedDisplayData, interaction.IssuerURI(), issuerProfileID,
		didResolver)

	subjectID, err := verifiable.SubjectID(vc.VC.Contents().Subject)
	require.NoError(t, err)

	statusVerifier, err := credential.NewStatusVerifierWithDIDResolver(didResolver, credential.NewStatusVerifierOpts())
	require.NoError(t, err)

	err = statusVerifier.Verify(vc)
	require.NoError(t, err)

	return subjectID
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
	require.NoError(t, err)
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
