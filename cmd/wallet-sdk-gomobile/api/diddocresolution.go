/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"fmt"

	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
)

// DIDDocResolution represents a DID document resolution object.
type DIDDocResolution struct {
	// Content is the full marshalled DID doc resolution object.
	Content []byte
}

// NewDIDDocResolution creates a new DIDDocResolution.
func NewDIDDocResolution(content []byte) *DIDDocResolution {
	return &DIDDocResolution{
		Content: content,
	}
}

// ID returns the ID of the DID document contained within this DIDDocResolution.
func (d *DIDDocResolution) ID() (string, error) {
	didDocResolutionParsed, err := diddoc.ParseDocumentResolution(d.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse DID document resolution content: %w", err)
	}

	return didDocResolutionParsed.DIDDocument.ID, nil
}

// AssertionMethodKeyID returns the id of the key that was used for assertion verification.
func (d *DIDDocResolution) AssertionMethodKeyID() (string, error) {
	didDocResolutionParsed, err := diddoc.ParseDocumentResolution(d.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse DID document resolution content: %w", err)
	}

	// look for assertion method
	verificationMethods := didDocResolutionParsed.DIDDocument.VerificationMethods(diddoc.AssertionMethod)

	if len(verificationMethods[diddoc.AssertionMethod]) > 0 {
		vm := verificationMethods[diddoc.AssertionMethod][0].VerificationMethod

		return vm.ID, nil
	}

	return "", fmt.Errorf("DID provided has no assertion method to use as a default signing key")
}
