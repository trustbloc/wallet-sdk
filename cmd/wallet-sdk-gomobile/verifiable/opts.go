/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package verifiable

import (
	"time"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// Opts contains all optional arguments that can be passed into the Parse function.
type Opts struct {
	disableProofCheck                bool
	documentLoader                   api.LDDocumentLoader
	httpTimeout                      *time.Duration
	additionalHeaders                api.Headers
	disableHTTPClientTLSVerification bool
}

// NewOpts returns a new Opts object for use with the Parse function.
func NewOpts() *Opts {
	return &Opts{}
}

// DisableProofCheck disables the proof check that normally happens when parsing the VC.
func (o *Opts) DisableProofCheck() *Opts {
	o.disableProofCheck = true

	return o
}

// SetDocumentLoader sets the document loader to use while parsing the VC.
// If this option isn't used, then a network-based document loader will be used.
func (o *Opts) SetDocumentLoader(documentLoader api.LDDocumentLoader) *Opts {
	o.documentLoader = documentLoader

	return o
}

// SetHTTPTimeoutNanoseconds sets the timeout (in nanoseconds) for HTTP calls made by the default network-based
// document loader. This option is only used if no document loader was explicitly set via the SetDocumentLoader option.
// Passing in 0 will disable timeouts.
func (o *Opts) SetHTTPTimeoutNanoseconds(timeout int64) *Opts {
	timeoutDuration := time.Duration(timeout)
	o.httpTimeout = &timeoutDuration

	return o
}

// AddHeader adds the given HTTP header to all REST calls made by network-based document loader
// In case SetDocumentLoader is used - this option does not affect http calls.
func (o *Opts) AddHeader(header *api.Header) *Opts {
	o.additionalHeaders.Add(header)

	return o
}

// DisableHTTPClientTLSVerify disables TLS verification. Should be used for testing purposes only.
func (o *Opts) DisableHTTPClientTLSVerify() *Opts {
	o.disableHTTPClientTLSVerification = true

	return o
}
