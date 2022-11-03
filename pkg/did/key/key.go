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

// Creator is used for creating did:key DID Documents.
type Creator struct {
	vdr *key.VDR
}

// NewCreator returns a new did:key document Creator.
func NewCreator() *Creator {
	return &Creator{vdr: key.New()}
}

// Create creates a new did:key document using the given rawKey and verificationMethodType.
func (d *Creator) Create(rawKey []byte, verificationMethodType string) (*did.DocResolution, error) {
	verificationMethod := did.VerificationMethod{Value: rawKey, Type: verificationMethodType}

	didDocArgument := &did.Doc{VerificationMethod: []did.VerificationMethod{verificationMethod}}

	return d.vdr.Create(didDocArgument)
}
