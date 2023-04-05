/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package wrapper contains wrappers that convert between the Go and gomobile APIs.
package wrapper

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// CredentialReaderWrapper wraps a gomobile-compatible version of api.CredentialReader and translates methods calls to
// their corresponding Go API versions.
type CredentialReaderWrapper struct {
	CredentialReader api.CredentialReader
}

// Get wraps Get of api.CredentialReader.
func (r *CredentialReaderWrapper) Get(id string) (*verifiable.Credential, error) {
	vc, err := r.CredentialReader.Get(id)
	if err != nil {
		return nil, err
	}

	return vc.VC, nil
}

// GetAll wraps GetAll of api.CredentialReader.
func (r *CredentialReaderWrapper) GetAll() ([]*verifiable.Credential, error) {
	vcs, err := r.CredentialReader.GetAll()
	if err != nil {
		return nil, err
	}

	var credentials []*verifiable.Credential

	for i := 0; i < vcs.Length(); i++ {
		credentials = append(credentials, vcs.AtIndex(i).VC)
	}

	return credentials, nil
}
