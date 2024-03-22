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

	checkCredentialDisplay(t, actualCredentialDisplay, expectedCredentialDisplay, checkClaims)
}

func checkResolveMetricsEvent(t *testing.T, metricsLogger *metricslogger.MetricsLogger, issuerProfileID string) {
	require.Len(t, metricsLogger.Events, 1)

	checkFetchIssuerMetadataMetricsEvent(t, metricsLogger.Events[0], "Resolve display", issuerProfileID)
}
