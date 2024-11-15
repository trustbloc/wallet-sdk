/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/display"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/metricslogger"
)

func ParseDisplayData(t *testing.T, displayData string) *display.Data {
	parsedDisplayData, err := display.ParseData(displayData)
	require.NoError(t, err)

	return parsedDisplayData
}

func ResolveDisplayData(t *testing.T, credentials *verifiable.CredentialsArray, expectedDisplayData *display.Data,
	issuerURI, issuerProfileID string, didResolver *did.Resolver,
) {
	metricsLogger := metricslogger.NewMetricsLogger()

	opts := display.NewOpts()
	opts.SetMetricsLogger(metricsLogger)
	opts.DisableHTTPClientTLSVerify()
	opts.SetDIDResolver(didResolver)

	displayData, err := display.Resolve(credentials, issuerURI, opts)
	require.NoError(t, err)
	CheckResolvedDisplayData(t, displayData, expectedDisplayData, true)

	checkResolveMetricsEvent(t, metricsLogger, issuerProfileID)
}

// CheckResolvedDisplayData function assumes that the display data object has only a single credential display.
// In the event we add a test case where there are multiple credential displays, then this function will need to be
// updated accordingly.
func CheckResolvedDisplayData(t *testing.T, actualDisplayData, expectedDisplayData *display.Data, checkClaims bool) {
	t.Helper()

	checkIssuerDisplay(t, actualDisplayData.IssuerDisplay(), expectedDisplayData.IssuerDisplay())

	require.Equal(t, expectedDisplayData.CredentialDisplaysLength(), actualDisplayData.CredentialDisplaysLength())

	actualCredentialDisplay := actualDisplayData.CredentialDisplayAtIndex(0)
	expectedCredentialDisplay := expectedDisplayData.CredentialDisplayAtIndex(0)

	if checkClaims && expectedDisplayData.CredentialDisplaysLength() > 1 {
		expectedClaims := claimsMap(expectedCredentialDisplay)

		for j := 0; j < actualDisplayData.CredentialDisplaysLength(); j++ {
			actualCredentialDisplay = actualDisplayData.CredentialDisplayAtIndex(j)
			actualClaims := claimsMap(actualCredentialDisplay)

			for k, v := range expectedClaims {
				if actualClaims[k] != v {
					break
				}
			}
		}
	}

	checkCredentialDisplay(t, actualCredentialDisplay, expectedCredentialDisplay, checkClaims)
}

func ResolveDisplayDataV2(t *testing.T, credentials *verifiable.CredentialsArrayV2, expectedDisplayData *display.Data,
	issuerURI, issuerProfileID string, didResolver *did.Resolver,
) {
	metricsLogger := metricslogger.NewMetricsLogger()

	opts := display.NewOpts()
	opts.SetMetricsLogger(metricsLogger)
	opts.DisableHTTPClientTLSVerify()
	opts.SetDIDResolver(didResolver)

	resolvedDisplayData, err := display.ResolveCredentialV2(credentials, issuerURI, opts)
	require.NoError(t, err)
	require.NotNil(t, resolvedDisplayData)

	CheckResolvedDisplayDataV2(t, resolvedDisplayData, expectedDisplayData)

	checkResolveMetricsEvent(t, metricsLogger, issuerProfileID)
}

func CheckResolvedDisplayDataV2(t *testing.T, resolvedDisplayData *display.Resolved, expectedDisplayData *display.Data) {
	t.Helper()

	require.Equal(t, 1, resolvedDisplayData.LocalizedIssuersLength())

	resolvedIssuerData := resolvedDisplayData.LocalizedIssuerAtIndex(0)
	expectedIssuerData := expectedDisplayData.IssuerDisplay()

	require.Equal(t, expectedIssuerData.Name(), resolvedIssuerData.Name())
	require.Equal(t, expectedIssuerData.Locale(), resolvedIssuerData.Locale())
	require.Equal(t, expectedIssuerData.BackgroundColor(), resolvedIssuerData.BackgroundColor())
	require.Equal(t, expectedIssuerData.TextColor(), resolvedIssuerData.TextColor())

	require.Equal(t, expectedDisplayData.CredentialDisplaysLength(), resolvedDisplayData.CredentialsLength())

	for i := 0; i < expectedDisplayData.CredentialDisplaysLength(); i++ {
		expectedCredentialDisplay := expectedDisplayData.CredentialDisplayAtIndex(i)
		resolvedCredentialDisplay := resolvedDisplayData.CredentialAtIndex(i)

		require.Equal(t, resolvedCredentialDisplay.LocalizedOverviewsLength(), 1)

		expectedOverview := expectedCredentialDisplay.Overview()
		resolvedOverview := resolvedCredentialDisplay.LocalizedOverviewAtIndex(0)

		require.Equal(t, expectedOverview.Name(), resolvedOverview.Name())
		require.Equal(t, expectedOverview.Locale(), resolvedOverview.Locale())
		require.Equal(t, expectedOverview.BackgroundColor(), resolvedOverview.BackgroundColor())
		require.Equal(t, expectedOverview.TextColor(), resolvedOverview.TextColor())
		require.Equal(t, expectedOverview.Logo(), resolvedOverview.Logo())

		require.Equal(t, expectedCredentialDisplay.ClaimsLength(), resolvedCredentialDisplay.SubjectsLength())
	}
}

func claimsMap(credentialDisplay *display.CredentialDisplay) map[string]string {
	m := make(map[string]string)

	for i := 0; i < credentialDisplay.ClaimsLength(); i++ {
		claim := credentialDisplay.ClaimAtIndex(i)
		m[claim.Label()] = claim.Value()
	}

	return m
}

func checkResolveMetricsEvent(t *testing.T, metricsLogger *metricslogger.MetricsLogger, issuerProfileID string) {
	require.Len(t, metricsLogger.Events, 1)

	checkFetchIssuerMetadataMetricsEvent(t, metricsLogger.Events[0], "Resolve display", issuerProfileID)
}
