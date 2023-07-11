/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

type createAuthorizationURLOpts struct {
	scopes []string
}

// CreateAuthorizationURLOpt is an option for the createAuthorizationURL method.
type CreateAuthorizationURLOpt func(opts *createAuthorizationURLOpts)

// WithScopes is an option for the createAuthorizationURL method that allows scopes to be passed in.
func WithScopes(scopes []string) CreateAuthorizationURLOpt {
	return func(opts *createAuthorizationURLOpts) {
		opts.scopes = scopes
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
