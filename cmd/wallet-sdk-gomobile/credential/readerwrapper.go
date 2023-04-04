/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
)

// ReaderWrapper wraps a gomobile-compatible version of Reader and translates methods calls to
// their corresponding Go API versions.
type ReaderWrapper struct {
	CredentialReader Reader
}

// Get wraps Get of api.CredentialReader.
func (r *ReaderWrapper) Get(id string) (*verifiable.Credential, error) {
	vc, err := r.CredentialReader.Get(id)
	if err != nil {
		return nil, err
	}

	return vc.VC, nil
}

// GetAll wraps GetAll of api.CredentialReader.
func (r *ReaderWrapper) GetAll() ([]*verifiable.Credential, error) {
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
