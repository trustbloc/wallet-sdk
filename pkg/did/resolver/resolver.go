/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resolver contains functions for resolving DIDs.
package resolver

import (
	"fmt"

	"github.com/trustbloc/wallet-sdk/pkg/api"

	"github.com/hyperledger/aries-framework-go-ext/component/vdr/jwk"
	"github.com/hyperledger/aries-framework-go-ext/component/vdr/longform"
	didDoc "github.com/hyperledger/aries-framework-go/component/models/did"
	"github.com/hyperledger/aries-framework-go/component/vdr"
	"github.com/hyperledger/aries-framework-go/component/vdr/httpbinding"
	"github.com/hyperledger/aries-framework-go/component/vdr/key"
	"github.com/hyperledger/aries-framework-go/component/vdr/web"

	diderrors "github.com/trustbloc/wallet-sdk/pkg/did"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// DIDResolver is used for resolving DID using supported DID methods.
type DIDResolver struct {
	vdr *vdr.Registry
}

// NewDIDResolver returns a new DID Resolver.
func NewDIDResolver(opts ...Opt) (*DIDResolver, error) {
	ion, err := longform.New()
	if err != nil {
		return nil, walleterror.NewExecutionError(
			diderrors.Module,
			diderrors.ResolverInitializationCode,
			diderrors.ResolverInitializationFailed,
			fmt.Errorf("initializing did:ion longform resolver: %w", err))
	}

	vdrOpts := []vdr.Option{
		vdr.WithVDR(key.New()),
		vdr.WithVDR(web.New()),
		vdr.WithVDR(jwk.New()),
		vdr.WithVDR(ion),
	}

	mergedOpts := mergeOpts(opts)

	if mergedOpts.resolverServerURI != "" {
		acceptOpt := httpbinding.WithAccept(func(method string) bool {
			// For now, let the resolver server act as a fallback for all DID methods the SDK does not recognize.
			return true
		})

		var httpTimeoutOpt httpbinding.Option

		if mergedOpts.httpTimeout != nil {
			httpTimeoutOpt = httpbinding.WithTimeout(*mergedOpts.httpTimeout)
		} else {
			httpTimeoutOpt = httpbinding.WithTimeout(api.DefaultHTTPTimeout)
		}

		httpBindingOpts := []httpbinding.Option{acceptOpt, httpTimeoutOpt}

		httpVDR, err := httpbinding.New(mergedOpts.resolverServerURI, httpBindingOpts...)
		if err != nil {
			return nil,
				walleterror.NewExecutionError(
					diderrors.Module,
					diderrors.ResolverInitializationCode,
					diderrors.ResolverInitializationFailed,
					fmt.Errorf("failed to initialize client for DID resolution server: %w", err))
		}

		vdrOpts = append(vdrOpts, vdr.WithVDR(httpVDR))
	}

	return &DIDResolver{
		vdr: vdr.New(vdrOpts...),
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
