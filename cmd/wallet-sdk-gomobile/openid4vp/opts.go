/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// Opts contains all optional arguments that can be passed into the NewInteraction function.
type Opts struct {
	documentLoader                   api.LDDocumentLoader
	activityLogger                   api.ActivityLogger
	metricsLogger                    api.MetricsLogger
	additionalHeaders                api.Headers
	disableHTTPClientTLSVerification bool
	disableOpenTelemetry             bool
}

// NewOpts returns a new Opts object.
func NewOpts() *Opts {
	return &Opts{}
}

// SetDocumentLoader sets a document loader to use.
func (o *Opts) SetDocumentLoader(documentLoader api.LDDocumentLoader) {
	o.documentLoader = documentLoader
}

// SetActivityLogger sets an activity logger to use.
func (o *Opts) SetActivityLogger(activityLogger api.ActivityLogger) {
	o.activityLogger = activityLogger
}

// SetMetricsLogger sets a metrics logger to use.
func (o *Opts) SetMetricsLogger(metricsLogger api.MetricsLogger) {
	o.metricsLogger = metricsLogger
}

// AddHeaders adds the given HTTP headers to all REST calls made to the verifier during the OpenID4VP flow.
func (o *Opts) AddHeaders(headers *api.Headers) {
	headersAsArray := headers.GetAll()

	for i := range headersAsArray {
		o.additionalHeaders.Add(&headersAsArray[i])
	}
}

// AddHeader adds the given HTTP header to all REST calls made to the issuer during the OpenID4CI flow.
func (o *Opts) AddHeader(header *api.Header) {
	o.additionalHeaders.Add(header)
}

// DisableHTTPClientTLSVerify disables tls verification, should be used only for test purposes.
func (o *Opts) DisableHTTPClientTLSVerify() {
	o.disableHTTPClientTLSVerification = true
}

// DisableOpenTelemetry disables sending of open telemetry header.
func (o *Opts) DisableOpenTelemetry() *Opts {
	o.disableOpenTelemetry = true

	return o
}
