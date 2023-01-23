/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"testing"

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
		issuerProfileID   string
		walletDIDMethod   string
		verifierProfileID string
		verifierDIDMethod string
	}

	tests := []test{
		{
			issuerProfileID:   "bank_issuer",
			walletDIDMethod:   "ion",
			verifierProfileID: "v_myprofile_jwt",
			verifierDIDMethod: "ion",
		},
		{
			issuerProfileID:   "bank_issuer",
			walletDIDMethod:   "key",
			verifierProfileID: "v_myprofile_jwt",
			verifierDIDMethod: "ion",
		},
		{
			issuerProfileID:   "bank_issuer",
			walletDIDMethod:   "jwk",
			verifierProfileID: "v_myprofile_jwt",
			verifierDIDMethod: "jwk",
		},
	}

	for _, tc := range tests {
		fmt.Println(fmt.Sprintf("running tests with issuerProfileID=%s walletDIDMethod=%s verifierDIDMethod=%s",
			tc.issuerProfileID, tc.walletDIDMethod, tc.verifierDIDMethod))

		testHelper := newVPTestHelper(t, tc.walletDIDMethod)

		issuedCredentials := testHelper.issueCredentials(t, tc.issuerProfileID)

		setup := oidc4vp.NewSetup(testenv.NewHttpRequest())

		err := setup.AuthorizeVerifierBypassAuth("test_org")
		require.NoError(t, err)

		initiateURL, err := setup.InitiateInteraction(tc.verifierProfileID)
		require.NoError(t, err)

		didResolver, err := did.NewResolver("")
		require.NoError(t, err)

		interaction := openid4vp.NewInteraction(
			initiateURL, testHelper.KMS, testHelper.KMS.GetCrypto(), didResolver, ld.NewDocLoader(), nil)

		query, err := interaction.GetQuery()
		require.NoError(t, err)

		verifiablePres, err := credential.NewInquirer(ld.NewDocLoader()).
			Query(query, credential.NewCredentialsOpt(issuedCredentials))
		require.NoError(t, err)

		matchedCreds, err := verifiablePres.Credentials()
		require.NoError(t, err)

		require.Equal(t, issuedCredentials.Length(), matchedCreds.Length())

		serializedIssuedVC, err := issuedCredentials.AtIndex(0).Serialize()
		require.NoError(t, err)

		serializedMatchedVC, err := matchedCreds.AtIndex(0).Serialize()
		require.NoError(t, err)

		require.Equal(t, serializedIssuedVC, serializedMatchedVC)

		verifiablePresContent, err := verifiablePres.Content()
		require.NoError(t, err)

		vm, err := testHelper.DIDDoc.AssertionMethod()
		require.NoError(t, err)

		err = interaction.PresentCredential(verifiablePresContent, vm)
		require.NoError(t, err)
	}
}

type vpTestHelper struct {
	KMS    *localkms.KMS
	DIDDoc *api.DIDDocResolution
}

func newVPTestHelper(t *testing.T, didMethod string) *vpTestHelper {
	kms, err := localkms.NewKMS(nil)
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

func (h *vpTestHelper) issueCredentials(t *testing.T, issuerProfileID string) *api.VerifiableCredentialsArray {
	oidc4ciSetup, err := oidc4ci.NewSetup(testenv.NewHttpRequest())
	require.NoError(t, err)

	err = oidc4ciSetup.AuthorizeIssuerBypassAuth("test_org")
	require.NoError(t, err)

	credentials := api.NewVerifiableCredentialsArray()

	for i := 0; i < 2; i++ {
		initiateIssuanceURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance(issuerProfileID)
		require.NoError(t, err)

		signerCreator, err := localkms.CreateSignerCreator(h.KMS)
		require.NoError(t, err)

		didResolver, err := did.NewResolver("")
		require.NoError(t, err)

		didID, err := h.DIDDoc.ID()
		require.NoError(t, err)

		clientConfig := openid4ci.ClientConfig{
			UserDID:       didID,
			ClientID:      "ClientID",
			SignerCreator: signerCreator,
			DIDResolver:   didResolver,
		}

		interaction, err := openid4ci.NewInteraction(initiateIssuanceURL, &clientConfig)
		require.NoError(t, err)

		authorizeResult, err := interaction.Authorize()
		require.NoError(t, err)
		require.False(t, authorizeResult.UserPINRequired)

		result, err := interaction.RequestCredential(&openid4ci.CredentialRequestOpts{})

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
