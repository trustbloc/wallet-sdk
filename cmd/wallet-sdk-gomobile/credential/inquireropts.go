/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// InquirerOpts contain all optionals arguments that can be passed into the NewInquirer function.
type InquirerOpts struct {
	documentLoader api.LDDocumentLoader
}

// NewInquirerOpts returns a new InquirerOpts object.
func NewInquirerOpts() *InquirerOpts {
	return &InquirerOpts{}
}

// SetDocumentLoader sets the document loader to use.
// If no document loader is explicitly set, then a network-based loader will be used.
func (o *InquirerOpts) SetDocumentLoader(documentLoader api.LDDocumentLoader) {
	o.documentLoader = documentLoader
}
