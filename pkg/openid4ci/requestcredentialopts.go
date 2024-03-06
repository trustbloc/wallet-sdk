/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "github.com/trustbloc/wallet-sdk/pkg/api"

type requestCredentialWithPreAuthOpts struct {
	pin                 string
	attestationVPSigner api.JWTSigner
	attestationVC       string
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

// WithAttestationVC is an option for the RequestCredentialWithPreAuth method that allows you to specify
// attestation VC, which may be required by the issuer.
func WithAttestationVC(attestationVPSigner api.JWTSigner, attestationVC string) RequestCredentialWithPreAuthOpt {
	return func(opts *requestCredentialWithPreAuthOpts) {
		opts.attestationVC = attestationVC
		opts.attestationVPSigner = attestationVPSigner
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
