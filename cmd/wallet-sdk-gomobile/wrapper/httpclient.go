/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

// HTTPClient is the HTTP client implementation that gets injected in from the gomobile layer.
// It makes use of the optional HTTP client parameters that are exposed from the various APIs.
type HTTPClient struct {
	additionalHeaders      []api.Header
	DisableTLSVerification bool // Should only be set to true for testing purposes.
	Timeout                *time.Duration
}

// NewHTTPClient returns a new HTTPClient.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{}
}

// AddHeaders adds the given headers to the list of headers that will get added to all HTTP calls made by this
// client implementation.
func (m *HTTPClient) AddHeaders(headers *api.Headers) {
	m.additionalHeaders = append(m.additionalHeaders, headers.GetAll()...)
}

// Do sends the given HTTP request and returns an HTTP response. If this HTTP client has any additional headers to be
// injected (from previous calls to AddHeaders), then those headers are added to the request before being executed
// by the default Go HTTP client.
func (m *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	for _, additionalHeader := range m.additionalHeaders {
		req.Header.Add(additionalHeader.Name, additionalHeader.Value)
	}

	httpClient := http.Client{}

	if m.DisableTLSVerification {
		//nolint:gosec // The ability to disable TLS is an option we provide that has to be explicitly set by the user.
		// By default, we don't disable TLS. We have documentation specifying that this option is only intended for
		// testing purposes - it's up to the user to use this option appropriately.
		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true}

		httpClient.Transport = &http.Transport{TLSClientConfig: tlsConfig}
	}

	if m.Timeout != nil {
		httpClient.Timeout = *m.Timeout
	} else {
		httpClient.Timeout = goapi.DefaultHTTPTimeout
	}

	return httpClient.Do(req)
}
