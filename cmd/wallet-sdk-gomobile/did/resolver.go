/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package did contains functionality related to DIDs.
package did

import (
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
)

// Resolver supports DID resolution.
type Resolver struct {
	resolver *resolver.DIDResolver
}

// NewResolver returns a new Resolver.
func NewResolver(opts *ResolverOpts) (*Resolver, error) {
	if opts == nil {
		opts = NewResolverOpts()
	}

	var goAPIResolverOpts []resolver.Opt

	if opts.resolverServerURI != "" {
		resolverServerURIOpt := resolver.WithResolverServerURI(opts.resolverServerURI)

		goAPIResolverOpts = append(goAPIResolverOpts, resolverServerURIOpt)
	}

	if opts.httpTimeout != nil {
		httpTimeoutOpt := resolver.WithHTTPTimeout(*opts.httpTimeout)

		goAPIResolverOpts = append(goAPIResolverOpts, httpTimeoutOpt)
	}

	httpClient := wrapper.NewHTTPClient(opts.httpTimeout, api.Headers{}, opts.disableHTTPClientTLSVerification)

	goAPIResolverOpts = append(goAPIResolverOpts, resolver.WithHTTPClient(httpClient))

	didResolver, err := resolver.NewDIDResolver(goAPIResolverOpts...)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &Resolver{resolver: didResolver}, nil
}

// Resolve resolves a DID.
func (d *Resolver) Resolve(did string) ([]byte, error) {
	didDocResolution, err := d.resolver.Resolve(did)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return didDocResolution.JSONBytes()
}
