/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

import (
	"net/http"
)

// DefaultHTTPClient creates an http.Client with default parameters.
func DefaultHTTPClient() *http.Client {
	return http.DefaultClient
}
