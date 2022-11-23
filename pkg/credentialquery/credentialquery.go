/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialquery allows querying credentials using presentation definition.
package credentialquery

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// Instance implements querying credentials using presentation definition.
type Instance struct {
	documentLoader ld.DocumentLoader
}

type queryOpts struct {
	// credentials is an array of Verifiable Credentials.
	credentials []*verifiable.Credential
	// CredentialReader allows for access to a VC storage mechanism.
	credentialReader api.CredentialReader
}

// QueryOpt is the query credential option.
type QueryOpt func(opts *queryOpts)

// WithCredentialsArray sets array of Verifiable Credentials. If specified,
// this takes precedence over the CredentialReader option.
func WithCredentialsArray(vcs []*verifiable.Credential) QueryOpt {
	return func(opts *queryOpts) {
		opts.credentials = vcs
	}
}

// WithCredentialReader sets credential reader that will be used to fetch credential.
func WithCredentialReader(credentialReader api.CredentialReader) QueryOpt {
	return func(opts *queryOpts) {
		opts.credentialReader = credentialReader
	}
}

// NewInstance returns new Instance.
func NewInstance(documentLoader ld.DocumentLoader) *Instance {
	return &Instance{documentLoader: documentLoader}
}

// Query returns credentials that match PresentationDefinition.
func (c *Instance) Query(
	query *presexch.PresentationDefinition,
	opts ...QueryOpt,
) (*verifiable.Presentation, error) {
	qOpts := &queryOpts{}
	for _, opt := range opts {
		opt(qOpts)
	}

	credentials := qOpts.credentials
	if len(credentials) == 0 {
		if qOpts.credentialReader == nil {
			return nil, fmt.Errorf("credentials array or credential reader option should be set")
		}

		var err error

		credentials, err = qOpts.credentialReader.GetAll()
		if err != nil {
			return nil, fmt.Errorf("credential reader failed: %w", err)
		}
	}

	return query.CreateVP(credentials, c.documentLoader, verifiable.WithDisabledProofCheck(),
		verifiable.WithJSONLDDocumentLoader(c.documentLoader))
}
