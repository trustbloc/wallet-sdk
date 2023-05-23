/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package key contains a did:key creator implementation.
package key

import (
	"github.com/hyperledger/aries-framework-go/component/models/did"
	"github.com/hyperledger/aries-framework-go/component/vdr/key"
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
func (d *Creator) Create(vm *did.VerificationMethod) (*did.DocResolution, error) {
	didDocArgument := &did.Doc{VerificationMethod: []did.VerificationMethod{*vm}}

	return d.vdr.Create(didDocArgument)
}
