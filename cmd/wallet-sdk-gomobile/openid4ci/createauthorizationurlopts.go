/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// CreateAuthorizationURLOpts contains all optional arguments that can be passed into the
// createAuthorizationURL method.
type CreateAuthorizationURLOpts struct {
	scopes *api.StringArray
}

// NewCreateAuthorizationURLOpts returns a new CreateAuthorizationURLOpts object.
func NewCreateAuthorizationURLOpts() *CreateAuthorizationURLOpts {
	return &CreateAuthorizationURLOpts{}
}

// SetScopes sets scopes to use in the URL created by the createAuthorizationURL method. If the authorization URL
// requires scopes to be set, then this option must be used.
func (c *CreateAuthorizationURLOpts) SetScopes(scopes *api.StringArray) *CreateAuthorizationURLOpts {
	c.scopes = scopes

	return c
}
