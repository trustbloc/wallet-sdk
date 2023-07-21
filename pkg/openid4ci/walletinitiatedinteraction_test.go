/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/trustbloc/wallet-sdk/pkg/openid4ci"

	"github.com/stretchr/testify/require"
)

func TestWalletInitiatedInteraction(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{
		t:                  t,
		credentialResponse: sampleCredentialResponse,
	}

	server := httptest.NewServer(issuerServerHandler)
	defer server.Close()

	issuerServerHandler.openIDConfig = &openid4ci.OpenIDConfig{
		TokenEndpoint: fmt.Sprintf("%s/oidc/token", server.URL),
	}

	issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential"}`,
		server.URL)

	config := getTestClientConfig(t)

	interaction, err := openid4ci.NewWalletInitiatedInteraction(server.URL, config)
	require.NoError(t, err)

	supportedCredentials, err := interaction.SupportedCredentials()
	require.NoError(t, err)
	require.NotNil(t, supportedCredentials)

	dynamicClientRegistrationSupported, err := interaction.DynamicClientRegistrationSupported()
	require.NoError(t, err)
	require.False(t, dynamicClientRegistrationSupported)

	dynamicClientRegistrationEndpoint, err := interaction.DynamicClientRegistrationEndpoint()
	require.EqualError(t, err,
		"INVALID_SDK_USAGE(OCI3-0000):issuer does not support dynamic client registration")
	require.Empty(t, dynamicClientRegistrationEndpoint)

	types := []string{"VerifiableCredential", "VerifiedEmployee"}

	// Needed to create the OAuth2 config object.
	authURL, err := interaction.CreateAuthorizationURL("clientID", "redirectURI",
		"jwt_vc_json", types, openid4ci.WithIssuerState("issuerState"))
	require.NoError(t, err)

	redirectURIWithParams := "redirectURI?code=1234&state=" + getStateFromAuthURL(t, authURL)

	credentials, err := interaction.RequestCredential(&jwtSignerMock{
		keyID: mockKeyID,
	}, redirectURIWithParams)
	require.NoError(t, err)
	require.Len(t, credentials, 1)
	require.NotEmpty(t, credentials[0])
}
