/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialsigner contains functionality for doing credential signing operations.
package credentialsigner

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// Signer is used to do credential signing operations.
type Signer struct {
	keyHandleReader  api.KeyHandleReader
	didResolver      api.DIDResolver
	credentialReader api.CredentialReader // Optional
	crypto           api.Crypto
}

// NewSigner returns a new Signer instance.
func NewSigner(keyHandleReader api.KeyHandleReader, didResolver api.DIDResolver, credentialReader api.CredentialReader,
	crypto api.Crypto,
) *Signer {
	return &Signer{
		keyHandleReader:  keyHandleReader,
		didResolver:      didResolver,
		credentialReader: credentialReader,
		crypto:           crypto,
	}
}

// Issue issues a credential.
// credential must be a valid marshalled VC in JSON form.
// TODO: Format of proofOpts to be determined.
// Returns a credential.
func (s *Signer) Issue(credential []byte, id, proofOpts string) ([]byte, error) {
	return []byte("Sample data"), nil
}

// Derive does something (TODO: Implement).
// credential must be a valid marshalled VC in JSON form.
// Returns a credential.
func (s *Signer) Derive(credential []byte, id string, jsonFrame []byte, nonce string) ([]byte, error) {
	return []byte("Sample credential"), nil
}

// Prove does something (TODO: Implement).
// credentials must be valid marshalled VCs in a JSON array.
// ids must be a JSON array.
// presentation must be a valid marshalled presentation in JSON form.
// TODO: Format of proofOpts to be determined.
// Returns a presentation.
func (s *Signer) Prove(credentials, ids, presentation []byte, proofOpts string) ([]byte, error) {
	return []byte("Sample presentation"), nil
}
