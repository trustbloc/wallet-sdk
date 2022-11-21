/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package wallethandle contains functionality for doing wallet import and export operations.
package wallethandle

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// A Handle exposes wallet import and export functionality.
type Handle struct {
	credentialReader api.CredentialReader
	credentialWriter api.CredentialWriter
	crypto           api.Crypto
}

// NewHandle returns a new Handle instance.
func NewHandle(credentialReader api.CredentialReader, credentialWriter api.CredentialWriter,
	crypto api.Crypto,
) *Handle {
	return &Handle{
		credentialReader: credentialReader,
		credentialWriter: credentialWriter,
		crypto:           crypto,
	}
}

// Import imports the given credentials into the wallet.
// The credentials are expected to be a JSON array.
func (w *Handle) Import(credentials []byte) error {
	return nil
}

// Export returns all credentials stored in this wallet.
func (w *Handle) Export() ([]byte, error) {
	return []byte("Sample data"), nil
}
