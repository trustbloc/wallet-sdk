/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialschema contains functionality for doing credential resolution with credential manifests.
package credentialschema

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// Resolver can resolve credentials.
type Resolver struct {
	credentialReader api.CredentialReader
}

// NewResolver creates a new Resolver.
// If a CredentialReader is specified here, then it will be used as the source for VCs in the Resolve method unless a
// CredentialReader exists in the ResolveOpts.
func NewResolver(credentialReader api.CredentialReader) *Resolver {
	return &Resolver{credentialReader: credentialReader}
}

// Schema represents the schema options for the Resolve method.
// Only one out of Schema and CredentialsSupported should be used for a given call to Resolve. If both are specified,
// then Schema will take precedence.
type Schema struct {
	// CredentialManifest is the Credential Manifest to use for resolving descriptors.
	CredentialManifest []byte
	// CredentialsSupported is not yet defined (TODO: define).
	CredentialsSupported []byte
}

// Credentials represents the different ways that credentials can be passed in to the Resolve method.
// At most one out of VCs and CredentialReader should be used for a given call to Resolve. If both are specified,
// then VCs will take precedence. If neither are specified, then the CredentialReader from the constructor (NewResolver)
// will be used instead.
// Optionally, CredentialID may be specified if only one VC out of VCs or the CredentialReader should be specified.
type Credentials struct {
	// VCs is a JSON array of Verifiable Credentials. If specified, this takes precedence over the CredentialReader
	// used in the constructor (NewResolver).
	VCs []byte
	// CredentialReader allows for access to a VC storage mechanism. If specified, this takes precedence over the
	// CredentialReader used in the constructor (NewResolver).
	CredentialReader api.CredentialReader
	// CredentialID specifies that only a single credential from VCs or the CredentialReader should be examined.
	// If not specified, then all credentials from VCs or CredentialReader will be examined.
	CredentialID string
}

// Resolve resolves the given credentials and returns resolved descriptors. See the Schema and Credentials structs
// for more information.
// Returns a JSON array of resolved descriptors.
func (c *Resolver) Resolve(schema *Schema, credentials *Credentials) ([]byte, error) {
	return []byte("Content"), nil
}
