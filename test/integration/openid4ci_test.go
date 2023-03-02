/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"

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
		walletKeyType       string
		expectedIssuerURI   string
		expectedDisplayData *display.Data
	}

	tests := []test{
		{
			issuerProfileID:     "bank_issuer_jwtsd",
			issuerDIDMethod:     "orb",
			walletDIDMethod:     "ion",
			expectedIssuerURI:   "http://localhost:8075/issuer/bank_issuer_jwtsd",
			expectedDisplayData: parseDisplayData(t, expectedDisplayDataBankIssuer),
			walletKeyType:       localkms.KeyTypeP384,
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

		offerCredentialURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance(tc.issuerProfileID)
		require.NoError(t, err)

		println(offerCredentialURL)

		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		didDoc := createDID(t, kms, tc.walletKeyType, tc.walletDIDMethod)

		didResolver, err := did.NewResolver("http://did-resolver.trustbloc.local:8072/1.0/identifiers")
		require.NoError(t, err)

		didID, err := didDoc.ID()
		require.NoError(t, err)

		activityLogger := mem.NewActivityLogger()

		metricsLogger := NewMetricsLogger()

		clientConfig := openid4ci.ClientConfig{
			ClientID:       "ClientID",
			DIDResolver:    didResolver,
			ActivityLogger: activityLogger,
			Crypto:         kms.GetCrypto(),
			MetricsLogger:  metricsLogger,
		}

		interaction, err := openid4ci.NewInteraction(offerCredentialURL, &clientConfig)
		require.NoError(t, err)

		authorizeResult, err := interaction.Authorize()
		require.NoError(t, err)
		require.False(t, authorizeResult.UserPINRequired)

		vm, err := didDoc.AssertionMethod()
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

		resolveDisplayData(t, credentials, tc.expectedDisplayData, interaction.IssuerURI(), tc.issuerProfileID)

		issuerURI := interaction.IssuerURI()
		require.Equal(t, tc.expectedIssuerURI, issuerURI)

		subID, err := verifiable.SubjectID(vc.VC.Subject)
		require.NoError(t, err)
		require.Contains(t, subID, didID)
		checkActivityLogAfterOpenID4CIFlow(t, activityLogger, tc.issuerProfileID, subID)
		checkMetricsLoggerAfterOpenID4CIFlow(t, metricsLogger, tc.issuerProfileID)
	}
}

func createDID(t *testing.T, kms *localkms.KMS, keyType, didMethod string) *api.DIDDocResolution {
	didCreator, err := did.NewCreatorWithKeyWriter(kms)
	require.NoError(t, err)

	metricsLogger := NewMetricsLogger()

	createDIDOpts := &api.CreateDIDOpts{
		KeyType:       keyType,
		MetricsLogger: metricsLogger,
	}

	didDoc, err := didCreator.Create(didMethod, createDIDOpts)
	require.NoError(t, err)

	checkDIDCreationMetricsLoggerEvents(t, metricsLogger)

	return didDoc
}

func checkDIDCreationMetricsLoggerEvents(t *testing.T, metricsLogger *MetricsLogger) {
	require.Len(t, metricsLogger.events, 1)

	require.Equal(t, "Creating DID", metricsLogger.events[0].Event())
	require.Empty(t, metricsLogger.events[0].ParentEvent())
	require.Positive(t, metricsLogger.events[0].DurationNanoseconds())
}

func parseDisplayData(t *testing.T, displayData string) *display.Data {
	parsedDisplayData, err := display.ParseData(displayData)
	require.NoError(t, err)

	return parsedDisplayData
}

func resolveDisplayData(t *testing.T, credentials *api.VerifiableCredentialsArray, expectedDisplayData *display.Data,
	issuerURI, issuerProfileID string,
) {
	metricsLogger := NewMetricsLogger()

	resolveDisplayOpts := &display.ResolveOpts{
		VCs:           credentials,
		IssuerURI:     issuerURI,
		MetricsLogger: metricsLogger,
	}

	displayData, err := display.Resolve(resolveDisplayOpts)
	require.NoError(t, err)
	checkResolvedDisplayData(t, displayData, expectedDisplayData)

	checkResolveMetricsEvent(t, metricsLogger, issuerProfileID)
}

func checkResolveMetricsEvent(t *testing.T, metricsLogger *MetricsLogger, issuerProfileID string) {
	require.Len(t, metricsLogger.events, 1)

	checkFetchIssuerMetadataMetricsEvent(t, metricsLogger.events[0], "Resolve display", issuerProfileID)
}

// For now, this function assumes that the display data object has only a single credential display.
// In the event we add a test case where there are multiple credential displays, then this function will need to be
// updated accordingly.
func checkResolvedDisplayData(t *testing.T, actualDisplayData, expectedDisplayData *display.Data) {
	t.Helper()

	checkIssuerDisplay(t, actualDisplayData.IssuerDisplay(), expectedDisplayData.IssuerDisplay())

	require.Equal(t, 1, actualDisplayData.CredentialDisplaysLength())

	actualCredentialDisplay := actualDisplayData.CredentialDisplayAtIndex(0)
	expectedCredentialDisplay := expectedDisplayData.CredentialDisplayAtIndex(0)

	checkCredentialDisplay(t, actualCredentialDisplay, expectedCredentialDisplay)
}

