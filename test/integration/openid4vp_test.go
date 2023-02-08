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

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/ld"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4vp"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4ci"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4vp"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/testenv"
)

func TestOpenID4VPFullFlow(t *testing.T) {
	type test struct {
		issuerProfileIDs  []string
		walletDIDMethod   string
		verifierProfileID string
	}

	tests := []test{
		{
			issuerProfileIDs:  []string{"bank_issuer"},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_myprofile_jwt_verified_employee",
		},
		{
			issuerProfileIDs:  []string{"bank_issuer"},
			walletDIDMethod:   "key",
			verifierProfileID: "v_myprofile_jwt_verified_employee",
		},
		{
			issuerProfileIDs:  []string{"bank_issuer"},
			walletDIDMethod:   "jwk",
			verifierProfileID: "v_myprofile_jwt_verified_employee",
		},
		{
			issuerProfileIDs:  []string{"bank_issuer_jwtsd"},
			walletDIDMethod:   "jwk",
			verifierProfileID: "v_myprofile_sdjwt",
		},
		{
			issuerProfileIDs:  []string{"drivers_license_issuer"},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_myprofile_jwt_drivers_license",
		},
		{
			issuerProfileIDs:  []string{"bank_issuer", "drivers_license_issuer"},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_myprofile_jwt_verified_employee",
		},
	}

	for i, tc := range tests {
		fmt.Printf("running test %d: issuerProfileIDs=%s verifierProfileID=%s "+
			"walletDIDMethod=%s\n", i,
			tc.issuerProfileIDs, tc.verifierProfileID, tc.walletDIDMethod)

		testHelper := newVPTestHelper(t, tc.walletDIDMethod)

		issuedCredentials := testHelper.issueCredentials(t, tc.issuerProfileIDs)
		println("Issued", issuedCredentials.Length(), "credentials")
		for k := 0; k < issuedCredentials.Length(); k++ {
			cred, _ := issuedCredentials.AtIndex(k).Serialize()
			println("Issued VC[", k, "]: ", cred)
		}

		setup := oidc4vp.NewSetup(testenv.NewHttpRequest())

		err := setup.AuthorizeVerifierBypassAuth("test_org")
		require.NoError(t, err)

		initiateURL, err := setup.InitiateInteraction(tc.verifierProfileID)
		require.NoError(t, err)

		didResolver, err := did.NewResolver("")
		require.NoError(t, err)

		activityLogger := mem.NewActivityLogger()

		cfg := openid4vp.NewClientConfig(
			testHelper.KMS, testHelper.KMS.GetCrypto(), didResolver, ld.NewDocLoader(), activityLogger)

		interaction := openid4vp.NewInteraction(initiateURL, cfg)

		query, err := interaction.GetQuery()
		require.NoError(t, err)
		println("query", string(query))

		verifiablePres, err := credential.NewInquirer(ld.NewDocLoader()).
			Query(query, credential.NewCredentialsOpt(issuedCredentials))
		require.NoError(t, err)

		matchedCreds, err := verifiablePres.Credentials()
		require.NoError(t, err)

		require.Equal(t, 1, matchedCreds.Length())

		serializedIssuedVC, err := issuedCredentials.AtIndex(0).Serialize()
		require.NoError(t, err)

		serializedMatchedVC, err := matchedCreds.AtIndex(0).Serialize()
		require.NoError(t, err)
		println(serializedMatchedVC)

		require.Equal(t, serializedIssuedVC, serializedMatchedVC)

		verifiablePresContent, err := verifiablePres.Content()
		require.NoError(t, err)

		vm, err := testHelper.DIDDoc.AssertionMethod()
		require.NoError(t, err)

		err = interaction.PresentCredential(verifiablePresContent, vm)
		require.NoError(t, err)

		checkActivityLogAfterOpenID4VPFlow(t, activityLogger, tc.verifierProfileID)
		fmt.Printf("done test %d\n", i)
	}
}

func checkActivityLogAfterOpenID4VPFlow(t *testing.T, activityLogger *mem.ActivityLogger, verifierProfileID string) {
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

type vpTestHelper struct {
	KMS    *localkms.KMS
	DIDDoc *api.DIDDocResolution
}

func newVPTestHelper(t *testing.T, didMethod string) *vpTestHelper {
	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	// create DID
	c, err := did.NewCreatorWithKeyWriter(kms)
	require.NoError(t, err)

	didDoc, err := c.Create(didMethod, &api.CreateDIDOpts{})
	require.NoError(t, err)

	return &vpTestHelper{
		KMS:    kms,
		DIDDoc: didDoc,
	}
}

func (h *vpTestHelper) issueCredentials(t *testing.T, issuerProfileIDs []string) *api.VerifiableCredentialsArray {
	oidc4ciSetup, err := oidc4ci.NewSetup(testenv.NewHttpRequest())
	require.NoError(t, err)

	err = oidc4ciSetup.AuthorizeIssuerBypassAuth("test_org")
	require.NoError(t, err)

	credentials := api.NewVerifiableCredentialsArray()

	for i := 0; i < len(issuerProfileIDs); i++ {
		initiateIssuanceURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance(issuerProfileIDs[i])
		require.NoError(t, err)

		didResolver, err := did.NewResolver("")
		require.NoError(t, err)

		clientConfig := openid4ci.ClientConfig{
			ClientID:    "ClientID",
			DIDResolver: didResolver,
			Crypto:      h.KMS.GetCrypto(),
		}

		interaction, err := openid4ci.NewInteraction(initiateIssuanceURL, &clientConfig)
		require.NoError(t, err)

		authorizeResult, err := interaction.Authorize()
		require.NoError(t, err)
		require.False(t, authorizeResult.UserPINRequired)

		vm, err := h.DIDDoc.AssertionMethod()
		require.NoError(t, err)

		result, err := interaction.RequestCredential(&openid4ci.CredentialRequestOpts{}, vm)

		require.NoError(t, err)
		require.NotEmpty(t, result)

		for i := 0; i < result.Length(); i++ {
			vc := result.AtIndex(i)

			serializedVC, err := vc.Serialize()
			require.NoError(t, err)

			println(serializedVC)
			credentials.Add(result.AtIndex(i))
		}

	}

	return credentials
}
