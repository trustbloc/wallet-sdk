/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package helpers

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didion"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didjwk"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didkey"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/metricslogger"
)

type CITestHelper struct {
	KMS            *localkms.KMS
	DIDDoc         *api.DIDDocResolution
	ActivityLogger *mem.ActivityLogger
	MetricsLogger  *metricslogger.MetricsLogger
}

func NewCITestHelper(t *testing.T, didMethod string, keyType string) *CITestHelper {
	t.Helper()

	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	if keyType == "" {
		keyType = localkms.KeyTypeED25519
	}

	jwk, err := kms.Create(keyType)
	require.NoError(t, err)

	var didDoc *api.DIDDocResolution

	switch didMethod {
	case "key":
		didDoc, err = didkey.Create(jwk)
		require.NoError(t, err)
	case "jwk":
		didDoc, err = didjwk.Create(jwk)
		require.NoError(t, err)
	case "ion":
		didDoc, err = didion.CreateLongForm(jwk)
		require.NoError(t, err)
	default:
		require.Fail(t, fmt.Sprintf("%s is not a supported DID method", didMethod))
	}

	return &CITestHelper{
		KMS:            kms,
		DIDDoc:         didDoc,
		ActivityLogger: mem.NewActivityLogger(),
		MetricsLogger:  metricslogger.NewMetricsLogger(),
	}
}

func (h *CITestHelper) CheckActivityLogAfterOpenID4CIFlow(t *testing.T, vcsAPIDirectURL,
	issuerProfileID, expectedSubjectID string,
) {
	numberOfActivitiesLogged := h.ActivityLogger.Length()
	require.Equal(t, 1, numberOfActivitiesLogged)

	activity := h.ActivityLogger.AtIndex(0)

	require.NotEmpty(t, activity.ID())
	require.Equal(t, goapi.LogTypeCredentialActivity, activity.Type())
	require.NotEmpty(t, activity.UnixTimestamp())
	require.Equal(t, vcsAPIDirectURL+"/issuer/"+issuerProfileID+"/v1.0", activity.Client())
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
	require.True(t, numberOfSubjectIDs > 0)

	subjectID := subjectIDs.AtIndex(0)
	require.Equal(t, expectedSubjectID, subjectID)
}

func (h *CITestHelper) CheckMetricsLoggerAfterOpenID4CIFlow(t *testing.T, issuerProfileID string) {
	require.Len(t, h.MetricsLogger.Events, 7)

	checkInteractionInstantiationMetricsEvent(t, h.MetricsLogger.Events[0])

	checkFetchIssuerMetadataMetricsEvent(t, h.MetricsLogger.Events[1],
		"Get issuer metadata", issuerProfileID)

	checkFetchOpenIDConfigMetricsEvent(t, h.MetricsLogger.Events[2], issuerProfileID)

	checkFetchTokenMetricsEvent(t, h.MetricsLogger.Events[3])

	checkFetchCredentialHTTPRequestMetricsEvent(t, h.MetricsLogger.Events[4])

	checkParseCredentialMetricsEvent(t, h.MetricsLogger.Events[5])

	checkRequestCredentialsMetricsEvent(t, h.MetricsLogger.Events[6])
}

func checkIssuerDisplay(t *testing.T, actualIssuerDisplay, expectedIssuerDisplay *display.IssuerDisplay) {
	t.Helper()

	require.Equal(t, expectedIssuerDisplay.Name(), actualIssuerDisplay.Name())
	require.Equal(t, expectedIssuerDisplay.Locale(), actualIssuerDisplay.Locale())
}

func checkCredentialDisplay(t *testing.T, actualCredentialDisplay,
	expectedCredentialDisplay *display.CredentialDisplay, checkClaims bool) {
	t.Helper()

	actualCredentialOverview := actualCredentialDisplay.Overview()
	expectedCredentialOverview := expectedCredentialDisplay.Overview()

	require.Equal(t, expectedCredentialOverview.Name(), actualCredentialOverview.Name())
	require.Equal(t, expectedCredentialOverview.Locale(), actualCredentialOverview.Locale())
	require.Equal(t, expectedCredentialOverview.Logo().URL(), actualCredentialOverview.Logo().URL())
	require.Equal(t, expectedCredentialOverview.Logo().AltText(), actualCredentialOverview.Logo().AltText())
	require.Equal(t, expectedCredentialOverview.BackgroundColor(), actualCredentialOverview.BackgroundColor())
	require.Equal(t, expectedCredentialOverview.TextColor(), actualCredentialOverview.TextColor())

	if !checkClaims {
		return
	}

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
				claim.RawID() == expectedClaim.RawID() &&
				claim.RawValue() == expectedClaim.RawValue() &&
				claim.Value() == expectedClaim.Value() &&
				claim.Locale() == expectedClaim.Locale() &&
				claim.IsMasked() == expectedClaim.IsMasked() &&
				claimOrderMatches(t, claim, expectedClaim) {
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

	for i := 0; i < len(expectedClaimsChecklist.Claims); i++ {
		if !expectedClaimsChecklist.Found[i] {
			expectedClaim := expectedClaimsChecklist.Claims[i]
			require.FailNow(t, "claim was expected but wasn't received",
				"[Claim ID: %s] [Pattern: %s] [Raw value: %s] [Label: %s] "+
					"[Value Type: %s] [Value: %s] [Order: %s] [Locale: %s]",
				expectedClaim.RawID, expectedClaim.Pattern, expectedClaim.RawValue, expectedClaim.Label,
				expectedClaim.ValueType, expectedClaim.Value, getOrderAsString(t, expectedClaim),
				expectedClaim.Locale)
		}
	}
}

func claimOrderMatches(t *testing.T, actualClaim, expectedClaim *display.Claim) bool {
	t.Helper()

	if actualClaim.HasOrder() {
		if !expectedClaim.HasOrder() {
			return false
		}

		actualClaimOrder, err := actualClaim.Order()
		require.NoError(t, err)

		expectedClaimOrder, err := expectedClaim.Order()
		require.NoError(t, err)

		if actualClaimOrder != expectedClaimOrder {
			return false
		}
	} else if expectedClaim.HasOrder() {
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

func checkInteractionInstantiationMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent) {
	require.Equal(t, "Instantiating OpenID4CI interaction object", metricsEvent.Event())
	require.Empty(t, metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}

func checkFetchOpenIDConfigMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent, issuerProfileID string) {
	expectedEndpoint := fmt.Sprintf("http://localhost:8075/oidc/idp/%s/v1.0/.well-known"+
		"/openid-configuration",
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
	expectedEndpoint := fmt.Sprintf("http://localhost:8075/oidc/idp/%s/v1.0/.well-known"+
		"/openid-credential-issuer",
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
