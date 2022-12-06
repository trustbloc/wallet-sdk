/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"fmt"
	"testing"

	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didcreator"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/didresolver"
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

func TestFullFlow(t *testing.T) {
	oidc4ciSetup, err := oidc4ci.NewSetup(testenv.NewHttpRequest())
	require.NoError(t, err)

	err = oidc4ciSetup.AuthorizeIssuer("test_org")
	require.NoError(t, err)

	initiateIssuanceURL, err := oidc4ciSetup.InitiatePreAuthorizedIssuance("bank_issuer")
	require.NoError(t, err)

	kms, err := localkms.NewKMS(nil)
	require.NoError(t, err)

	// create DID
	c, err := didcreator.NewCreatorWithKeyWriter(kms)
	require.NoError(t, err)

	didDoc, err := c.Create("key", &api.CreateDIDOpts{})
	require.NoError(t, err)

	fmt.Println(string(didDoc))

	signerCreator, err := localkms.CreateSignerCreator(kms)
	require.NoError(t, err)

	didResolver := didresolver.NewDIDResolver()

	didDocResolutionParsed, err := diddoc.ParseDocumentResolution(didDoc)
	require.NoError(t, err)

	clientConfig := openid4ci.ClientConfig{
		UserDID:       didDocResolutionParsed.DIDDocument.ID,
		ClientID:      "ClientID",
		SignerCreator: signerCreator,
		DIDResolver:   didResolver,
	}

	interaction, err := openid4ci.NewInteraction(initiateIssuanceURL, &clientConfig)
	require.NoError(t, err)

	authorizeResult, err := interaction.Authorize()
	require.NoError(t, err)
	require.False(t, authorizeResult.UserPINRequired)

	credential, err := interaction.RequestCredential(&openid4ci.CredentialRequestOpts{})

	require.NoError(t, err)
	require.NotEmpty(t, credential)

	println("credential:", string(credential))
}
