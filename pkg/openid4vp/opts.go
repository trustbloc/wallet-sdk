/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/log/consolelogger"
)

type opts struct {
	httpClient httpClient
	logger     api.Logger
}

// An Opt is a single option for an OpenID4VP instance.
type Opt func(opts *opts)

// WithHTTPClient is an option for an OpenID4VP instance that allows a caller to specify their own HTTP client
// implementation.
func WithHTTPClient(httpClient httpClient) Opt {
	return func(opts *opts) {
		opts.httpClient = httpClient
	}
}

// WithLogger is an option for an OpenID4VP instance that allows a caller to specify their own logger implementation.
func WithLogger(logger api.Logger) Opt {
	return func(opts *opts) {
		opts.logger = logger
	}
}

func processOpts(options []Opt) (httpClient, api.Logger) {
	opts := mergeOpts(options)

	if opts.httpClient == nil {
		opts.httpClient = common.DefaultHTTPClient()
	}

	if opts.logger == nil {
		opts.logger = consolelogger.NewConsoleLogger()
	}

	return opts.httpClient, opts.logger
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
