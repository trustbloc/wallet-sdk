/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"time"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// Opts contains all optional arguments that can be passed into the NewInteraction function.
type Opts struct {
	documentLoader                   api.LDDocumentLoader
	activityLogger                   api.ActivityLogger
	metricsLogger                    api.MetricsLogger
	additionalHeaders                api.Headers
	disableHTTPClientTLSVerification bool
	disableOpenTelemetry             bool
	httpTimeout                      *time.Duration
	kms                              *localkms.KMS
}

// NewOpts returns a new Opts object.
func NewOpts() *Opts {
	return &Opts{}
}

// SetDocumentLoader sets a document loader to use.
func (o *Opts) SetDocumentLoader(documentLoader api.LDDocumentLoader) *Opts {
	o.documentLoader = documentLoader

	return o
}

// SetActivityLogger sets an activity logger to use.
func (o *Opts) SetActivityLogger(activityLogger api.ActivityLogger) *Opts {
	o.activityLogger = activityLogger

	return o
}

// SetMetricsLogger sets a metrics logger to use.
func (o *Opts) SetMetricsLogger(metricsLogger api.MetricsLogger) *Opts {
	o.metricsLogger = metricsLogger

	return o
}

// SetHTTPTimeoutNanoseconds sets the timeout (in nanoseconds) for HTTP calls.
// Passing in 0 will disable timeouts.
func (o *Opts) SetHTTPTimeoutNanoseconds(timeout int64) *Opts {
	timeoutDuration := time.Duration(timeout)
	o.httpTimeout = &timeoutDuration

	return o
}

// AddHeaders adds the given HTTP headers to all REST calls made to the verifier during the OpenID4VP flow.
func (o *Opts) AddHeaders(headers *api.Headers) *Opts {
	headersAsArray := headers.GetAll()

	for i := range headersAsArray {
		o.additionalHeaders.Add(&headersAsArray[i])
	}

	return o
}

// AddHeader adds the given HTTP header to all REST calls made to the issuer during the OpenID4CI flow.
func (o *Opts) AddHeader(header *api.Header) *Opts {
	o.additionalHeaders.Add(header)

	return o
}

// DisableHTTPClientTLSVerify disables tls verification, should be used only for test purposes.
func (o *Opts) DisableHTTPClientTLSVerify() *Opts {
	o.disableHTTPClientTLSVerification = true

	return o
}

// DisableOpenTelemetry disables sending of open telemetry header.
func (o *Opts) DisableOpenTelemetry() *Opts {
	o.disableOpenTelemetry = true

	return o
}

// EnableAddingDIProofs enables the adding of data integrity proofs to presentations sent to the verifier. It requires
// a KMS to be passed in.
// Deprecated: DI proofs are now enabled by default. Their usage depends on the proof types supported by the verifier.
func (o *Opts) EnableAddingDIProofs(kms *localkms.KMS) *Opts {
	o.kms = kms

	return o
}

// NewPresentCredentialOpts returns a new PresentCredentialOpts object.
func NewPresentCredentialOpts() *PresentCredentialOpts {
	return &PresentCredentialOpts{}
}

// PresentCredentialOpts contains options for present credential operation.
type PresentCredentialOpts struct {
	scopeClaims map[string]string

	attestationVM                *api.VerificationMethod
	attestationVC                string
	serializedInteractionDetails string
}

// AddScopeClaim adds scope claim with given name.
func (o *PresentCredentialOpts) AddScopeClaim(claimName, claimJSON string) *PresentCredentialOpts {
	if o.scopeClaims == nil {
		o.scopeClaims = map[string]string{}
	}

	o.scopeClaims[claimName] = claimJSON

	return o
}

// SetAttestationVC is an option for the RequestCredentialWithPreAuth method that allows you to specify
// attestation VC, which may be required by the verifier.
func (o *PresentCredentialOpts) SetAttestationVC(
	vm *api.VerificationMethod, vc string,
) *PresentCredentialOpts {
	o.attestationVM = vm
	o.attestationVC = vc

	return o
}

// SetInteractionDetails extends authorization response with interaction details.
func (o *PresentCredentialOpts) SetInteractionDetails(
	serializedInteractionDetails string,
) *PresentCredentialOpts {
	o.serializedInteractionDetails = serializedInteractionDetails

	return o
}
