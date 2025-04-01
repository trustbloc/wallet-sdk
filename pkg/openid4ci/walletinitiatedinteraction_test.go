/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
	"github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

func TestWalletInitiatedInteractionFlow(t *testing.T) {
	t.Run("Token endpoint defined in the issuer's metadata", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:                  t,
			credentialResponse: sampleCredentialResponse,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerMetadata := strings.ReplaceAll(sampleIssuerMetadata, serverURLPlaceholder, server.URL)

		issuerServerHandler.issuerMetadata = modifyCredentialMetadata(t, issuerMetadata, func(m *issuer.Metadata) {
			m.RegistrationEndpoint = nil
		})

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
			"jwt_vc_json", types, openid4ci.WithIssuerState("issuerState"),
			openid4ci.WithCredentialContext([]string{"credentialContext"}))
		require.NoError(t, err)

		redirectURIWithParams := "redirectURI?code=1234&state=" + getStateFromAuthURL(t, authURL)

		credentials, err := interaction.RequestCredential(&jwtSignerMock{
			keyID: mockKeyID,
		}, redirectURIWithParams)
		require.NoError(t, err)
		require.Len(t, credentials, 1)
		require.NotEmpty(t, credentials[0])
	})
}

func TestWalletInitiatedInteraction_IssuerMetadata(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		issuerServerHandler.issuerMetadata = fmt.Sprintf(`{"credential_endpoint":"%s/credential",`+
			`"token_endpoint":"%s/oidc/token"}`, server.URL, server.URL)

		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewWalletInitiatedInteraction(server.URL, config)
		require.NoError(t, err)

		issuerMetadata, err := interaction.IssuerMetadata()
		require.NoError(t, err)
		require.NotNil(t, issuerMetadata)
	})
	t.Run("Fail to fetch issuer metadata", func(t *testing.T) {
		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewWalletInitiatedInteraction("", config)
		require.NoError(t, err)

		issuerMetadata, err := interaction.IssuerMetadata()
		require.EqualError(t, err, "METADATA_FETCH_FAILED(OCI1-0004):failed to get issuer metadata: "+
			"failed to get response from the issuer's metadata endpoint: "+
			`Get "/.well-known/openid-credential-issuer": unsupported protocol scheme ""`)
		require.Nil(t, issuerMetadata)
	})
}

func TestWalletInitiatedInteraction_VerifyIssuer(t *testing.T) {
	t.Run("", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:              t,
			issuerMetadata: "{}",
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		config := getTestClientConfig(t)

		interaction, err := openid4ci.NewWalletInitiatedInteraction(server.URL, config)
		require.NoError(t, err)

		serviceURL, err := interaction.VerifyIssuer()
		require.ErrorContains(t, err, "DID service validation failed")
		require.Empty(t, serviceURL)
	})
}
