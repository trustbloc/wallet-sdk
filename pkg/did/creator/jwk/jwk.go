/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package jwk contains a did:jwk creator implementation.
package jwk

import (
	"fmt"

	"github.com/trustbloc/did-go/doc/did"
	jwkvdr "github.com/trustbloc/did-go/method/jwk"
)

// Creator creates did:jwk DID Documents.
type Creator struct {
	vdr *jwkvdr.VDR
}

// NewCreator initializes a did:jwk DID creator.
func NewCreator() *Creator {
	return &Creator{
		vdr: jwkvdr.New(),
	}
}

// Create creates a did:jwk DID Doc from a given JWK.
func (creator *Creator) Create(vm *did.VerificationMethod) (*did.DocResolution, error) {
	docRes, err := creator.vdr.Create(&did.Doc{
		VerificationMethod: []did.VerificationMethod{
			*vm,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("creating did:jwk DID Document: %w", err)
	}

	return docRes, nil
}
