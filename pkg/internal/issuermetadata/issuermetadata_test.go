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

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/internal/issuermetadata"
)

//go:embed testdata/sample_issuer_metadata.json
var sampleIssuerMetadata string

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
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: sampleIssuerMetadata}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil, "")
		require.NoError(t, err)
		require.NotNil(t, issuerMetadata)
	})
	t.Run("Fail to reach issuer OpenID config endpoint", func(t *testing.T) {
		issuerMetadata, err := issuermetadata.Get("http://BadURL", http.DefaultClient, nil, "")
		require.Contains(t, err.Error(), `Get "http://BadURL/.well-known/openid-credential-issuer":`+
			` dial tcp: lookup BadURL`)
		require.Nil(t, issuerMetadata)
	})
	t.Run("Fail to get issuer metadata: server error", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{metadataRequestShouldFail: true}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil, "")
		require.Contains(t, err.Error(), "openid configuration endpoint: "+
			"expected status code 200 but got status code 500 with response body test failure instead")
		require.Nil(t, issuerMetadata)
	})
	t.Run("Fail to unmarshal response from issuer OpenID config endpoint", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: "invalid"}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, nil, "")
		require.Contains(t, err.Error(), "failed to unmarshal response from the issuer's OpenID "+
			"configuration endpoint: invalid character 'i' looking for beginning of value")
		require.Nil(t, issuerMetadata)
	})
	t.Run("Fail to log metrics event", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: sampleIssuerMetadata}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL, http.DefaultClient, &failingMetricsLogger{}, "")
		require.Contains(t, err.Error(), "failed to log event (Event=Fetch issuer metadata via an HTTP GET "+
			"request to http://127.0.0.1:")
		require.Nil(t, issuerMetadata)
	})
}
