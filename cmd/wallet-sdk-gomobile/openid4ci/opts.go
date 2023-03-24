/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// Opts contains all optional arguments that can be passed into the NewInteraction function.
type Opts struct {
	activityLogger                   api.ActivityLogger
	metricsLogger                    api.MetricsLogger
	disableVCProofChecks             bool
	additionalHeaders                api.Headers
	disableHTTPClientTLSVerification bool
	documentLoader                   api.LDDocumentLoader
}

// NewOpts returns a new Opts object.
func NewOpts() *Opts {
	return &Opts{}
}

// DisableVCProofChecks disables VC proof checks during the OpenID4CI interaction flow.
func (o *Opts) DisableVCProofChecks() {
	o.disableVCProofChecks = true
}

// AddHeaders adds the given HTTP headers to all REST calls made to the issuer during the OpenID4CI flow.
func (o *Opts) AddHeaders(headers *api.Headers) {
	headersAsArray := headers.GetAll()

	for i := range headersAsArray {
		o.additionalHeaders.Add(&headersAsArray[i])
	}
}

// DisableHTTPClientTLSVerify disables tls verification, should be used only for test purposes.
func (o *Opts) DisableHTTPClientTLSVerify() {
	o.disableHTTPClientTLSVerification = true
}

// SetDocumentLoader sets the document loader to use when parsing VCs received from the issuer.
// If no document loader is explicitly set, then a network-based loader will be used.
func (o *Opts) SetDocumentLoader(documentLoader api.LDDocumentLoader) {
	o.documentLoader = documentLoader
}

// SetActivityLogger sets an activity logger to be used for logging activities.
// If this option isn't used, then no activities will be logged.
func (o *Opts) SetActivityLogger(activityLogger api.ActivityLogger) {
	o.activityLogger = activityLogger
}

// SetMetricsLogger sets a metrics logger to use.
func (o *Opts) SetMetricsLogger(metricsLogger api.MetricsLogger) {
	o.metricsLogger = metricsLogger
}
