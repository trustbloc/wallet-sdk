/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package didresolver contains functions for resolving DIDs.
package didresolver

import (
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
)

// Resolver supports DID resolution.
type Resolver struct {
	resolver *resolver.DIDResolver
}

// NewDIDResolver returns a new DIDResolver.
func NewDIDResolver() *Resolver {
	return &Resolver{resolver: resolver.NewDIDResolver()}
}

// Resolve resolves a DID.
func (d *Resolver) Resolve(did string) ([]byte, error) {
	didDocResolution, err := d.resolver.Resolve(did)
	if err != nil {
		return nil, err
	}

	return didDocResolution.JSONBytes()
}
