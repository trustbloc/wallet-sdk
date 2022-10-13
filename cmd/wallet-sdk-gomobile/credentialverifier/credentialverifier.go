/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialverifier contains functionality for doing credential verification.
package credentialverifier

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// Verifier is used to verify credentials.
type Verifier struct {
	keyHandleReader  api.KeyHandleReader
	didResolver      api.DIDResolver
	credentialReader api.CredentialReader
	crypto           api.Crypto
}

// NewVerifier returns a new Verifier.
func NewVerifier(keyHandleReader api.KeyHandleReader, didResolver api.DIDResolver,
	credentialReader api.CredentialReader, crypto api.Crypto,
) *Verifier {
	return &Verifier{
		keyHandleReader:  keyHandleReader,
		didResolver:      didResolver,
		credentialReader: credentialReader,
		crypto:           crypto,
	}
}

// VerifyOpts represents the various options for the Verify method.
// Only of these three should be used for a given call to Verify. If multiple options are used, then one of them will
// take precedence over the other per the following order: CredentialID, RawCredential, RawPresentation.
type VerifyOpts struct {
	// CredentialID is the ID of the credential to be verified.
	// A credential with the given ID must be found in this Verifier's credentialReader.
	CredentialID string
	// RawCredential is the raw credential to be verified.
	RawCredential []byte
	// RawPresentation is the raw presentation to be verified.
	RawPresentation []byte
}

// Verify verifies the given credential or presentation. See the VerifyOpts struct for more information.
func (*Verifier) Verify(verifyOpts VerifyOpts) error {
	return nil
}
