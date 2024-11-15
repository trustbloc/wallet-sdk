/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oauth2_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/oauth2"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	goapioauth2 "github.com/trustbloc/wallet-sdk/pkg/oauth2"
)

type mockIssuerServerHandler struct {
	t                   *testing.T
	emptyResponseObject bool
}

func (m *mockIssuerServerHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	testIssuedAtAndExpiresAtValue := 10

	var response goapioauth2.RegisterClientResponse

	if m.emptyResponseObject {
		response = goapioauth2.RegisterClientResponse{}
	} else {
		response = goapioauth2.RegisterClientResponse{
			ClientID:              "ClientID",
			ClientSecret:          "ClientSecret",
			ClientIDIssuedAt:      &testIssuedAtAndExpiresAtValue,
			ClientSecretExpiresAt: &testIssuedAtAndExpiresAtValue,
			RegisteredMetadata: goapioauth2.RegisteredMetadata{
				RedirectURIs:            []string{"RedirectURI1"},
				TokenEndpointAuthMethod: "TokenEndpointAuthMethod",
				GrantTypes:              []string{"GrantType1"},
				ResponseTypes:           []string{"ResponseType1"},
				ClientName:              "ClientName",
				ClientURI:               "ClientURI",
				LogoURI:                 "LogoURI",
				Scope:                   "scope1 scope2",
				Contacts:                []string{"Contact1"},
				TOSURI:                  "TOSURI",
				PolicyURI:               "PolicyURI",
				JWKSetURI:               "JWKSetURI",
				JWKSet:                  &goapi.JSONWebKeySet{},
				SoftwareID:              "SoftwareID",
				SoftwareVersion:         "SoftwareVersion",
			},
		}
	}

	responseBytes, err := json.Marshal(response)
	assert.NoError(m.t, err)

	w.WriteHeader(http.StatusCreated)

	_, err = w.Write(responseBytes)
	assert.NoError(m.t, err)
}

func TestRegisterClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t: t,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		opts := oauth2.NewRegisterClientOpts()
		opts.SetHTTPTimeoutNanoseconds(int64(time.Minute))
		opts.DisableHTTPClientTLSVerify()

		testHeader := api.NewHeader("test", "test")
		testHeaders := api.NewHeaders()
		testHeaders.Add(testHeader)

		opts.AddHeaders(testHeaders)
		opts.AddHeader(testHeader)

		t.Run("Success - without initial access token", func(t *testing.T) {
			response, err := oauth2.RegisterClient(server.URL, &oauth2.ClientMetadata{}, opts)
			require.NoError(t, err)
			require.Equal(t, "ClientID", response.ClientID())
			require.Equal(t, "ClientSecret", response.ClientSecret())

			require.True(t, response.HasClientIDIssuedAt())
			clientIDIssuedAt, err := response.ClientIDIssuedAt()
			require.NoError(t, err)
			require.Equal(t, 10, clientIDIssuedAt)

			require.True(t, response.HasClientSecretExpiresAt())
			clientSecretExpiresAt, err := response.ClientSecretExpiresAt()
			require.NoError(t, err)
			require.Equal(t, 10, clientSecretExpiresAt)

			clientMetadata := response.RegisteredMetadata()

			require.Equal(t, 1, clientMetadata.RedirectURIs().Length())
			require.Equal(t, "RedirectURI1", clientMetadata.RedirectURIs().AtIndex(0))
			require.Equal(t, "TokenEndpointAuthMethod", clientMetadata.TokenEndpointAuthMethod())
			require.Equal(t, 1, clientMetadata.GrantTypes().Length())
			require.Equal(t, "GrantType1", clientMetadata.GrantTypes().AtIndex(0))
			require.Equal(t, 1, clientMetadata.ResponseTypes().Length())
			require.Equal(t, "ResponseType1", clientMetadata.ResponseTypes().AtIndex(0))
			require.Equal(t, "ClientName", clientMetadata.ClientName())
			require.Equal(t, "ClientURI", clientMetadata.ClientURI())
			require.Equal(t, "LogoURI", clientMetadata.LogoURI())
			require.Equal(t, 2, clientMetadata.Scopes().Length())
			require.Equal(t, "scope1", clientMetadata.Scopes().AtIndex(0))
			require.Equal(t, "scope2", clientMetadata.Scopes().AtIndex(1))
			require.Equal(t, 1, clientMetadata.Contacts().Length())
			require.Equal(t, "Contact1", clientMetadata.Contacts().AtIndex(0))
			require.Equal(t, "TOSURI", clientMetadata.TOSURI())
			require.Equal(t, "PolicyURI", clientMetadata.PolicyURI())
			require.Equal(t, "JWKSetURI", clientMetadata.JWKSetURI())
			require.NotNil(t, clientMetadata.JWKSet())
			require.Equal(t, "SoftwareID", clientMetadata.SoftwareID())
			require.Equal(t, "SoftwareVersion", clientMetadata.SoftwareVersion())
		})
		t.Run("Success - with initial access token", func(t *testing.T) {
			response, err := oauth2.RegisterClient(server.URL, &oauth2.ClientMetadata{},
				oauth2.NewRegisterClientOpts().SetInitialAccessBearerToken("token"))
			require.NoError(t, err)
			require.NotEmpty(t, response)
		})
	})
	t.Run("Blank registration endpoint", func(t *testing.T) {
		response, err := oauth2.RegisterClient("", nil, nil)
		require.EqualError(t, err, "registration endpoint cannot be blank")
		require.Nil(t, response)
	})
	t.Run("Empty response object", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:                   t,
			emptyResponseObject: true,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		response, err := oauth2.RegisterClient(server.URL, nil, nil)
		require.NoError(t, err, "registration endpoint cannot be blank")

		require.False(t, response.HasClientIDIssuedAt())

		clientIDIssuedAt, err := response.ClientIDIssuedAt()
		require.EqualError(t, err, "the register client response object does not "+
			"specify when the client ID was issued")
		require.Equal(t, -1, clientIDIssuedAt)

		clientSecretExpiresAt, err := response.ClientSecretExpiresAt()
		require.EqualError(t, err, "the register client response object does not "+
			"specify when the client secret expires")
		require.Equal(t, -1, clientSecretExpiresAt)
	})
}
