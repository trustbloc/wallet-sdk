/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"net/http"

	noopactivitylogger "github.com/trustbloc/wallet-sdk/pkg/activitylogger/noop"
	"github.com/trustbloc/wallet-sdk/pkg/api"
	noopmetricslogger "github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
)

type opts struct {
	httpClient     httpClient
	activityLogger api.ActivityLogger
	metricsLogger  api.MetricsLogger
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

// WithMetricsLogger is an option for an OpenID4VP instance that allows a caller to specify their MetricsLogger.
// If used, then performance metrics events will be pushed to the given MetricsLogger implementation.
// If this option is not used, then metrics logging will be disabled.
func WithMetricsLogger(metricsLogger api.MetricsLogger) Opt {
	return func(opts *opts) {
		opts.metricsLogger = metricsLogger
	}
}

func processOpts(options []Opt) (
	httpClient,
	api.ActivityLogger,
	api.MetricsLogger,
) {
	opts := mergeOpts(options)

	if opts.httpClient == nil {
		opts.httpClient = &http.Client{Timeout: api.DefaultHTTPTimeout}
	}

	if opts.activityLogger == nil {
		opts.activityLogger = noopactivitylogger.NewActivityLogger()
	}

	if opts.metricsLogger == nil {
		opts.metricsLogger = noopmetricslogger.NewMetricsLogger()
	}

	return opts.httpClient, opts.activityLogger, opts.metricsLogger
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
