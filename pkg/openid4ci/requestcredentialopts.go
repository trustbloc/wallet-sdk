/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

type requestCredentialWithPreAuthOpts struct {
	pin string
}

// RequestCredentialWithPreAuthOpt is an option for the RequestCredentialWithPreAuth method.
type RequestCredentialWithPreAuthOpt func(opts *requestCredentialWithPreAuthOpts)

// WithPIN is an option for the RequestCredentialWithPreAuth method that allows you to specify a PIN, which may be
// required by the issuer. Check the issuer capabilities object first to determine this.
func WithPIN(pin string) RequestCredentialWithPreAuthOpt {
	return func(opts *requestCredentialWithPreAuthOpts) {
		opts.pin = pin
	}
}

func processRequestCredentialWithPreAuthOpts(opts []RequestCredentialWithPreAuthOpt) *requestCredentialWithPreAuthOpts {
	processedOpts := &requestCredentialWithPreAuthOpts{}

	for _, opt := range opts {
		if opt != nil {
			opt(processedOpts)
		}
	}

	return processedOpts
}
