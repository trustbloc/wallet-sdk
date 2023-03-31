/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

// StatusVerifierOpts contains optional parameters for initializing a credential StatusVerifier.
type StatusVerifierOpts struct{}

// NewStatusVerifierOpts returns a StatusVerifierOpts object.
func NewStatusVerifierOpts() *StatusVerifierOpts {
	return &StatusVerifierOpts{}
}
