/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package vcparse contains a function for parsing Verifiable Credentials from a serialized format into the VC type
// used in the mobile bindings.
package vcparse

import (
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
)

// Opts represents various options for the Parse function.
type Opts struct {
	DisableProofCheck bool
	DocumentLoader    api.LDDocumentLoader
}

// NewOpts returns a new Opts object for use with the Parse function.
func NewOpts(disableProofCheck bool, documentLoader api.LDDocumentLoader) *Opts {
	return &Opts{
		DisableProofCheck: disableProofCheck,
		DocumentLoader:    documentLoader,
	}
}

// Parse parses the given serialized VC into a VC object.
func Parse(vc string, opts *Opts) (*api.VerifiableCredential, error) {
	if opts == nil {
		opts = &Opts{}
	}

	var parseCredentialOpts []verifiable.CredentialOpt

	if opts.DisableProofCheck {
		parseCredentialOpts = append(parseCredentialOpts, verifiable.WithDisabledProofCheck())
	}

	if opts.DocumentLoader == nil {
		parseCredentialOpts = append(parseCredentialOpts,
			verifiable.WithJSONLDDocumentLoader(ld.NewDefaultDocumentLoader(http.DefaultClient)))
	} else {
		wrappedLoader := &wrapper.DocumentLoaderWrapper{
			DocumentLoader: opts.DocumentLoader,
		}

		parseCredentialOpts = append(parseCredentialOpts, verifiable.WithJSONLDDocumentLoader(wrappedLoader))
	}

	verifiableCredential, err := verifiable.ParseCredential([]byte(vc), parseCredentialOpts...)
	if err != nil {
		return nil, err
	}

	return api.NewVerifiableCredential(verifiableCredential), nil
}
