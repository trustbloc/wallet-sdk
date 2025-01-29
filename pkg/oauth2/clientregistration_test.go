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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/oauth2"
)

type mockIssuerServerHandler struct {
	t                             *testing.T
	returnInternalServerErrorCode bool
	writeEmptyBody                bool
}

func (m *mockIssuerServerHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if m.returnInternalServerErrorCode {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)

	if m.writeEmptyBody {
		return
	}

	response := oauth2.RegisterClientResponse{
		ClientID:           "Test",
		RegisteredMetadata: oauth2.RegisteredMetadata{ClientName: "ClientName"},
	}

	responseBytes, err := json.Marshal(response)
	assert.NoError(m.t, err)

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

		opt := oauth2.WithHTTPClient(&http.Client{})

		t.Run("Success - without initial access token", func(t *testing.T) {
			response, err := oauth2.RegisterClient(server.URL, &oauth2.ClientMetadata{}, opt)
			require.NoError(t, err)
			require.NotEmpty(t, response)
		})
		t.Run("Success - with initial access token", func(t *testing.T) {
			response, err := oauth2.RegisterClient(server.URL, &oauth2.ClientMetadata{},
				oauth2.WithInitialAccessBearerToken("token"))
			require.NoError(t, err)
			require.NotEmpty(t, response)
		})
	})
	t.Run("Blank registration endpoint", func(t *testing.T) {
		response, err := oauth2.RegisterClient("", nil)
		require.EqualError(t, err, "registration endpoint cannot be blank")
		require.Nil(t, response)
	})
	t.Run("Server doesn't return 201 status code", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:                             t,
			returnInternalServerErrorCode: true,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		response, err := oauth2.RegisterClient(server.URL, nil)
		require.ErrorContains(t, err, "expected status code 201 but got status code 500 with response body  instead")
		require.Nil(t, response)
	})
	t.Run("Server returns empty body, resulting in a JSON unmarshal failure", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{
			t:              t,
			writeEmptyBody: true,
		}

		server := httptest.NewServer(issuerServerHandler)
		defer server.Close()

		response, err := oauth2.RegisterClient(server.URL, nil)
		require.EqualError(t, err, "failed to unmarshal response body into a RegisterClientResponse: "+
			"unexpected end of JSON input")
		require.Nil(t, response)
	})
	t.Run("Fail to create HTTP request", func(t *testing.T) {
		response, err := oauth2.RegisterClient("%", nil)
		require.EqualError(t, err, `parse "%": invalid URL escape "%"`)
		require.Nil(t, response)
	})
}
