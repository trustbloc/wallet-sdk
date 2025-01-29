/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

type createAuthorizationURLOpts struct {
	scopes                             []string
	issuerState                        *string
	useOAuthDiscoverableClientIDScheme bool
	context                            []string
}

// CreateAuthorizationURLOpt is an option for the CreateAuthorizationURL method.
type CreateAuthorizationURLOpt func(opts *createAuthorizationURLOpts)

// WithScopes is an option for the createAuthorizationURL method that allows scopes to be passed in.
func WithScopes(scopes []string) CreateAuthorizationURLOpt {
	return func(opts *createAuthorizationURLOpts) {
		opts.scopes = scopes
	}
}

// WithIssuerState is an option for the CreateAuthorizationURL method that specifies an issuer state to be included in
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
func WithIssuerState(issuerState string) CreateAuthorizationURLOpt {
	return func(opts *createAuthorizationURLOpts) {
		opts.issuerState = &issuerState
	}
}

// WithOAuthDiscoverableClientIDScheme is an option for the CreateAuthorizationURL method that will cause the
// OAuth Discoverable Client ID scheme to be specified in the authorization URL.
// See https://mattrglobal.github.io/draft-looker-oauth-client-id-scheme/draft-looker-oauth-client-id-scheme.html
// for more information on the requirements of this scheme. Note that successful usage of this scheme requires the
// issuer to support it and for the client to meet the scheme's pre-requisites. Be sure to also set your client ID
// appropriately as required by the scheme.
func WithOAuthDiscoverableClientIDScheme() CreateAuthorizationURLOpt {
	return func(opts *createAuthorizationURLOpts) {
		opts.useOAuthDiscoverableClientIDScheme = true
	}
}

// WithCredentialContext is an option for the CreateAuthorizationURL method that specifies an credential context
// See https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#appendix-A.1.2.3
// for more information about authorization request.
func WithCredentialContext(context []string) CreateAuthorizationURLOpt {
	return func(opts *createAuthorizationURLOpts) {
		opts.context = context
	}
}

func processCreateAuthorizationURLOpts(opts []CreateAuthorizationURLOpt) *createAuthorizationURLOpts {
	processedOpts := &createAuthorizationURLOpts{}

	for _, opt := range opts {
		if opt != nil {
			opt(processedOpts)
		}
	}

	return processedOpts
}
