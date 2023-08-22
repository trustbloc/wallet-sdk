/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package verifiable

import (
	"net/http"

	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
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
		httpClient := &http.Client{}

		if opts.httpTimeout != nil {
			httpClient.Timeout = *opts.httpTimeout
		} else {
			httpClient.Timeout = goapi.DefaultHTTPTimeout
		}

		parseCredentialOpts = append(parseCredentialOpts,
			verifiable.WithJSONLDDocumentLoader(ld.NewDefaultDocumentLoader(httpClient)))
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
