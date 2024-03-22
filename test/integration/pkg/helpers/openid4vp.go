/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package helpers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didion"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didjwk"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didkey"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/metricslogger/stderr"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/metricslogger"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4ci"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/testenv"
)

type VPTestHelper struct {
	KMS    *localkms.KMS
	DIDDoc *api.DIDDocResolution
}

type IssuerInfo struct {
	ProfileID string
	IssuerURI string
}

func NewVPTestHelper(t *testing.T, didMethod string, keyType string) *VPTestHelper {
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

	return &VPTestHelper{
		KMS:    kms,
		DIDDoc: didDoc,
	}
}

func (h *VPTestHelper) IssueCredentials(t *testing.T, vcsAPIDirectURL string, issuerProfileIDs []string,
	claimData []map[string]interface{}, documentLoader api.LDDocumentLoader,
) (*verifiable.CredentialsArray, map[string]IssuerInfo) {
	oidc4ciSetup, err := oidc4ci.NewSetup(testenv.NewHttpRequest())
	require.NoError(t, err)

	err = oidc4ciSetup.AuthorizeIssuerBypassAuth("f13d1va9lp403pb9lyj89vk55", vcsAPIDirectURL)
	require.NoError(t, err)

	credentials := verifiable.NewCredentialsArray()

	issuerInfo := map[string]IssuerInfo{}

	for i := 0; i < len(issuerProfileIDs); i++ {
		offerCredentialURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance(issuerProfileIDs[i],
			[]oidc4ci.CredentialConfiguration{
				{
					ClaimData: claimData[i],
				},
			},
		)
		require.NoError(t, err)

		opts := did.NewResolverOpts()
		opts.SetResolverServerURI("http://did-resolver.trustbloc.local:8072/1.0/identifiers")

		didResolver, err := did.NewResolver(opts)
		require.NoError(t, err)

		requiredArgs := openid4ci.NewIssuerInitiatedInteractionArgs(offerCredentialURL, h.KMS.GetCrypto(), didResolver)

		interactionOptionalArgs := openid4ci.NewInteractionOpts()
		interactionOptionalArgs.SetMetricsLogger(stderr.NewMetricsLogger())
		interactionOptionalArgs.SetDocumentLoader(documentLoader)
		interactionOptionalArgs.DisableHTTPClientTLSVerify()

		interaction, err := openid4ci.NewIssuerInitiatedInteraction(requiredArgs, interactionOptionalArgs)
		require.NoError(t, err)

		require.True(t, interaction.PreAuthorizedCodeGrantTypeSupported())

		preAuthorizedCodeGrantParams, err := interaction.PreAuthorizedCodeGrantParams()
		require.NoError(t, err)

		require.False(t, preAuthorizedCodeGrantParams.PINRequired())

		vm, err := h.DIDDoc.AssertionMethod()
		require.NoError(t, err)

		result, err := interaction.RequestCredentialWithPreAuth(vm, nil)
		require.NoError(t, err)
		require.NotEmpty(t, result)

		for i := 0; i < result.Length(); i++ {
			vc := result.AtIndex(i)

			serializedVC, err := vc.Serialize()
			require.NoError(t, err)

			println(serializedVC)
			credentials.Add(vc)

			issuerInfo[vc.ID()] = IssuerInfo{
				ProfileID: issuerProfileIDs[i],
				IssuerURI: interaction.IssuerURI(),
			}
		}

	}

	return credentials, issuerInfo
}

func (h *VPTestHelper) CheckActivityLogAfterOpenID4VPFlow(t *testing.T, activityLogger *mem.ActivityLogger, verifierProfileID string) {
	numberOfActivitiesLogged := activityLogger.Length()
	require.Equal(t, 1, numberOfActivitiesLogged)

	activity := activityLogger.AtIndex(0)

	require.NotEmpty(t, activity.ID())
	require.Equal(t, goapi.LogTypeCredentialActivity, activity.Type())
	require.NotEmpty(t, activity.UnixTimestamp())
	require.Equal(t, verifierProfileID, activity.Client())
	require.Equal(t, "oidc-presentation", activity.Operation())
	require.Equal(t, goapi.ActivityLogStatusSuccess, activity.Status())
	require.Equal(t, 0, activity.Params().AllKeyValuePairs().Length())
}

func (h *VPTestHelper) CheckMetricsLoggerAfterOpenID4VPFlow(t *testing.T, metricsLogger *metricslogger.MetricsLogger) {
	require.Len(t, metricsLogger.Events, 3)

	h.checkFetchRequestObjectMetricsEvent(t, metricsLogger.Events[0])
	h.checkSendAuthorizedResponseMetricsEvent(t, metricsLogger.Events[1])
	h.checkPresentCredentialMetricsEvent(t, metricsLogger.Events[2])
}

func (h *VPTestHelper) checkFetchRequestObjectMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent) {
	require.Contains(t, metricsEvent.Event(), "Fetch request object via an HTTP GET request to "+
		"http://localhost:8075/request-object/")
	require.Equal(t, "Instantiating OpenID4VP interaction object", metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}

func (h *VPTestHelper) checkGetQueryMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent) {
	require.Equal(t, "Get query", metricsEvent.Event())
	require.Empty(t, metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}

func (h *VPTestHelper) checkSendAuthorizedResponseMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent) {
	require.Equal(t, "Send authorized response via an HTTP POST request to "+
		"http://localhost:8075/oidc/present",
		metricsEvent.Event())
	require.Equal(t, "Present credential", metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}

func (h *VPTestHelper) checkPresentCredentialMetricsEvent(t *testing.T, metricsEvent *api.MetricsEvent) {
	require.Equal(t, "Present credential", metricsEvent.Event())
	require.Empty(t, metricsEvent.ParentEvent())
	require.Positive(t, metricsEvent.DurationNanoseconds())
}
