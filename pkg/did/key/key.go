/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package key contains a did:key creator implementation.
package key

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/key"
)

// DIDKeyCreator is used for creating did:key DID Documents.
type DIDKeyCreator struct {
	vdr *key.VDR
}

// NewDIDKeyCreator returns a new DIDKeyCreator.
func NewDIDKeyCreator() *DIDKeyCreator {
	return &DIDKeyCreator{vdr: key.New()}
}

// Create creates a new DID document using the given public key.
func (d *DIDKeyCreator) Create(rawKey []byte, keyType string) (*did.DocResolution, error) {
	verificationMethod := did.VerificationMethod{Value: rawKey, Type: keyType}

	didDocArgument := &did.Doc{VerificationMethod: []did.VerificationMethod{verificationMethod}}

	return d.vdr.Create(didDocArgument)
}
