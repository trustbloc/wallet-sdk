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

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
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
)

func TestOpenID4CIFullFlow(t *testing.T) {
	type test struct {
		issuerProfileID     string
		issuerDIDMethod     string
		walletDIDMethod     string
		expectedIssuerURI   string
		expectedDisplayData *openid4ci.DisplayData
	}

	tests := []test{
		{
			issuerProfileID:     "bank_issuer_jwtsd",
			issuerDIDMethod:     "orb",
			walletDIDMethod:     "ion",
			expectedIssuerURI:   "http://localhost:8075/issuer/bank_issuer_jwtsd",
			expectedDisplayData: parseDisplayData(t, expectedDisplayDataBankIssuer),
		},
		{
			issuerProfileID:     "bank_issuer",
			issuerDIDMethod:     "orb",
			walletDIDMethod:     "ion",
			expectedDisplayData: parseDisplayData(t, expectedDisplayDataBankIssuer),
			expectedIssuerURI:   "http://localhost:8075/issuer/bank_issuer",
		},
		{
			issuerProfileID:     "did_ion_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "key",
			expectedDisplayData: parseDisplayData(t, expectedDisplayDataDIDION),
			expectedIssuerURI:   "http://localhost:8075/issuer/did_ion_issuer",
		},
		{
			issuerProfileID:     "drivers_license_issuer",
			issuerDIDMethod:     "ion",
			walletDIDMethod:     "ion",
			expectedDisplayData: parseDisplayData(t, expectedDisplayDataDriversLicenseIssuer),
			expectedIssuerURI:   "http://localhost:8075/issuer/drivers_license_issuer",
		},
	}

	oidc4ciSetup, err := oidc4ci.NewSetup(testenv.NewHttpRequest())
	require.NoError(t, err)

	err = oidc4ciSetup.AuthorizeIssuerBypassAuth("test_org")
	require.NoError(t, err)

	for _, tc := range tests {
		fmt.Println(fmt.Sprintf("running tests with issuerProfileID=%s issuerDIDMethod=%s walletDIDMethod=%s",
			tc.issuerProfileID, tc.issuerDIDMethod, tc.walletDIDMethod))

		initiateIssuanceURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance(tc.issuerProfileID)
		require.NoError(t, err)

		println(initiateIssuanceURL)

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		// create DID
		c, err := did.NewCreatorWithKeyWriter(kms)
		require.NoError(t, err)

		didDoc, err := c.Create(tc.walletDIDMethod, &api.CreateDIDOpts{})
		require.NoError(t, err)

		didResolver, err := did.NewResolver("")
		require.NoError(t, err)

		didID, err := didDoc.ID()
		require.NoError(t, err)

		activityLogger := mem.NewActivityLogger()

		clientConfig := openid4ci.ClientConfig{
			ClientID:       "ClientID",
			DIDResolver:    didResolver,
			ActivityLogger: activityLogger,
			Crypto:         kms.GetCrypto(),
		}

		interaction, err := openid4ci.NewInteraction(initiateIssuanceURL, &clientConfig)
		require.NoError(t, err)

		authorizeResult, err := interaction.Authorize()
		require.NoError(t, err)
		require.False(t, authorizeResult.UserPINRequired)

		vm, err := didDoc.AssertionMethod()
		require.NoError(t, err)

		credential, err := interaction.RequestCredential(openid4ci.NewCredentialRequestOpts(""), vm)
		require.NoError(t, err)
		require.NotNil(t, credential)

		vc := credential.AtIndex(0)

		serializedVC, err := vc.Serialize()
		require.NoError(t, err)

		println("credential:", serializedVC)
		require.NoError(t, err)
		require.Contains(t, vc.VC.Issuer.ID, tc.issuerDIDMethod)

		displayData, err := interaction.ResolveDisplay("")
		require.NoError(t, err)
		checkResolvedDisplayData(t, displayData, tc.expectedDisplayData)

		issuerURI := interaction.IssuerURI()
		require.Equal(t, tc.expectedIssuerURI, issuerURI)

		subID, err := verifiable.SubjectID(vc.VC.Subject)
		require.NoError(t, err)
		require.Contains(t, subID, didID)
		checkActivityLogAfterOpenID4CIFlow(t, activityLogger, tc.issuerProfileID, subID)
	}
}

func parseDisplayData(t *testing.T, displayData string) *openid4ci.DisplayData {
	parsedDisplayData, err := openid4ci.ParseDisplayData(displayData)
	require.NoError(t, err)

	return parsedDisplayData
}

// For now, this function assumes that the display data object has only a single credential display.
// In the event we add a test case where there are multiple credential displays, then this function will need to be
// updated accordingly.
func checkResolvedDisplayData(t *testing.T, actualDisplayData, expectedDisplayData *openid4ci.DisplayData) {
	t.Helper()

	checkIssuerDisplay(t, actualDisplayData.IssuerDisplay(), expectedDisplayData.IssuerDisplay())

	require.Equal(t, 1, actualDisplayData.CredentialDisplaysLength())

	actualCredentialDisplay := actualDisplayData.CredentialDisplayAtIndex(0)
	expectedCredentialDisplay := expectedDisplayData.CredentialDisplayAtIndex(0)

	checkCredentialDisplay(t, actualCredentialDisplay, expectedCredentialDisplay)
}

