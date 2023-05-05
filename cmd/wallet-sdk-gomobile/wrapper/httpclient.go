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

// NewHTTPClient returns a new HTTP client using the given parameters.
func NewHTTPClient(timeout *time.Duration, additionalHeaders api.Headers,
	disableTLSVerification bool,
) *http.Client {
	if timeout == nil {
		defaultTimeout := goapi.DefaultHTTPTimeout

		timeout = &defaultTimeout
	}

	httpClient := &http.Client{
		Timeout: *timeout,
	}

	roundTripper := &headerInjectionRoundTripper{additionalHeaders: additionalHeaders.GetAll()}

	//nolint:gosec // The ability to disable TLS is an option we provide that has to be explicitly set by the user.
	// By default, we don't disable TLS. We have documentation specifying that this option is only intended for
	// testing purposes - it's up to the user to use this option appropriately.
	roundTripper.tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: disableTLSVerification}

	httpClient.Transport = roundTripper

	return httpClient
}

type headerInjectionRoundTripper struct {
	tlsConfig         *tls.Config
	additionalHeaders []api.Header
}

func (h *headerInjectionRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Per the docs for the http.RoundTripper RoundTrip method, a RoundTripper should not modify the request, so we
	// clone the request first and then inject the headers into the cloned request.
	clonedReq := req.Clone(req.Context())

	for _, additionalHeader := range h.additionalHeaders {
		clonedReq.Header.Add(additionalHeader.Name, additionalHeader.Value)
	}

	defaultTransport := &http.Transport{TLSClientConfig: h.tlsConfig}

	return defaultTransport.RoundTrip(clonedReq)
}
