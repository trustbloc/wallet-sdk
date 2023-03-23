/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/stretchr/testify/require"

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

// Run this lines to make test work locally
// echo '127.0.0.1 testnet.orb.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 file-server.trustbloc.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 did-resolver.trustbloc.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 oidc-provider.example.com' | sudo tee -a /etc/hosts
// echo '127.0.0.1 vc-rest-echo.trustbloc.local' | sudo tee -a /etc/hosts

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

func TestOpenID4CIFullFlow(t *testing.T) {
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
	}

	type test struct {
		issuerProfileID     string
		issuerDIDMethod     string
		walletDIDMethod     string
		walletKeyType       string
		expectedIssuerURI   string
		expectedDisplayData *display.Data
		claimData           map[string]interface{}
	}

	tests := []test{
		{
			issuerProfileID:     "bank_issuer_jwtsd",
			issuerDIDMethod:     "orb",
			walletDIDMethod:     "ion",
			expectedIssuerURI:   "http://localhost:8075/issuer/bank_issuer_jwtsd",
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
			expectedIssuerURI:   "http://localhost:8075/issuer/bank_issuer",
		},
		{
			issuerProfileID:     "did_ion_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "key",
			claimData:           verifiableEmployeeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataDIDION),
			expectedIssuerURI:   "http://localhost:8075/issuer/did_ion_issuer",
		},
		{
			issuerProfileID:     "drivers_license_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "ion",
			claimData:           driverLicenseClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedDisplayDataDriversLicenseIssuer),
			expectedIssuerURI:   "http://localhost:8075/issuer/drivers_license_issuer",
		},
		{
			issuerProfileID:     "university_degree_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "ion",
			claimData:           universityDegreeClaims,
			expectedDisplayData: helpers.ParseDisplayData(t, expectedUniversityDegreeIssuer),
			expectedIssuerURI:   "http://localhost:8075/issuer/university_degree_issuer",
		},
	}

	oidc4ciSetup, err := oidc4ci.NewSetup(testenv.NewHttpRequest())
	require.NoError(t, err)

	err = oidc4ciSetup.AuthorizeIssuerBypassAuth("test_org", vcsAPIDirectURL)
	require.NoError(t, err)

	vcStatusVerifier, err := credential.NewStatusVerifier()
	require.NoError(t, err)

	for _, tc := range tests {
		fmt.Println(fmt.Sprintf("running tests with issuerProfileID=%s issuerDIDMethod=%s walletDIDMethod=%s",
			tc.issuerProfileID, tc.issuerDIDMethod, tc.walletDIDMethod))

		offerCredentialURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance(tc.issuerProfileID, tc.claimData)
		require.NoError(t, err)

		println(offerCredentialURL)

		testHelper := helpers.NewCITestHelper(t, tc.walletDIDMethod, tc.walletKeyType)

		didResolver, err := did.NewResolver(didResolverURL)
		require.NoError(t, err)

		didID, err := testHelper.DIDDoc.ID()
		require.NoError(t, err)

		clientConfig := openid4ci.ClientConfig{
			ClientID:       "ClientID",
			DIDResolver:    didResolver,
			ActivityLogger: testHelper.ActivityLogger,
			Crypto:         testHelper.KMS.GetCrypto(),
			MetricsLogger:  testHelper.MetricsLogger,
		}

		clientConfig.SetDocumentLoader(&documentLoaderReverseWrapper{DocumentLoader: testutil.DocumentLoader(t)})

		interaction, err := openid4ci.NewInteraction(offerCredentialURL, &clientConfig)
		require.NoError(t, err)

		authorizeResult, err := interaction.Authorize()
		require.NoError(t, err)
		require.False(t, authorizeResult.UserPINRequired)

		vm, err := testHelper.DIDDoc.AssertionMethod()
		require.NoError(t, err)

		credentials, err := interaction.RequestCredential(openid4ci.NewCredentialRequestOpts(""), vm)
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

		testHelper.CheckActivityLogAfterOpenID4CIFlow(t, vcsAPIDirectURL, tc.issuerProfileID, subID)
		testHelper.CheckMetricsLoggerAfterOpenID4CIFlow(t, tc.issuerProfileID)
	}
}
