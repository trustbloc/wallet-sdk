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
	//go:embed testdata/sample_issuer_metadata.jwt
	sampleIssuerMetadataJWT string
	//go:embed testdata/sample_issuer_metadata_with_order.jwt
	// Note that this sample is not properly signed - it exists just to test our handling of
	// the optional order field when received inside a JWT.
	sampleIssuerMetadataWithOrderJWT string
	//go:embed testdata/sample_jwt_without_issuer_metadata.jwt
	// This is the sample JWT taken directly from JWT.io.
	sampleJWTWithoutIssuerMetadata string
)

type mockIssuerServerHandler struct {
	issuerMetadata            string
	metadataRequestShouldFail bool
}

func (m *mockIssuerServerHandler) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	var err error

	if m.metadataRequestShouldFail {
		writer.WriteHeader(http.StatusInternalServerError)
		_, err = writer.Write([]byte("test failure"))
	} else {
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

			issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil,
				"", nil)
			require.NoError(t, err)
			require.NotNil(t, issuerMetadata)

			displayNameClaim, exists := issuerMetadata.CredentialsSupported[0].CredentialSubject["displayName"]
			require.True(t, exists)

			require.NotNil(t, displayNameClaim.Order)

			order, err := displayNameClaim.OrderAsInt()
			require.NoError(t, err)
			require.Equal(t, 3, order)

			jobTitleClaim, exists := issuerMetadata.CredentialsSupported[0].CredentialSubject["jobTitle"]
			require.True(t, exists)

			require.Nil(t, jobTitleClaim.Order)

			order, err = jobTitleClaim.OrderAsInt()
			require.EqualError(t, err, "order is nil or an unsupported type")
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

			displayNameClaim, exists := issuerMetadata.CredentialsSupported[0].CredentialSubject["displayName"]
			require.True(t, exists)

			require.NotNil(t, displayNameClaim.Order)

			order, err := displayNameClaim.OrderAsInt()
			require.NoError(t, err)
			require.Equal(t, 3, order)

			jobTitleClaim, exists := issuerMetadata.CredentialsSupported[0].CredentialSubject["jobTitle"]
			require.True(t, exists)

			require.Nil(t, jobTitleClaim.Order)

			order, err = jobTitleClaim.OrderAsInt()
			require.EqualError(t, err, "order is nil or an unsupported type")
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
	t.Run("Missing signature verifier", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: "invalid"}
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
		issuerServerHandler := &mockIssuerServerHandler{}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil,
			"", &mockVerifier{})
		require.EqualError(t, err, "failed to parse the response from the issuer's OpenID Credential "+
			"Issuer endpoint as JSON or as a JWT: unexpected end of JSON input\nJWT of compacted JWS form is "+
			"supported only")
		require.Nil(t, issuerMetadata)
	})
	t.Run("No issuer metadata in JWT", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: sampleJWTWithoutIssuerMetadata}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil,
			"", &mockVerifier{})
		require.EqualError(t, err, "issuer's OpenID configuration endpoint returned a JWT, but no "+
			"issuer metadata was detected (well_known_openid_issuer_configuration field is missing)")
		require.Nil(t, issuerMetadata)
	})
}

type mockVerifier struct{}

func (m *mockVerifier) CheckJWTProof(jose.Headers, string, []byte, []byte) error {
	return nil
}
