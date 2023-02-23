/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuermetadata_test

import (
	_ "embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

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

func TestGet(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: sampleIssuerMetadata}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL)
		require.NoError(t, err)
		require.NotNil(t, issuerMetadata)
	})
	t.Run("Fail to reach issuer OpenID config endpoint", func(t *testing.T) {
		issuerMetadata, err := issuermetadata.Get("http://BadURL")
		require.Contains(t, err.Error(), `Get "http://BadURL/.well-known/openid-credential-issuer":`+
			` dial tcp: lookup BadURL:`)
		require.Nil(t, issuerMetadata)
	})
	t.Run("Fail to get issuer metadata: server error", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{metadataRequestShouldFail: true}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL)
		require.Contains(t, err.Error(), "received status code [500] with body [test failure] from "+
			"issuer's OpenID credential issuer endpoint")
		require.Nil(t, issuerMetadata)
	})
	t.Run("Fail to unmarshal response from issuer OpenID config endpoint", func(t *testing.T) {
		issuerServerHandler := &mockIssuerServerHandler{issuerMetadata: "invalid"}
		server := httptest.NewServer(issuerServerHandler)

		defer server.Close()

		issuerMetadata, err := issuermetadata.Get(server.URL)
		require.Contains(t, err.Error(), "failed to unmarshal response from the issuer's OpenID "+
			"configuration endpoint: invalid character 'i' looking for beginning of value")
		require.Nil(t, issuerMetadata)
	})
}
