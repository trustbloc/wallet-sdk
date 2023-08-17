/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"time"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
)

// InteractionOpts contains all optional arguments that can be passed into the NewIssuerInitiatedInteraction function.
type InteractionOpts struct {
	activityLogger                   api.ActivityLogger
	metricsLogger                    api.MetricsLogger
	disableVCProofChecks             bool
	additionalHeaders                api.Headers
	disableHTTPClientTLSVerification bool
	documentLoader                   api.LDDocumentLoader
	disableOpenTelemetry             bool
	httpTimeout                      *time.Duration
	kms                              *localkms.KMS
}

// NewInteractionOpts returns a new InteractionOpts object.
func NewInteractionOpts() *InteractionOpts {
	return &InteractionOpts{}
}

// DisableVCProofChecks disables VC proof checks during the OpenID4CI interaction flow.
func (o *InteractionOpts) DisableVCProofChecks() *InteractionOpts {
	o.disableVCProofChecks = true

	return o
}

// SetHTTPTimeoutNanoseconds sets the timeout (in nanoseconds) for HTTP calls.
// Passing in 0 will disable timeouts.
func (o *InteractionOpts) SetHTTPTimeoutNanoseconds(timeout int64) *InteractionOpts {
	timeoutDuration := time.Duration(timeout)
	o.httpTimeout = &timeoutDuration

	return o
}

// AddHeaders adds the given HTTP headers to all REST calls made to the issuer during the OpenID4CI flow.
func (o *InteractionOpts) AddHeaders(headers *api.Headers) *InteractionOpts {
	headersAsArray := headers.GetAll()

	for i := range headersAsArray {
		o.additionalHeaders.Add(&headersAsArray[i])
	}

	return o
}

// AddHeader adds the given HTTP header to all REST calls made to the issuer during the OpenID4CI flow.
func (o *InteractionOpts) AddHeader(header *api.Header) *InteractionOpts {
	o.additionalHeaders.Add(header)

	return o
}

// DisableHTTPClientTLSVerify disables tls verification, should be used only for test purposes.
func (o *InteractionOpts) DisableHTTPClientTLSVerify() *InteractionOpts {
	o.disableHTTPClientTLSVerification = true

	return o
}

// SetDocumentLoader sets the document loader to use when parsing VCs received from the issuer.
// If no document loader is explicitly set, then a network-based loader will be used.
func (o *InteractionOpts) SetDocumentLoader(documentLoader api.LDDocumentLoader) *InteractionOpts {
	o.documentLoader = documentLoader

	return o
}

// SetActivityLogger sets an activity logger to be used for logging activities.
// If this option isn't used, then no activities will be logged.
func (o *InteractionOpts) SetActivityLogger(activityLogger api.ActivityLogger) *InteractionOpts {
	o.activityLogger = activityLogger

	return o
}

// SetMetricsLogger sets a metrics logger to use.
func (o *InteractionOpts) SetMetricsLogger(metricsLogger api.MetricsLogger) *InteractionOpts {
	o.metricsLogger = metricsLogger

	return o
}

// DisableOpenTelemetry disables sending of open telemetry header.
func (o *InteractionOpts) DisableOpenTelemetry() *InteractionOpts {
	o.disableOpenTelemetry = true

	return o
}

// EnableDIProofChecks enables data integrity proof checks for received VCs. It requires a KMS to be passed in.
func (o *InteractionOpts) EnableDIProofChecks(kms *localkms.KMS) *InteractionOpts {
	o.kms = kms

	return o
}
