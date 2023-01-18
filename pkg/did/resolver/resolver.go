/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resolver contains functions for resolving DIDs.
package resolver

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go-ext/component/vdr/longform"
	didDoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/httpbinding"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/key"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/web"

	"github.com/trustbloc/wallet-sdk/pkg/common"
	diderrors "github.com/trustbloc/wallet-sdk/pkg/did"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// DIDResolver is used for resolving DID using supported DID methods.
type DIDResolver struct {
	vdr *vdr.Registry
}

// NewDIDResolver new DID Resolver instance.
//
// Parameter resolverServerURI (optional): if present, provides the URI for a DID resolution server.
func NewDIDResolver(resolverServerURI string) (*DIDResolver, error) {
	ion, err := longform.New()
	if err != nil {
		return nil, walleterror.NewExecutionError(
			diderrors.Module,
			diderrors.ResolverInitializationCode,
			diderrors.ResolverInitializationFailed,
			fmt.Errorf("initializing did:ion longform resolver: %w", err))
	}

	opts := []vdr.Option{
		vdr.WithVDR(key.New()),
		vdr.WithVDR(web.New()),
		vdr.WithVDR(ion),
	}

	if resolverServerURI != "" {
		httpVDR, err := httpbinding.New(resolverServerURI, httpbinding.WithHTTPClient(common.DefaultHTTPClient()),
			httpbinding.WithAccept(func(method string) bool {
				// For now, let the resolver server act as a fallback for all DID methods the sdk does not recognize.
				return true
			}))
		if err != nil {
			return nil,
				walleterror.NewExecutionError(
					diderrors.Module,
					diderrors.ResolverInitializationCode,
					diderrors.ResolverInitializationFailed,
					fmt.Errorf("failed to initialize client for DID resolution server: %w", err))
		}

		opts = append(opts, vdr.WithVDR(httpVDR))
	}

	return &DIDResolver{
		vdr: vdr.New(opts...),
	}, nil
}

// Resolve resolves a DID.
func (d *DIDResolver) Resolve(did string) (*didDoc.DocResolution, error) {
	res, err := d.vdr.Resolve(did)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			diderrors.Module,
			diderrors.ResolutionFailedCode,
			diderrors.ResolutionFailedError,
			fmt.Errorf("resolve %s : %w", did, err),
		)
	}

	return res, nil
}
