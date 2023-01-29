/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"testing"

	goapi "github.com/trustbloc/wallet-sdk/pkg/api"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4ci"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/testenv"
)

// Run this lines to make test work locally
// echo '127.0.0.1 testnet.orb.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 file-server.trustbloc.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 did-resolver.trustbloc.local' | sudo tee -a /etc/hosts
// echo '127.0.0.1 oidc-provider.example.com' | sudo tee -a /etc/hosts
// echo '127.0.0.1 vc-rest-echo.trustbloc.local' | sudo tee -a /etc/hosts

func TestOpenID4CIFullFlow(t *testing.T) {
	type test struct {
		issuerProfileID string
		issuerDIDMethod string
		walletDIDMethod string
	}

	tests := []test{
		{issuerProfileID: "bank_issuer", issuerDIDMethod: "orb", walletDIDMethod: "ion"},
		{issuerProfileID: "did_ion_issuer", issuerDIDMethod: "ion", walletDIDMethod: "key"},
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

		signerCreator, err := localkms.CreateSignerCreator(kms)
		require.NoError(t, err)

		didResolver, err := did.NewResolver("")
		require.NoError(t, err)

		didID, err := didDoc.ID()
		require.NoError(t, err)

		activityLogger := mem.NewActivityLogger()

		clientConfig := openid4ci.ClientConfig{
			UserDID:        didID,
			ClientID:       "ClientID",
			SignerCreator:  signerCreator,
			DIDResolver:    didResolver,
			ActivityLogger: activityLogger,
		}

		interaction, err := openid4ci.NewInteraction(initiateIssuanceURL, &clientConfig)
		require.NoError(t, err)

		authorizeResult, err := interaction.Authorize()
		require.NoError(t, err)
		require.False(t, authorizeResult.UserPINRequired)

		credential, err := interaction.RequestCredential(openid4ci.NewCredentialRequestOpts(""))

		require.NoError(t, err)
		require.NotNil(t, credential)

		vc := credential.AtIndex(0)

		serializedVC, err := vc.Serialize()
		require.NoError(t, err)

		println("credential:", serializedVC)
		require.NoError(t, err)
		require.Contains(t, vc.VC.Issuer.ID, tc.issuerDIDMethod)

		subID, err := verifiable.SubjectID(vc.VC.Subject)
		require.Contains(t, subID, didID)

		checkActivityLogAfterOpenID4CIFlow(t, activityLogger, tc.issuerProfileID, subID)
	}
}

func checkActivityLogAfterOpenID4CIFlow(t *testing.T, activityLogger *mem.ActivityLogger,
	issuerProfileID, expectedSubjectID string,
) {
	numberOfActivitiesLogged := activityLogger.Length()
	require.Equal(t, 1, numberOfActivitiesLogged)

	activity := activityLogger.AtIndex(0)

	require.NotEmpty(t, activity.ID)
	require.Equal(t, goapi.LogTypeCredentialActivity, activity.Type)
	require.NotEmpty(t, activity.Time)
	require.NotNil(t, activity.Data)
	require.Equal(t, oidc4ci.VCSAPIDirect+"/"+issuerProfileID, activity.Data.Client)
	require.Equal(t, "oidc-issuance", activity.Data.Operation)
	require.Equal(t, goapi.ActivityLogStatusSuccess, activity.Data.Status)
	require.NotNil(t, activity.Data.Params)
	require.Equal(t, fmt.Sprintf(`{"subjectIDs":["%s"]}`, expectedSubjectID),
		string(activity.Data.Params.Data))
}
