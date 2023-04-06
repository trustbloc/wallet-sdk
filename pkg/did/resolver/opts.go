/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package resolver

import "time"

type opts struct {
	resolverServerURI string
	httpTimeout       *time.Duration
}

// An Opt is a single option for a Resolver instance.
type Opt func(opts *opts)

// WithResolverServerURI provides a URI for a DID resolution server.
func WithResolverServerURI(resolverServerURI string) Opt {
	return func(opts *opts) {
		opts.resolverServerURI = resolverServerURI
	}
}

// WithHTTPTimeout sets a timeout for HTTP calls.
// Only used if a resolver server URI has been set.
// Passing in 0 will disable timeouts.
func WithHTTPTimeout(timeout time.Duration) Opt {
	return func(opts *opts) {
		opts.httpTimeout = &timeout
	}
}

func mergeOpts(options []Opt) *opts {
	resolveOpts := &opts{}

	for _, opt := range options {
		if opt != nil {
			opt(resolveOpts)
		}
	}

	return resolveOpts
}