func checkIssuerDisplay(t *testing.T, actualIssuerDisplay, expectedIssuerDisplay *openid4ci.IssuerDisplay) {
	t.Helper()

	require.Equal(t, expectedIssuerDisplay.Name(), actualIssuerDisplay.Name())
	require.Equal(t, expectedIssuerDisplay.Locale(), actualIssuerDisplay.Locale())
}

func checkCredentialDisplay(t *testing.T, actualCredentialDisplay, expectedCredentialDisplay *openid4ci.CredentialDisplay) {
	t.Helper()

	actualCredentialOverview := actualCredentialDisplay.Overview()
	expectedCredentialOverview := expectedCredentialDisplay.Overview()

	require.Equal(t, expectedCredentialOverview.Name(), actualCredentialOverview.Name())
	require.Equal(t, expectedCredentialOverview.Locale(), actualCredentialOverview.Locale())
	require.Equal(t, expectedCredentialOverview.Logo().URL(), actualCredentialOverview.Logo().URL())
	require.Equal(t, expectedCredentialOverview.Logo().AltText(), actualCredentialOverview.Logo().AltText())
	require.Equal(t, expectedCredentialOverview.BackgroundColor(), actualCredentialOverview.BackgroundColor())
	require.Equal(t, expectedCredentialOverview.TextColor(), actualCredentialOverview.TextColor())

	require.Equal(t, expectedCredentialDisplay.ClaimsLength(), actualCredentialDisplay.ClaimsLength())

	// Since the claims object in the supported_credentials object from the issuer is a map which effectively gets
	// converted to an array of resolved claims, the order of resolved claims can differ from run-to-run. The code
	// below checks to ensure we have the expected claims in any order.

	expectedClaims := make([]*openid4ci.Claim, expectedCredentialDisplay.ClaimsLength())

	for i := 0; i < len(expectedClaims); i++ {
		expectedClaims[i] = expectedCredentialDisplay.ClaimAtIndex(i)
	}

	expectedClaimsChecklist := struct {
		Claims []*openid4ci.Claim
		Found  []bool
	}{
		Claims: expectedClaims,
		Found:  make([]bool, len(expectedClaims)),
	}

	for i := 0; i < actualCredentialDisplay.ClaimsLength(); i++ {
		claim := actualCredentialDisplay.ClaimAtIndex(i)

		for j := 0; j < len(expectedClaimsChecklist.Claims); j++ {
			expectedClaim := expectedClaimsChecklist.Claims[j]
			if claim.Label() == expectedClaim.Label() &&
				claim.Value() == expectedClaim.Value() &&
				claim.Locale() == expectedClaim.Locale() {
				if expectedClaimsChecklist.Found[j] {
					require.FailNow(t, "duplicate claim found: ",
						"[Label: %s] [Value: %s] [Locale: %s]",
						claim.Label(), claim.Value(), claim.Locale())
				}

				expectedClaimsChecklist.Found[j] = true

				break
			}

			if j == len(expectedClaimsChecklist.Claims)-1 {
				require.FailNow(t, "received unexpected claim: ",
					"[Label: %s] [Value: %s] [Locale: %s]",
					claim.Label(), claim.Value(), claim.Locale())
			}
		}
	}

	for i := 0; i < len(expectedClaimsChecklist.Claims); i++ {
		if !expectedClaimsChecklist.Found[i] {
			expectedClaim := expectedClaimsChecklist.Claims[i]
			require.FailNow(t, "the following claim was expected but wasn't received: ",
				"[Label: %s] [Value: %s] [Locale: %s]",
				expectedClaim.Label(), expectedClaim.Value(), expectedClaim.Locale())
		}
	}
}

func checkActivityLogAfterOpenID4CIFlow(t *testing.T, activityLogger *mem.ActivityLogger,
	issuerProfileID, expectedSubjectID string,
) {
	numberOfActivitiesLogged := activityLogger.Length()
	require.Equal(t, 1, numberOfActivitiesLogged)

	activity := activityLogger.AtIndex(0)

	require.NotEmpty(t, activity.ID())
	require.Equal(t, goapi.LogTypeCredentialActivity, activity.Type())
	require.NotEmpty(t, activity.UnixTimestamp())
	require.Equal(t, oidc4ci.VCSAPIDirect+"/"+issuerProfileID, activity.Client())
	require.Equal(t, "oidc-issuance", activity.Operation())
	require.Equal(t, goapi.ActivityLogStatusSuccess, activity.Status())

	params := activity.Params()
	require.NotNil(t, params)

	keyValuePairs := params.AllKeyValuePairs()

	numberOfKeyValuePairs := keyValuePairs.Length()

	require.Equal(t, 1, numberOfKeyValuePairs)

	keyValuePair := keyValuePairs.AtIndex(0)

	key := keyValuePair.Key()
	require.Equal(t, "subjectIDs", key)

	subjectIDs, err := keyValuePair.ValueStringArray()
	require.NoError(t, err)

	numberOfSubjectIDs := subjectIDs.Length()
	require.Equal(t, 1, numberOfSubjectIDs)

	subjectID := subjectIDs.AtIndex(0)
	require.Equal(t, expectedSubjectID, subjectID)
}
