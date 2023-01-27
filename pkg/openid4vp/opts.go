/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"github.com/trustbloc/wallet-sdk/pkg/activitylogger/noop"
	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
)

type opts struct {
	httpClient     httpClient
	activityLogger api.ActivityLogger
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

// WithActivityLogger is an option for an OpenID4VP instance that allows a caller to specify their ActivityLogger.
// The caller can check their ActivityLogger after credential(s) are presented to see what activities took place.
// If not specified, then credential activity will not be logged.
func WithActivityLogger(activityLogger api.ActivityLogger) Opt {
	return func(opts *opts) {
		opts.activityLogger = activityLogger
	}
}

func processOpts(options []Opt) (httpClient, api.ActivityLogger) {
	opts := mergeOpts(options)

	if opts.httpClient == nil {
		opts.httpClient = common.DefaultHTTPClient()
	}

	if opts.activityLogger == nil {
		opts.activityLogger = noop.NewActivityLogger()
	}

	return opts.httpClient, opts.activityLogger
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
