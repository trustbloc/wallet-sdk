/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"fmt"

	diddoc "github.com/trustbloc/did-go/doc/did"

	"github.com/trustbloc/wallet-sdk/pkg/models"
)

// VerificationMethod represents a DID verification method.
type VerificationMethod struct {
	ID   string
	Type string
	Key  models.VerificationKey
}

// ToSDKVerificationMethod returns this VerificationMethod as a models.VerificationMethod.
func (vm *VerificationMethod) ToSDKVerificationMethod() *models.VerificationMethod {
	return &models.VerificationMethod{
		ID:   vm.ID,
		Type: vm.Type,
		Key:  vm.Key,
	}
}

// NewVerificationMethod creates VerificationMethod.
func NewVerificationMethod(keyID, vmType string) *VerificationMethod {
	return &VerificationMethod{ID: keyID, Type: vmType}
}

// DIDDocResolution represents a DID document resolution object.
type DIDDocResolution struct {
	// Content is the full marshalled DID doc resolution object.
	Content string
}

// NewDIDDocResolution creates a new DIDDocResolution.
func NewDIDDocResolution(content string) *DIDDocResolution {
	return &DIDDocResolution{
		Content: content,
	}
}

// ID returns the ID of the DID document contained within this DIDDocResolution.
func (d *DIDDocResolution) ID() (string, error) {
	didDocResolutionParsed, err := diddoc.ParseDocumentResolution([]byte(d.Content))
	if err != nil {
		return "", fmt.Errorf("failed to parse DID document resolution content: %w", err)
	}

	return didDocResolutionParsed.DIDDocument.ID, nil
}

// AssertionMethod returns did assertion verification method.
func (d *DIDDocResolution) AssertionMethod() (*VerificationMethod, error) {
	didDocResolutionParsed, err := diddoc.ParseDocumentResolution([]byte(d.Content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse DID document resolution content: %w", err)
	}

	// look for assertion method
	verificationMethods := didDocResolutionParsed.DIDDocument.VerificationMethods(diddoc.AssertionMethod)

	if len(verificationMethods[diddoc.AssertionMethod]) > 0 {
		vm := verificationMethods[diddoc.AssertionMethod][0].VerificationMethod

		vmJWK := vm.JSONWebKey()
		if vmJWK != nil {
			return &VerificationMethod{ID: vm.ID, Type: vm.Type, Key: models.VerificationKey{JSONWebKey: vmJWK}}, nil
		}

		return &VerificationMethod{ID: vm.ID, Type: vm.Type, Key: models.VerificationKey{Raw: vm.Value}}, nil
	}

	return nil, fmt.Errorf("DID provided has no assertion method to use as a default signing key")
}
