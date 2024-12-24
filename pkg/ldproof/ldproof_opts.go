/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package ldproof

import (
	diddoc "github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/vc-go/presexch"
)

type options struct {
	ldpType            *presexch.LdpType
	verificationMethod *diddoc.VerificationMethod
	did                string
	challenge          string
	domain             string
}

// Opt is an option for adding linked data proof.
type Opt func(opts *options)

// WithLdpType sets the supported JSON-LD proof type.
func WithLdpType(ldpType *presexch.LdpType) Opt {
	return func(opts *options) {
		opts.ldpType = ldpType
	}
}

// WithVerificationMethod sets the verification method to get the associated signing key.
func WithVerificationMethod(vm *diddoc.VerificationMethod) Opt {
	return func(opts *options) {
		opts.verificationMethod = vm
	}
}

// WithDID sets did.
func WithDID(did string) Opt {
	return func(opts *options) {
		opts.did = did
	}
}

// WithChallenge sets challenge.
func WithChallenge(challenge string) Opt {
	return func(opts *options) {
		opts.challenge = challenge
	}
}

// WithDomain sets domain.
func WithDomain(domain string) Opt {
	return func(opts *options) {
		opts.domain = domain
	}
}
