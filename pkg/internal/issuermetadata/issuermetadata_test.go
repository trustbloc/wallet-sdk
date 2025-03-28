/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuermetadata_test

import (
	_ "embed"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trustbloc/kms-go/doc/jose"
	"github.com/trustbloc/vc-go/proof/defaults"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
	"github.com/trustbloc/wallet-sdk/pkg/internal/issuermetadata"
)

var (
	//go:embed testdata/sample_issuer_metadata.json
	sampleIssuerMetadataJSON string
	//go:embed testdata/sample_issuer_metadata.jwt.json
	sampleIssuerMetadataJWT string
	//go:embed testdata/sample_issuer_metadata_with_order.jwt.json
	// Note that this sample is not properly signed - it exists just to test our handling of
	// the optional order field when received inside a JWT.
	sampleIssuerMetadataWithOrderJWT string
	//go:embed testdata/sample_jwt_without_issuer_metadata.jwt.json
	// This is the sample JWT taken directly from JWT.io.
	sampleJWTWithoutIssuerMetadata string
)

type mockIssuerServerHandler struct {
	issuerMetadata            string
	metadataRequestShouldFail bool
	invalidResponseObject     bool
}

func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	var err error

	switch {
	case m.metadataRequestShouldFail:
		writer.WriteHeader(http.StatusInternalServerError)
		_, err = writer.Write([]byte("test failure"))
	case m.invalidResponseObject:
		_, err = writer.Write([]byte(`["response1","response2"]`))
	default:
		_, err = writer.Write([]byte(m.issuerMetadata))
	}

	if err != nil {
		println(err.Error())
	}
}

type failingMetricsLogger struct{}

func (f *failingMetricsLogger) Log(metricsEvent *api.MetricsEvent) error {
	return fmt.Errorf("failed to log event (Event=%s)", metricsEvent.Event)
}

func TestGet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Parsing from JSON", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: sampleIssuerMetadataJSON}
			server := httptest.NewServer(issuerServerHandler)

			defer server.Close()

			issuerMetadata, err := issuermetadata.Get(server.URL, nil, nil,
				"", nil)
			require.NoError(t, err)
			require.NotNil(t, issuerMetadata)

			credentialConf := issuerMetadata.CredentialConfigurationsSupported["VerifiedEmployee_ldp_vc_v1"]
			_, exists := credentialConf.CredentialDefinition.CredentialSubject["displayName"]
			require.True(t, exists)

			order, err := credentialConf.ClaimOrderAsInt("displayName")
			require.NoError(t, err)
			require.Equal(t, 3, order)

			_, exists = credentialConf.CredentialDefinition.CredentialSubject["jobTitle"]
			require.True(t, exists)

			order, err = credentialConf.ClaimOrderAsInt("jobTitle")
			require.EqualError(t, err, "order is not specified")
			require.Equal(t, -1, order)
		})
		t.Run("Parsing from JWT", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: sampleIssuerMetadataJWT}
			server := httptest.NewServer(issuerServerHandler)

			defer server.Close()

			didResolver, err := resolver.NewDIDResolver()
			require.NoError(t, err)

			jwtVerifier := defaults.NewDefaultProofChecker(common.NewVDRKeyResolver(didResolver))

			issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil,
				"", jwtVerifier)
			require.NoError(t, err)
			require.NotNil(t, issuerMetadata)
		})
		t.Run("Parsing from JWT with an order field (mock verifier)", func(t *testing.T) {
			issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: sampleIssuerMetadataWithOrderJWT}
			server := httptest.NewServer(issuerServerHandler)

			defer server.Close()

			// For this sample, the signature is not actually valid, so we need to pass in a mock verifier.
			issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil,
				"", &mockVerifier{})
			require.NoError(t, err)
			require.NotNil(t, issuerMetadata)

			credentialConf := issuerMetadata.CredentialConfigurationsSupported["VerifiedEmployee_ldp_vc_v1"]
			credentialDefinition := credentialConf.CredentialDefinition

			_, exists := credentialDefinition.CredentialSubject["displayName"]
			require.True(t, exists)

			order, err := credentialConf.ClaimOrderAsInt("displayName")
			require.NoError(t, err)
			require.Equal(t, 0, order)

			_, exists = credentialDefinition.CredentialSubject["jobTitle"]
			require.True(t, exists)

			order, err = credentialConf.ClaimOrderAsInt("jobTitle")
			require.EqualError(t, err, "order is not specified")
			require.Equal(t, -1, order)
		})
	})
	t.Run("Fail to reach issuer OpenID config endpoint", func(t *testing.T) {
		issuerMetadata, err := issuermetadata.Get("http://BadURL", http.DefaultClient, nil,
			"", nil)
		require.Contains(t, err.Error(), `Get "http://BadURL/.well-known/openid-credential-issuer":`+
			` dial tcp: lookup BadURL`)
		require.Nil(t, issuerMetadata)
	})
	t.Run("Fail to get issuer metadata: server error", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{metadataRequestShouldFail: true}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil,
			"", nil)
		require.Contains(t, err.Error(), "failed to get response from the issuer's metadata endpoint: "+
			"expected status code 200 but got status code 500 with response body test failure instead")
		require.Nil(t, issuerMetadata)
	})
	t.Run("Fail to decode issuer metadata", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{invalidResponseObject: true}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil,
			"", nil)
		require.Contains(t, err.Error(), "decode metadata")
		require.Nil(t, issuerMetadata)
	})
	t.Run("Missing signature verifier", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: `{"signed_metadata": "a.b"}`}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil, "",
			nil)
		require.Contains(t, err.Error(), "missing signature verifier")
		require.Nil(t, issuerMetadata)
	})
	t.Run("Fail to log metrics event", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: sampleIssuerMetadataJSON}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, &failingMetricsLogger{},
			"", nil)
		require.Contains(t, err.Error(), "failed to log event (Event=Fetch issuer metadata via an HTTP GET "+
			"request to http://127.0.0.1:")
		require.Nil(t, issuerMetadata)
	})
	t.Run("Fail to parse", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: `{"signed_metadata": "a.b"}`}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil,
			"", &mockVerifier{})
		require.EqualError(t, err, "failed to parse the response from the issuer's OpenID Credential "+
			"Issuer endpoint as JSON or as a JWT: JWT of compacted JWS form is "+
			"supported only")
		require.Nil(t, issuerMetadata)
	})
	t.Run("No issuer metadata in JWT", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: sampleJWTWithoutIssuerMetadata}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil,
			"", &mockVerifier{})
		require.NoError(t, err)
		require.NotNil(t, issuerMetadata)
	})
}

type mockVerifier struct{}

func (m *mockVerifier) CheckJWTProof(jose.Headers, string, []byte, []byte) error {
	return nil
}
