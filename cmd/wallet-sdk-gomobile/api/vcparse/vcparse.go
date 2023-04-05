/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

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

// Parse parses the given serialized VC into a VC object.
func Parse(vc string, opts *Opts) (*api.VerifiableCredential, error) {
	if opts == nil {
		opts = &Opts{}
	}

	var parseCredentialOpts []verifiable.CredentialOpt

	if opts.disableProofCheck {
		parseCredentialOpts = append(parseCredentialOpts, verifiable.WithDisabledProofCheck())
	}

	if opts.documentLoader == nil {
		parseCredentialOpts = append(parseCredentialOpts,
			verifiable.WithJSONLDDocumentLoader(ld.NewDefaultDocumentLoader(http.DefaultClient)))
	} else {
		wrappedLoader := &wrapper.DocumentLoaderWrapper{
			DocumentLoader: opts.documentLoader,
		}

		parseCredentialOpts = append(parseCredentialOpts, verifiable.WithJSONLDDocumentLoader(wrappedLoader))
	}

	verifiableCredential, err := verifiable.ParseCredential([]byte(vc), parseCredentialOpts...)
	if err != nil {
		return nil, err
	}

	return api.NewVerifiableCredential(verifiableCredential), nil
}
