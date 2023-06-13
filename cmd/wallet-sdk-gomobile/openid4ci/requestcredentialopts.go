/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

// RequestCredentialWithPreAuthOpts contains all optional arguments that can be passed into the
// RequestCredentialWithPreAuth method.
type RequestCredentialWithPreAuthOpts struct {
	pin string
}

// NewRequestCredentialWithPreAuthOpts returns a new RequestCredentialWithPreAuthOpts object.
func NewRequestCredentialWithPreAuthOpts() *RequestCredentialWithPreAuthOpts {
	return &RequestCredentialWithPreAuthOpts{}
}

// SetPIN is an option for the RequestCredentialWithPreAuth method that allows you to specify a PIN, which may be
// required by the issuer. Check the issuer capabilities object first to determine this.
func (r *RequestCredentialWithPreAuthOpts) SetPIN(pin string) *RequestCredentialWithPreAuthOpts {
	r.pin = pin

	return r
}

// RequestCredentialWithAuthOpts contains all optional arguments that can be passed into the
// RequestCredentialWithAuth method.
type RequestCredentialWithAuthOpts struct{}