func checkIssuerDisplay(t *testing.T, actualIssuerDisplay, expectedIssuerDisplay *display.IssuerDisplay) {
	t.Helper()

	require.Equal(t, expectedIssuerDisplay.Name(), actualIssuerDisplay.Name())
	require.Equal(t, expectedIssuerDisplay.Locale(), actualIssuerDisplay.Locale())
}

func checkCredentialDisplay(t *testing.T, actualCredentialDisplay, expectedCredentialDisplay *display.CredentialDisplay) {
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

	expectedClaims := make([]*display.Claim, expectedCredentialDisplay.ClaimsLength())

	for i := 0; i < len(expectedClaims); i++ {
		expectedClaims[i] = expectedCredentialDisplay.ClaimAtIndex(i)
	}

	expectedClaimsChecklist := struct {
		Claims []*display.Claim
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
						"[Label: %s] [Value Type: %s] [Value: %s] [Locale: %s]",
						claim.Label(), claim.ValueType(), claim.Value(), claim.Locale())
				}

				expectedClaimsChecklist.Found[j] = true

				break
			}

			if j == len(expectedClaimsChecklist.Claims)-1 {
				require.FailNow(t, "received unexpected claim: ",
					"[Label: %s] [Value Type: %s] [Value: %s] [Locale: %s]",
					claim.Label(), claim.ValueType(), claim.Value(), claim.Locale())
			}
		}
	}

	for i := 0; i < len(expectedClaimsChecklist.Claims); i++ {
		if !expectedClaimsChecklist.Found[i] {
			expectedClaim := expectedClaimsChecklist.Claims[i]
			require.FailNow(t, "the following claim was expected but wasn't received: ",
				"[Label: %s] [Value Type: %s] [Value: %s] [Locale: %s]",
				expectedClaim.Label, expectedClaim.ValueType, expectedClaim.Value, expectedClaim.Locale)
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
	require.Equal(t, oidc4ci.VCSAPIDirect+"/issuer/"+issuerProfileID, activity.Client())
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

func checkMetricsLoggerAfterOpenID4CIFlow(t *testing.T, metricsLogger *MetricsLogger, issuerProfileID string) {
	require.Len(t, metricsLogger.events, 7)

	checkInteractionInstantiationMetricsEvent(t, metricsLogger.events[0])

	checkFetchOpenIDConfigMetricsEvent(t, metricsLogger.events[1], issuerProfileID)

	checkFetchTokenMetricsEvent(t, metricsLogger.events[2])

	checkFetchIssuerMetadataMetricsEvent(t, metricsLogger.events[3],
		"Request credential(s) from issuer", issuerProfileID)

	checkFetchCredentialHTTPRequestMetricsEvent(t, metricsLogger.events[4])

	checkParseCredentialMetricsEvent(t, metricsLogger.events[5])

	checkRequestCredentialsMetricsEvent(t, metricsLogger.events[6])
}

func checkInteractionInstantiationMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent) {
	require.Equal(t, "Instantiating OpenID4CI interaction object", metricsEvent.Event())
	require.Empty(t, metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}

func checkFetchOpenIDConfigMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent, issuerProfileID string) {
	expectedEndpoint := fmt.Sprintf("http://localhost:8075/issuer/%s/.well-known/openid-configuration",
		issuerProfileID)

	require.Equal(t,
		fmt.Sprintf("Fetch issuer's OpenID configuration via an HTTP GET request to %s", expectedEndpoint),
		metricsEvent.Event())
	require.Equal(t, "Request credential(s) from issuer", metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}

func checkFetchTokenMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent) {
	expectedEndpoint := "http://localhost:8075/oidc/token"

	require.Equal(t, fmt.Sprintf("Fetch token via an HTTP POST request to %s", expectedEndpoint),
		metricsEvent.Event())
	require.Equal(t, "Request credential(s) from issuer", metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}

func checkFetchIssuerMetadataMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent, expectedParentEvent,
	issuerProfileID string,
) {
	expectedEndpoint := fmt.Sprintf("http://localhost:8075/issuer/%s/.well-known/openid-credential-issuer",
		issuerProfileID)

	require.Equal(t, fmt.Sprintf("Fetch issuer metadata via an HTTP GET request to %s", expectedEndpoint),
		metricsEvent.Event())
	require.Equal(t, expectedParentEvent, metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}

func checkFetchCredentialHTTPRequestMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent) {
	expectedEndpoint := "http://localhost:8075/oidc/credential"

	require.Equal(t, fmt.Sprintf("Fetch credential 1 of 1 via an HTTP POST request to %s",
		expectedEndpoint), metricsEvent.Event())
	require.Equal(t, "Request credential(s) from issuer", metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}

func checkParseCredentialMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent) {
	require.Equal(t, "Parsing and checking proof for received credential 1 of 1", metricsEvent.Event())
	require.Equal(t, "Request credential(s) from issuer", metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}

func checkRequestCredentialsMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent) {
	require.Equal(t, "Request credential(s) from issuer", metricsEvent.Event())
	require.Empty(t, metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}
