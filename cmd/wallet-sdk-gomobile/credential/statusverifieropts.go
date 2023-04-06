/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import "time"

// StatusVerifierOpts contains optional parameters for initializing a credential StatusVerifier.
type StatusVerifierOpts struct {
	httpTimeout *time.Duration
}

// NewStatusVerifierOpts returns a StatusVerifierOpts object.
func NewStatusVerifierOpts() *StatusVerifierOpts {
	return &StatusVerifierOpts{}
}

// SetHTTPTimeoutNanoseconds sets the timeout (in nanoseconds) for HTTP calls.
// Passing in 0 will disable timeouts.
func (o *StatusVerifierOpts) SetHTTPTimeoutNanoseconds(timeout int64) *StatusVerifierOpts {
	timeoutDuration := time.Duration(timeout)
	o.httpTimeout = &timeoutDuration

	return o
}
