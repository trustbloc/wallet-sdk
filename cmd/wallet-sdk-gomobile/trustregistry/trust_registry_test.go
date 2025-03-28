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

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4vp"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/trustregistry"
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

	registryConfig := &trustregistry.RegistryConfig{
		EvaluateIssuanceURL:        server.URL + evaluateIssuanceURL,
		DisableHTTPClientTLSVerify: true,
	}

	registryConfig.AddHeader(api.NewHeader("X-Trace", "request1"))

	registry := trustregistry.NewRegistry(registryConfig)

	t.Run("Success", func(t *testing.T) {
		result, err := registry.EvaluateIssuance(&trustregistry.IssuanceRequest{
			IssuerDID: "did:web:correct.com",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.Allowed)
	})

	t.Run("Success: with credential offers", func(t *testing.T) {
		issuanceRequest := &trustregistry.IssuanceRequest{
			IssuerDID: "did:web:correct.com",
		}

		issuanceRequest.AddCredentialOffers(&trustregistry.CredentialOffer{})

		result, err := registry.EvaluateIssuance(issuanceRequest)

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
		require.Equal(t, "unauthorized issuer, empty credentials", result.DenyReason())
	})

	t.Run("Invalid server URI", func(t *testing.T) {
		result, err := trustregistry.NewRegistry(&trustregistry.RegistryConfig{
			EvaluateIssuanceURL: "http://invalid",
		}).EvaluateIssuance(&trustregistry.IssuanceRequest{
			IssuerDID: "did:web:forbidden.com",
		})

		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("Invalid server URI", func(t *testing.T) {
		result, err := trustregistry.NewRegistry(&trustregistry.RegistryConfig{
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

	registry := trustregistry.NewRegistry(&trustregistry.RegistryConfig{
		EvaluatePresentationURL: server.URL + evaluatePresentationURL,
	})

	t.Run("Success", func(t *testing.T) {
		result, err := registry.EvaluatePresentation((&trustregistry.PresentationRequest{
			VerifierDID: "did:web:correct.com",
		}).AddCredentialClaims(trustregistry.LegacyNewCredentialClaimsToCheck(
			"test_id",
			api.NewStringArray().Append("TestType"),
			"issuer_id", 0, 0),
		).AddCredentialClaims(&trustregistry.CredentialClaimsToCheck{
			CredentialID:        "cred2",
			CredentialTypes:     api.NewStringArray().Append("type2"),
			CredentialClaimKeys: &openid4vp.CredentialClaimKeys{},
		}))

		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.Allowed)
		require.Equal(t, 1, result.RequestedAttestationLength())
		require.Equal(t, "attestation1", result.RequestedAttestationAtIndex(0))
	})

	t.Run("Forbidden", func(t *testing.T) {
		result, err := registry.EvaluatePresentation(&trustregistry.PresentationRequest{
			VerifierDID: "did:web:forbidden.com",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.False(t, result.Allowed)
		require.Equal(t, "didForbidden", result.ErrorCode)
	})

	t.Run("Invalid server URI", func(t *testing.T) {
		result, err := trustregistry.NewRegistry(&trustregistry.RegistryConfig{
			EvaluatePresentationURL: "http://invalid",
		}).EvaluatePresentation(&trustregistry.PresentationRequest{
			VerifierDID: "did:web:forbidden.com",
		})

		require.Error(t, err)
		require.Nil(t, result)
	})

	t.Run("Invalid server URI", func(t *testing.T) {
		result, err := trustregistry.NewRegistry(&trustregistry.RegistryConfig{
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
		testsupport.HandleEvaluateIssuanceRequestWithAttestation(w, r)

		return
	}

	if strings.HasSuffix(r.URL.Path, evaluatePresentationURL) {
		testsupport.HandleEvaluatePresentationRequestWithAttestation(w, r)

		return
	}

	http.Error(w, "Unexpected handler", http.StatusInternalServerError)
}
