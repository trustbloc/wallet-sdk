/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package jwk contains a did:jwk creator implementation.
package jwk

import (
	"errors"
	"fmt"

	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"

	"github.com/trustbloc/did-go/doc/did"
	jwkvdr "github.com/trustbloc/did-go/method/jwk"
)

// ErrorModule is the error module name used for errors relating to did:jwk creation.
const ErrorModule = "DIDJWK"

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

// Create creates a new did:key document using the given verification method.
func Create(jsonWebKey *jwk.JWK) (*did.DocResolution, error) {
	if jsonWebKey == nil {
		return nil, walleterror.NewInvalidSDKUsageError(
			ErrorModule, errors.New("jwk object cannot be nil"))
	}

	vm, err := did.NewVerificationMethodFromJWK("#"+jsonWebKey.KeyID, "JsonWebKey2020", "", jsonWebKey)
	if err != nil {
		return nil, err
	}

	didDocArgument := &did.Doc{VerificationMethod: []did.VerificationMethod{*vm}}

	return jwkvdr.New().Create(didDocArgument)
}
