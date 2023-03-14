/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/ld"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4vp"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/helpers"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/metricslogger"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/setup/oidc4vp"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/testenv"
)

func TestOpenID4VPFullFlow(t *testing.T) {
	type test struct {
		issuerProfileIDs  []string
		walletDIDMethod   string
		verifierProfileID string
		signingKeyType    string
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
			signingKeyType:    localkms.KeyTypeP384,
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
		{
			issuerProfileIDs:  []string{"bank_issuer", "bank_issuer"},
			walletDIDMethod:   "ion",
			verifierProfileID: "v_myprofile_jwt_verified_employee",
		},
	}

	for i, tc := range tests {
		fmt.Printf("running test %d: issuerProfileIDs=%s verifierProfileID=%s "+
			"walletDIDMethod=%s\n", i,
			tc.issuerProfileIDs, tc.verifierProfileID, tc.walletDIDMethod)

		testHelper := helpers.NewVPTestHelper(t, tc.walletDIDMethod, tc.signingKeyType)

		issuedCredentials := testHelper.IssueCredentials(t, vcsAPIDirectURL, tc.issuerProfileIDs)
		println("Issued", issuedCredentials.Length(), "credentials")
		for k := 0; k < issuedCredentials.Length(); k++ {
			cred, _ := issuedCredentials.AtIndex(k).Serialize()
			println("Issued VC[", k, "]: ", cred)
		}

		setup := oidc4vp.NewSetup(testenv.NewHttpRequest())

		err := setup.AuthorizeVerifierBypassAuth("test_org", vcsAPIDirectURL)
		require.NoError(t, err)

		initiateURL, err := setup.InitiateInteraction(tc.verifierProfileID)
		require.NoError(t, err)

		didResolver, err := did.NewResolver(didResolverURL)
		require.NoError(t, err)

		activityLogger := mem.NewActivityLogger()

		docLoader := ld.NewDocLoader()

		metricsLogger := metricslogger.NewMetricsLogger()

		cfg := openid4vp.NewClientConfig(
			testHelper.KMS, testHelper.KMS.GetCrypto(), didResolver, docLoader, activityLogger)
		cfg.MetricsLogger = metricsLogger

		interaction := openid4vp.NewInteraction(initiateURL, cfg)

		query, err := interaction.GetQuery()
		require.NoError(t, err)
		println("query", string(query))

		inquirer := credential.NewInquirer(docLoader)
		require.NoError(t, err)

		requirements, err := inquirer.GetSubmissionRequirements(query, credential.NewCredentialsOpt(issuedCredentials))
		require.GreaterOrEqual(t, requirements.Len(), 1)
		require.GreaterOrEqual(t, requirements.AtIndex(0).DescriptorLen(), 1)

		requirementDescriptor := requirements.AtIndex(0).DescriptorAtIndex(0)
		require.GreaterOrEqual(t, requirementDescriptor.MatchedVCs.Length(), 1)

		selectedCreds := api.NewVerifiableCredentialsArray()
		selectedCreds.Add(requirementDescriptor.MatchedVCs.AtIndex(0))

		verifiablePres, err := inquirer.Query(query, credential.NewCredentialsOpt(selectedCreds))
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

		err = interaction.PresentCredential(selectedCreds)
		require.NoError(t, err)

		testHelper.CheckActivityLogAfterOpenID4VPFlow(t, activityLogger, tc.verifierProfileID)
		testHelper.CheckMetricsLoggerAfterOpenID4VPFlow(t, metricsLogger)

		fmt.Printf("done test %d\n", i)
	}
}
