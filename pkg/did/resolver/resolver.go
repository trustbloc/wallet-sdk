/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package resolver contains functions for resolving DIDs.
package resolver

import (
	"fmt"

	didDoc "github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/did-go/method/httpbinding"
	"github.com/trustbloc/did-go/method/jwk"
	"github.com/trustbloc/did-go/method/key"
	"github.com/trustbloc/did-go/method/web"
	"github.com/trustbloc/did-go/vdr"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	longform "github.com/trustbloc/sidetree-go/pkg/vdr/sidetreelongform"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	diderrors "github.com/trustbloc/wallet-sdk/pkg/did"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// DIDResolver is used for resolving DID using supported DID methods.
type DIDResolver struct {
	vdr        *vdr.Registry
	httpClient httpClient
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
		vdr:        vdr.New(vdrOpts...),
		httpClient: mergedOpts.httpClient,
	}, nil
}

// Resolve resolves a DID.
func (d *DIDResolver) Resolve(did string) (*didDoc.DocResolution, error) {
	res, err := d.vdr.Resolve(did, vdrapi.WithOption(web.HTTPClientOpt, d.httpClient))
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
