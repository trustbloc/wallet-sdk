/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resolver contains functions for resolving DIDs.
package resolver

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go-ext/component/vdr/longform"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/key"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/web"

	didDoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/vdr"
)

// DIDResolver is used for resolving DID using supported DID methods.
type DIDResolver struct {
	vdr *vdr.Registry
}

// NewDIDResolver new DID Resolver instance.
func NewDIDResolver() (*DIDResolver, error) {
	ion, err := longform.New()
	if err != nil {
		return nil, fmt.Errorf("initializing did:ion longform resolver: %w", err)
	}

	return &DIDResolver{
		vdr: vdr.New(
			vdr.WithVDR(key.New()),
			vdr.WithVDR(web.New()),
			vdr.WithVDR(ion),
		),
	}, nil
}

// Resolve resolves a DID.
func (d *DIDResolver) Resolve(did string) (*didDoc.DocResolution, error) {
	res, err := d.vdr.Resolve(did)
	if err != nil {
		return nil, fmt.Errorf("resolve %s : %w", did, err)
	}

	return res, nil
}
