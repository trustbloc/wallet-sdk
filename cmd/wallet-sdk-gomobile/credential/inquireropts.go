/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"time"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// InquirerOpts contain all optionals arguments that can be passed into the NewInquirer function.
type InquirerOpts struct {
	documentLoader api.LDDocumentLoader
	httpTimeout    *time.Duration
	didResolver    api.DIDResolver
}

// NewInquirerOpts returns a new InquirerOpts object.
func NewInquirerOpts() *InquirerOpts {
	return &InquirerOpts{}
}

// SetDocumentLoader sets the document loader to use.
// If no document loader is explicitly set, then a network-based loader will be used.
func (o *InquirerOpts) SetDocumentLoader(documentLoader api.LDDocumentLoader) *InquirerOpts {
	o.documentLoader = documentLoader

	return o
}

// SetHTTPTimeoutNanoseconds sets the timeout (in nanoseconds) for HTTP calls made by the default network-based
// document loader. This option is only used if no document loader was explicitly set via the SetDocumentLoader option.
// Passing in 0 will disable timeouts.
func (o *InquirerOpts) SetHTTPTimeoutNanoseconds(timeout int64) *InquirerOpts {
	timeoutDuration := time.Duration(timeout)
	o.httpTimeout = &timeoutDuration

	return o
}

// SetDIDResolver sets the did resolver that required of some implementations of selective disclosure.
func (o *InquirerOpts) SetDIDResolver(didResolver api.DIDResolver) *InquirerOpts {
	o.didResolver = didResolver

	return o
}
