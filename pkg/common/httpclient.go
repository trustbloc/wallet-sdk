/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"crypto/tls"
	"net/http"
)

// DefaultHTTPClient creates a http.Client with default parameters.
func DefaultHTTPClient() *http.Client {
	return http.DefaultClient
}

// InsecureHTTPClient creates a http.Client with disabled tls checks. Should be used only in tests.
func InsecureHTTPClient() *http.Client {
	//nolint:gosec //This can be ignored, becomes this function will be used only in tests.
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true}

	return &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
}
