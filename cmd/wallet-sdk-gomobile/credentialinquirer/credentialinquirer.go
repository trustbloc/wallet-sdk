/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialinquirer contains functionality for doing credential queries.
package credentialinquirer

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/presentationexchange"
)

// Inquirer can perform credential queries.
type Inquirer struct {
	credentialReader     api.CredentialReader
	presentationExchange presentationexchange.Exchange
}

// NewInquirer returns a new Inquirer.
func NewInquirer(credentialReader api.CredentialReader, presentationExchange presentationexchange.Exchange) *Inquirer {
	return &Inquirer{
		credentialReader:     credentialReader,
		presentationExchange: presentationExchange,
	}
}

// Credentials represents the different ways that credentials can be passed in to the Query method.
// At most one out of VCs and CredentialReader should be used for a given call to Query. If both are specified,
// then VCs will take precedence. If neither are specified, then the CredentialReader from the constructor (NewInquirer)
// will be used instead.
type Credentials struct {
	// VCs is a JSON array of Verifiable Credentials. If specified, this takes precedence over the CredentialReader
	// used in the constructor (NewInquirer).
	VCs []byte
	// CredentialReader allows for access to a VC storage mechanism. If specified, this takes precedence over the
	// CredentialReader used in the constructor (NewInquirer).
	CredentialReader api.CredentialReader
}

// Query returns credentials that match the given query.
// Query is the filter to be applied to the credentials (TODO: define format).
// See the Credentials struct for more information on how to use it with this method.
// Credentials are returned as a JSON array.
func (*Inquirer) Query(query string, credentials *Credentials) ([]byte, error) {
	return []byte("Example credentials"), nil
}
