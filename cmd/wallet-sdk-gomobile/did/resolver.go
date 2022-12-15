/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import (
	// helps gomobile bind api.DIDResolver interface to Resolver implementation in ios-bindings.
	_ "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
)

// Resolver supports DID resolution.
type Resolver struct {
	resolver *resolver.DIDResolver
}

// NewResolver returns a new Resolver.
func NewResolver(resolverServerURI string) (*Resolver, error) {
	didResolver, err := resolver.NewDIDResolver(resolverServerURI)
	if err != nil {
		return nil, err
	}

	return &Resolver{resolver: didResolver}, nil
}

// Resolve resolves a DID.
func (d *Resolver) Resolve(did string) ([]byte, error) {
	didDocResolution, err := d.resolver.Resolve(did)
	if err != nil {
		return nil, err
	}

	return didDocResolution.JSONBytes()
}
