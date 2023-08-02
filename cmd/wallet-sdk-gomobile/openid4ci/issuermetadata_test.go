/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	_ "embed"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
)

//go:embed testdata/sample_issuer_metadata.json
var sampleIssuerMetadata string

func TestIssuerMetadata(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:              t,
		issuerMetadata: sampleIssuerMetadata,
	}
	server := httptest.NewServer(issuerServerHandler)

	defer server.Close()

	requestURI := createCredentialOfferIssuanceURI(t, server.URL, false)

	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	interactionRequiredArgs, interactionOptionalArgs := getTestArgs(t, requestURI, kms,
		nil, nil, false)

	interaction, err := openid4ci.NewIssuerInitiatedInteraction(interactionRequiredArgs, interactionOptionalArgs)
	require.NoError(t, err)

	issuerMetadata, err := interaction.IssuerMetadata()
	require.NoError(t, err)

	credentialIssuer := issuerMetadata.CredentialIssuer()
	require.Equal(t, "https://server.example.com", credentialIssuer)

	localizedIssuerDisplays := issuerMetadata.LocalizedIssuerDisplays()
	require.NotNil(t, localizedIssuerDisplays)
	require.Equal(t, 2, localizedIssuerDisplays.Length())

	firstLocalizedDisplay := localizedIssuerDisplays.AtIndex(0)
	require.Equal(t, "Example University", firstLocalizedDisplay.Name())
	require.Equal(t, "en-US", firstLocalizedDisplay.Locale())
	require.Equal(t, "https://server.example.com", firstLocalizedDisplay.URL())

	secondLocalizedDisplay := localizedIssuerDisplays.AtIndex(1)
	require.Equal(t, "サンプル大学", secondLocalizedDisplay.Name())
	require.Equal(t, "jp-JA", secondLocalizedDisplay.Locale())
	require.Equal(t, "https://server.example.com", secondLocalizedDisplay.URL())

	require.Nil(t, localizedIssuerDisplays.AtIndex(2))

	supportedCredentials := issuerMetadata.SupportedCredentials()
	require.NotNil(t, supportedCredentials)
	require.Equal(t, 1, supportedCredentials.Length())

	supportedCredential := supportedCredentials.AtIndex(0)
	require.NotNil(t, supportedCredential)
	require.Equal(t, "jwt_vc_json", supportedCredential.Format())
	require.Equal(t, "UniversityDegreeCredential", supportedCredential.ID())

	types := supportedCredential.Types()
	require.NotNil(t, types)
	require.Equal(t, 2, types.Length())
	require.Equal(t, "VerifiableCredential", types.AtIndex(0))
	require.Equal(t, "UniversityDegreeCredential", types.AtIndex(1))

	require.Nil(t, supportedCredentials.AtIndex(1))

	localizedDisplays := supportedCredential.LocalizedDisplays()
	require.NotNil(t, localizedDisplays)
	require.Equal(t, 1, localizedDisplays.Length())

	localizedDisplay := localizedDisplays.AtIndex(0)
	require.NotNil(t, localizedDisplay)
	require.Equal(t, "University Credential", localizedDisplay.Name())
	require.Equal(t, "en-US", localizedDisplay.Locale())
	require.Equal(t, "#12107c", localizedDisplay.BackgroundColor())
	require.Equal(t, "#FFFFFF", localizedDisplay.TextColor())

	credentialLogo := localizedDisplay.Logo()
	require.NotNil(t, credentialLogo)
	require.Equal(t, "https://exampleuniversity.com/public/logo.png", credentialLogo.URL())
	require.Equal(t, "a square logo of a university", credentialLogo.AltText())

	require.Nil(t, localizedDisplays.AtIndex(1))
}
