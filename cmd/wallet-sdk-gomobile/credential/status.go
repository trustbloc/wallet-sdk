/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"net/http"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/pkg/credentialstatus"
)

// StatusVerifier verifies credential status.
type StatusVerifier struct {
	verifier *credentialstatus.Verifier
}

// NewStatusVerifier creates a credential status verifier.
func NewStatusVerifier() (*StatusVerifier, error) {
	v, err := credentialstatus.NewVerifier(&credentialstatus.Config{
		HTTPClient: http.DefaultClient,
	})
	if err != nil {
		return nil, err
	}

	return &StatusVerifier{
		verifier: v,
	}, nil
}

// Verify verifies credential status.
func (s *StatusVerifier) Verify(vc *api.VerifiableCredential) error {
	return s.verifier.Verify(vc.VC)
}
