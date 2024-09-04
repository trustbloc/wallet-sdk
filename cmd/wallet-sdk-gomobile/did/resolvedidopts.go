/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import "time"

// ResolverOpts contains all optional arguments that can be passed into the ResolveDID method.
type ResolverOpts struct {
	resolverServerURI                string
	httpTimeout                      *time.Duration
	disableHTTPClientTLSVerification bool
}

// NewResolverOpts returns a new ResolverOpts object.
func NewResolverOpts() *ResolverOpts {
	return &ResolverOpts{}
}

// SetResolverServerURI sets a resolver server to use when resolving certain types of DIDs.
func (c *ResolverOpts) SetResolverServerURI(resolverServerURI string) *ResolverOpts {
	c.resolverServerURI = resolverServerURI

	return c
}

// SetHTTPTimeoutNanoseconds sets the timeout (in nanoseconds) for HTTP calls.
// Only used if a resolver server URI was set.
// Passing in 0 will disable timeouts.
func (c *ResolverOpts) SetHTTPTimeoutNanoseconds(timeout int64) *ResolverOpts {
	timeoutDuration := time.Duration(timeout)
	c.httpTimeout = &timeoutDuration

	return c
}

// DisableHTTPClientTLSVerify disables tls verification, should be used only for test purposes.
func (c *ResolverOpts) DisableHTTPClientTLSVerify() *ResolverOpts {
	c.disableHTTPClientTLSVerification = true

	return c
}
