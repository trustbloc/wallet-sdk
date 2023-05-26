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

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
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
}

func TestOpenID4CIFullFlow(t *testing.T) {
	println("!!! Ensure the test certificate is imported into your keychain, and make sure the following" +
		"entries are in your hosts file:")
	println(`127.0.0.1 testnet.orb.local
          127.0.0.1 file-server.trustbloc.local
          127.0.0.1 did-resolver.trustbloc.local
          127.0.0.1 vc-rest-echo.trustbloc.local
          127.0.0.1 api-gateway.trustbloc.local
          127.0.0.1 cognito-mock.trustbloc.local`)

	println("Beginning pre-auth code flow tests.")
	doPreAuthCodeFlowTest(t)
	println("Completed pre-auth code flow tests.")
	println("Beginning auth code flow test.")
	doAuthCodeFlowTest(t)
	println("Completed auth code flow test.")
}

func doPreAuthCodeFlowTest(t *testing.T) {
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
			expectedIssuerURI:   "http://localhost:8075/issuer/university_degree_issuer_bbs/v1.0",
		},
		{
			issuerProfileID:     "bank_issuer_jwtsd",
			issuerDIDMethod:     "orb",
			walletDIDMethod:     "ion",
			expectedIssuerURI:   "http://localhost:8075/issuer/bank_issuer_jwtsd/v1.0",
			claimData:           verifiableEmployeeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataBankIssuer),
			walletKeyType:       localkms.KeyTypeP384,
		},
		{
			issuerProfileID:     "bank_issuer",
			issuerDIDMethod:     "orb",
			walletDIDMethod:     "ion",
			claimData:           verifiableEmployeeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataBankIssuer),
			expectedIssuerURI:   "http://localhost:8075/issuer/bank_issuer/v1.0",
		},
		{
			issuerProfileID:     "did_ion_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "key",
			claimData:           verifiableEmployeeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataDIDION),
			expectedIssuerURI:   "http://localhost:8075/issuer/did_ion_issuer/v1.0",
		},
		{
			issuerProfileID:     "drivers_license_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "ion",
			claimData:           driverLicenseClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataDriversLicenseIssuer),
			expectedIssuerURI:   "http://localhost:8075/issuer/drivers_license_issuer/v1.0",
		},
		{
			issuerProfileID:     "university_degree_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "ion",
			claimData:           universityDegreeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedUniversityDegreeIssuer),
			expectedIssuerURI:   "http://localhost:8075/issuer/university_degree_issuer/v1.0",
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

		interactionRequiredArgs := openid4ci.NewArgs(offerCredentialURL, testHelper.KMS.GetCrypto(), didResolver)

		interactionOptionalArgs := openid4ci.NewOpts()
		interactionOptionalArgs.SetDocumentLoader(&documentLoaderReverseWrapper{DocumentLoader: testutil.DocumentLoader(t)})
		interactionOptionalArgs.SetActivityLogger(testHelper.ActivityLogger)
		interactionOptionalArgs.SetMetricsLogger(testHelper.MetricsLogger)
		interactionOptionalArgs.DisableHTTPClientTLSVerify()

		interaction, err := openid4ci.NewInteraction(interactionRequiredArgs, interactionOptionalArgs)
		require.NoError(t, err)

		traceIDs = append(traceIDs, interaction.OTelTraceID())

		require.True(t, interaction.IssuerCapabilities().PreAuthorizedCodeGrantTypeSupported())

		preAuthorizedCodeGrantParams, err := interaction.IssuerCapabilities().PreAuthorizedCodeGrantParams()
		require.NoError(t, err)

		require.False(t, preAuthorizedCodeGrantParams.PINRequired())

		vm, err := testHelper.DIDDoc.AssertionMethod()
		require.NoError(t, err)

		credentials, err := interaction.RequestCredential(vm)
		require.NoError(t, err)
		require.NotNil(t, credentials)

		vc := credentials.AtIndex(0)

		serializedVC, err := vc.Serialize()
		require.NoError(t, err)

		println("credential:", serializedVC)
		require.NoError(t, err)
		require.Contains(t, vc.VC.Issuer.ID, tc.issuerDIDMethod)

		helpers.ResolveDisplayData(t, credentials, tc.expectedDisplayData, interaction.IssuerURI(), tc.issuerProfileID)

		issuerURI := interaction.IssuerURI()
		require.Equal(t, tc.expectedIssuerURI, issuerURI)

		subID, err := verifiable.SubjectID(vc.VC.Subject)
		require.NoError(t, err)
		require.Contains(t, subID, didID)

		require.NoError(t, vcStatusVerifier.Verify(vc))

		testHelper.CheckActivityLogAfterOpenID4CIFlow(t, vcsAPIDirectURL,
			tc.issuerProfileID, subID)
		testHelper.CheckMetricsLoggerAfterOpenID4CIFlow(t, tc.issuerProfileID)
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

func doAuthCodeFlowTest(t *testing.T) {
	credentialOfferURL := oidc4ci.InitiateAuthCodeIssuance(t)

	testHelper := helpers.NewCITestHelper(t, "ion", "")

	opts := did.NewResolverOpts()
	opts.SetResolverServerURI(didResolverURL)

	didResolver, err := did.NewResolver(opts)
	require.NoError(t, err)

	interactionRequiredArgs := openid4ci.NewArgs(credentialOfferURL, testHelper.KMS.GetCrypto(), didResolver)
	interactionOptionalArgs := openid4ci.NewOpts().DisableHTTPClientTLSVerify()

	interaction, err := openid4ci.NewInteraction(interactionRequiredArgs, interactionOptionalArgs)
	require.NoError(t, err)

	redirectURIWithAuthCode := getRedirectURIWithAuthCode(t, interaction)

	vm, err := testHelper.DIDDoc.AssertionMethod()
	require.NoError(t, err)

	credentials, err := interaction.RequestCredentialWithAuth(vm, redirectURIWithAuthCode)
	require.NoError(t, err)
	require.NotNil(t, credentials)
	require.Equal(t, 1, credentials.Length())
}

func getRedirectURIWithAuthCode(t *testing.T, interaction *openid4ci.Interaction) string {
	scopes := api.NewStringArray()
	scopes.Append("openid")
	scopes.Append("profile")

	authURL, err := interaction.CreateAuthorizationURLWithScopes("oidc4vc_client",
		"http://127.0.0.1/callback", scopes)
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
