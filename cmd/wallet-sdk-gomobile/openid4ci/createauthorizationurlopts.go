/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// CreateAuthorizationURLOpts contains all optional arguments that can be passed into the
// createAuthorizationURL method.
type CreateAuthorizationURLOpts struct {
	scopes      *api.StringArray
	issuerState *string
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

// SetIssuerState is an option for the createAuthorizationURL method that specifies an issuer state to be included in
// the authorization URL.
//
// For an issuer-instantiated flow, this option should not be required in most cases. Typically, if an issuer requires
// this parameter, it will be included in the original credential offer, and in such cases the createAuthorizationURL
// method will automatically include it in the authorization URL without requiring this option to be used.
// Since the spec leaves open the possibility that the issuer_state parameter can come from some other place,
// this option exists to allow for compatibility with such scenarios. However, the spec also states that if the
// credential offer specifies an issuer state, then it MUST be used in the authorization URL. Thus, in order to prevent
// potential confusion, if the credential offer already has an issuer state value, but a caller still uses this option,
// then an error will be returned by the CreateAuthorizationURL method. If needed, a caller can check the IssuerState
// field in the AuthorizationCodeGrantParams object.
//
// For a wallet-instantiated flow, an issuer state may be required by some issuers. There is no credential offer
// in a wallet-instantiated flow, so if an issuer state is required then it must always be explicitly provided using
// this option.
func (c *CreateAuthorizationURLOpts) SetIssuerState(issuerState string) *CreateAuthorizationURLOpts {
	c.issuerState = &issuerState

	return c
}
