/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package verifiable

import (
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/memstorage/legacy"
)

// ParseCredential parses the given serialized VC into a VC object.
func ParseCredential(vc string, opts *Opts) (*Credential, error) {
	if opts == nil {
		opts = &Opts{}
	}

	var parseCredentialOpts []verifiable.CredentialOpt

	if opts.disableProofCheck {
		parseCredentialOpts = append(parseCredentialOpts, verifiable.WithDisabledProofCheck())
	}

	if opts.documentLoader == nil {
		httpClient := wrapper.NewHTTPClient(opts.httpTimeout, opts.additionalHeaders, opts.disableHTTPClientTLSVerification)

		goAPIDocumentLoader, err := common.CreateJSONLDDocumentLoader(httpClient, legacy.NewProvider())
		if err != nil {
			return nil, wrapper.ToMobileError(err)
		}

		parseCredentialOpts = append(parseCredentialOpts,
			verifiable.WithJSONLDDocumentLoader(goAPIDocumentLoader))
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

	return NewCredential(verifiableCredential), nil
}
