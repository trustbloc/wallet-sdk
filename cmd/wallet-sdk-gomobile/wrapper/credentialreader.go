/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package wrapper contains wrappers that convert between the Go and gomobile APIs.
package wrapper

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// CredentialReaderWrapper wraps a gomobile-compatible version of api.CredentialReader and translates methods calls to
// their corresponding Go API versions.
type CredentialReaderWrapper struct {
	CredentialReader api.CredentialReader
	DocumentLoader   ld.DocumentLoader
}

// Get wraps Get of api.CredentialReader.
func (r *CredentialReaderWrapper) Get(id string) (*verifiable.Credential, error) {
	vc, err := r.CredentialReader.Get(id)
	if err != nil {
		return nil, err
	}

	cred, err := verifiable.ParseCredential(vc.Content, verifiable.WithDisabledProofCheck(),
		verifiable.WithJSONLDDocumentLoader(r.DocumentLoader))
	if err != nil {
		return nil, fmt.Errorf("verifiable credential parse failed: %w", err)
	}

	return cred, nil
}

// GetAll wraps GetAll of api.CredentialReader.
func (r *CredentialReaderWrapper) GetAll() ([]*verifiable.Credential, error) {
	vcs, err := r.CredentialReader.GetAll()
	if err != nil {
		return nil, err
	}

	var credentials []*verifiable.Credential

	for i := 0; i < vcs.Length(); i++ {
		cred, credErr := verifiable.ParseCredential(vcs.AtIndex(i).Content, verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(r.DocumentLoader))
		if credErr != nil {
			return nil, fmt.Errorf("verifiable credential parse failed: %w", credErr)
		}

		credentials = append(credentials, cred)
	}

	return credentials, nil
}
