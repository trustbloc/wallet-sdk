/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import "time"

// ValidateLinkedDomainsOpts contains all optional arguments that can be passed into the ValidateLinkedDomains function.
type ValidateLinkedDomainsOpts struct {
	httpTimeout *time.Duration
}

// NewValidateLinkedDomainsOpts returns a new ValidateLinkedDomainsOpts object.
func NewValidateLinkedDomainsOpts() *ValidateLinkedDomainsOpts {
	return &ValidateLinkedDomainsOpts{}
}

// SetHTTPTimeoutNanoseconds sets the timeout (in nanoseconds) for HTTP calls.
// Passing in 0 will disable timeouts.
func (o *ValidateLinkedDomainsOpts) SetHTTPTimeoutNanoseconds(timeout int64) *ValidateLinkedDomainsOpts {
	timeoutDuration := time.Duration(timeout)
	o.httpTimeout = &timeoutDuration

	return o
}
