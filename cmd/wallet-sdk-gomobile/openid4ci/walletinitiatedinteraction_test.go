/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	arieskms "github.com/trustbloc/kms-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	goapiopenid4ci "github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

func TestWalletInitiatedInteraction_Flow(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:                  t,
		credentialResponse: sampleCredentialResponse,
	}
	server := httptest.NewServer(issuerServerHandler)

	issuerServerHandler.openIDConfig = &goapiopenid4ci.OpenIDConfig{
		TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
	}

	issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential",`+
		`"credential_issuer":"https://server.example.com",`+
		`"credentials_supported":[{"format":"jwt_vc_json","types":["VerifiableCredential",`+
		`"UniversityDegreeCredential"]},{"format":"ldp_vc","types":["VerifiableCredential",`+
		`"DriversLicenseCredential"]}]}`, server.URL)

	defer server.Close()

	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	resolver := &mockResolver{keyWriter: kms}

	requiredArgs := openid4ci.NewWalletInitiatedInteractionArgs(server.URL, kms.GetCrypto(), resolver)

	opts := openid4ci.NewInteractionOpts()
	opts.DisableVCProofChecks()

	interaction, err := openid4ci.NewWalletInitiatedInteraction(requiredArgs, opts)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	issuerMetadata, err := interaction.IssuerMetadata()
	require.NoError(t, err)
	require.NotNil(t, issuerMetadata)

	supportedCredentials := issuerMetadata.SupportedCredentials()
	require.Equal(t, 2, supportedCredentials.Length())

	require.Equal(t, "jwt_vc_json", supportedCredentials.AtIndex(0).Format())
	require.Equal(t, 2, supportedCredentials.AtIndex(0).Types().Length())
	require.Equal(t, "VerifiableCredential", supportedCredentials.AtIndex(0).Types().AtIndex(0))
	require.Equal(t, "UniversityDegreeCredential", supportedCredentials.AtIndex(0).Types().AtIndex(1))

	require.Equal(t, "ldp_vc", supportedCredentials.AtIndex(1).Format())
	require.Equal(t, 2, supportedCredentials.AtIndex(1).Types().Length())
	require.Equal(t, "VerifiableCredential", supportedCredentials.AtIndex(1).Types().AtIndex(0))
	require.Equal(t, "DriversLicenseCredential", supportedCredentials.AtIndex(1).Types().AtIndex(1))

	dynamicClientRegistrationSupported, err := interaction.DynamicClientRegistrationSupported()
	require.NoError(t, err)
	require.False(t, dynamicClientRegistrationSupported)

	dynamicClientRegistrationEndpoint, err := interaction.DynamicClientRegistrationEndpoint()
	requireErrorContains(t, err, "INVALID_SDK_USAGE")
	requireErrorContains(t, err, "issuer does not support dynamic client registration")
	require.Empty(t, dynamicClientRegistrationEndpoint)

	credentialTypes := api.NewStringArray().Append("type")

	createAuthorizationURLOpts := openid4ci.NewCreateAuthorizationURLOpts().SetIssuerState("IssuerState")

	authURL, err := interaction.CreateAuthorizationURL("client", "redirectURI",
		"format", credentialTypes, createAuthorizationURLOpts)
	require.NoError(t, err)
	require.NotEmpty(t, authURL)

	keyHandle, err := kms.Create(arieskms.ED25519)
	require.NoError(t, err)

	pkBytes, err := keyHandle.JWK.PublicKeyBytes()
	require.NoError(t, err)

	redirectURIWithParams := "redirectURI?code=1234&state=" + getStateFromAuthURL(t, authURL)

	result, err := interaction.RequestCredential(&api.VerificationMethod{
		ID:   "did:example:12345#testId",
		Type: "Ed25519VerificationKey2018",
		Key:  models.VerificationKey{Raw: pkBytes},
	}, redirectURIWithParams, nil)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestNewWalletInitiatedInteraction(t *testing.T) {
	interaction, err := openid4ci.NewWalletInitiatedInteraction(nil, nil)
	requireErrorContains(t, err, "INVALID_SDK_USAGE")
	requireErrorContains(t, err, "args object must be provided")
	require.Nil(t, interaction)
}

func TestWalletInitiatedInteraction_DynamicClientRegistrationSupported_Failure(t *testing.T) {
	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	resolver := &mockResolver{keyWriter: kms}

	requiredArgs := openid4ci.NewWalletInitiatedInteractionArgs("", kms.GetCrypto(), resolver)

	interaction, err := openid4ci.NewWalletInitiatedInteraction(requiredArgs, nil)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	supported, err := interaction.DynamicClientRegistrationSupported()
	requireErrorContains(t, err, "ISSUER_OPENID_CONFIG_FETCH_FAILED")
	require.False(t, supported)
}

func TestWalletInitiatedInteraction_RequestCredential_Failure(t *testing.T) {
	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	resolver := &mockResolver{keyWriter: kms}

	requiredArgs := openid4ci.NewWalletInitiatedInteractionArgs("", kms.GetCrypto(), resolver)

	interaction, err := openid4ci.NewWalletInitiatedInteraction(requiredArgs, nil)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	credentials, err := interaction.RequestCredential(nil, "", nil)
	requireErrorContains(t, err, "INVALID_SDK_USAGE")
	requireErrorContains(t, err, "verification method must be provided")
	require.Nil(t, credentials)
}

func getStateFromAuthURL(t *testing.T, authURL string) string {
	t.Helper()

	parsedURI, err := url.Parse(authURL)
	require.NoError(t, err)

	return parsedURI.Query().Get("state")
}
