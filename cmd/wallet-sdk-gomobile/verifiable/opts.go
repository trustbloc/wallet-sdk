/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package verifiable

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// Opts contains all optional arguments that can be passed into the Parse function.
type Opts struct {
	disableProofCheck bool
	documentLoader    api.LDDocumentLoader
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
func (o *Opts) SetDocumentLoader(documentLoader api.LDDocumentLoader) *Opts {
	o.documentLoader = documentLoader

	return o
}
