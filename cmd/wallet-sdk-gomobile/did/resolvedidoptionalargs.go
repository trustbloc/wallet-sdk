/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

// ResolverOpts contains all optional arguments that can be passed into the ResolveDID method.
type ResolverOpts struct {
	resolverServerURI string
}

// NewResolverOpts returns a new ResolverOpts object.
func NewResolverOpts() *ResolverOpts {
	return &ResolverOpts{}
}

// SetResolverServerURI sets a resolver server to use when resolving certain types of DIDs.
func (c *ResolverOpts) SetResolverServerURI(resolverServerURI string) {
	c.resolverServerURI = resolverServerURI
}
