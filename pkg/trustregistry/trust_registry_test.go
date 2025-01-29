/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/
package trustregistry_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/trustregistry"
	"github.com/trustbloc/wallet-sdk/pkg/trustregistry/testsupport"
)

const (
	evaluateIssuanceURL     = "/wallet/interactions/issuance"
	evaluatePresentationURL = "/wallet/interactions/presentation"
)

func TestRegistry_EvaluateIssuance(t *testing.T) {
	serverHandler := &mockTrustRegistryHandler{}

	server := httptest.NewServer(serverHandler)
	defer server.Close()

	registry := trustregistry.New(&trustregistry.RegistryConfig{
		EvaluateIssuanceURL: server.URL + evaluateIssuanceURL,
	})

	t.Run("Success", func(t *testing.T) {
		result, err := registry.EvaluateIssuance(&trustregistry.IssuanceRequest{
			IssuerDID: "did:web:correct.com",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.Allowed)
	})

	t.Run("Forbidden", func(t *testing.T) {
		result, err := registry.EvaluateIssuance(&trustregistry.IssuanceRequest{
			IssuerDID: "did:web:forbidden.com",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.False(t, result.Allowed)
		require.Equal(t, "didForbidden", result.ErrorCode)
	})

	t.Run("Invalid server URI", func(t *testing.T) {
		result, err := trustregistry.New(&trustregistry.RegistryConfig{
			EvaluateIssuanceURL: "http://invalid",
		}).EvaluateIssuance(&trustregistry.IssuanceRequest{
			IssuerDID: "did:web:forbidden.com",
		})

		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("Invalid server URI", func(t *testing.T) {
		result, err := trustregistry.New(&trustregistry.RegistryConfig{
			EvaluateIssuanceURL: server.URL + "/invalid-json",
		}).EvaluateIssuance(&trustregistry.IssuanceRequest{})

		require.Error(t, err)
		require.Nil(t, result)
	})
}

func TestRegistry_EvaluatePresentation(t *testing.T) {
	serverHandler := &mockTrustRegistryHandler{}

	server := httptest.NewServer(serverHandler)
	defer server.Close()

	registry := trustregistry.New(&trustregistry.RegistryConfig{
		EvaluatePresentationURL: server.URL + evaluatePresentationURL,
	})

	t.Run("Success", func(t *testing.T) {
		result, err := registry.EvaluatePresentation(&trustregistry.PresentationRequest{
			VerifierDid: "did:web:correct.com",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.Allowed)
	})

	t.Run("Forbidden", func(t *testing.T) {
		result, err := registry.EvaluatePresentation(&trustregistry.PresentationRequest{
			VerifierDid: "did:web:forbidden.com",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.False(t, result.Allowed)
		require.Equal(t, "didForbidden", result.ErrorCode)
	})

	t.Run("Invalid server URI", func(t *testing.T) {
		result, err := trustregistry.New(&trustregistry.RegistryConfig{
			EvaluatePresentationURL: "http://invalid",
		}).EvaluatePresentation(&trustregistry.PresentationRequest{
			VerifierDid: "did:web:forbidden.com",
		})

		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("Invalid server URI", func(t *testing.T) {
		result, err := trustregistry.New(&trustregistry.RegistryConfig{
			EvaluatePresentationURL: server.URL + "/invalid-json",
		}).EvaluatePresentation(&trustregistry.PresentationRequest{})

		require.Error(t, err)
		require.Nil(t, result)
	})
}

type mockTrustRegistryHandler struct{}

func (m *mockTrustRegistryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "invalid-json") {
		w.Header().Set("Content-Type", "application/json")

		_, err := w.Write([]byte("`````"))
		if err != nil {
			log.Printf("Write failed: %v", err)
		}

		w.WriteHeader(http.StatusOK)

		return
	}

	if strings.HasSuffix(r.URL.Path, evaluateIssuanceURL) {
		testsupport.HandleEvaluateIssuanceRequest(w, r)

		return
	}

	if strings.HasSuffix(r.URL.Path, evaluatePresentationURL) {
		testsupport.HandleEvaluatePresentationRequest(w, r)

		return
	}

	http.Error(w, "Unexpected handler", http.StatusInternalServerError)
}
