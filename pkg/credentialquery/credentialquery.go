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
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
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

// GetSubmissionRequirements returns information about VCs matching requirements.
func (c *Instance) GetSubmissionRequirements(
	query *presexch.PresentationDefinition,
	opts ...QueryOpt,
) ([]*presexch.MatchedSubmissionRequirement, error) {
	qOpts := &queryOpts{}
	for _, opt := range opts {
		opt(qOpts)
	}

	credentials, err := getCredentials(qOpts)
	if err != nil {
		return nil, err
	}

	// TODO: https://github.com/trustbloc/wallet-sdk/issues/165 remove this code after to re enable Schema check.
	for i := range query.InputDescriptors {
		query.InputDescriptors[i].Schema = nil
	}

	results, err := query.MatchSubmissionRequirement(
		credentials,
		c.documentLoader,
	)
	if err != nil {
		return nil,
			walleterror.NewValidationError(
				module,
				FailToGetMatchRequirementsResultsCode,
				FailToGetMatchRequirementsResultsError,
				err)
	}

	return results, nil
}

func getCredentials(qOpts *queryOpts) ([]*verifiable.Credential, error) {
	credentials := qOpts.credentials
	if len(credentials) == 0 {
		if qOpts.credentialReader == nil {
			return nil, walleterror.NewValidationError(
				module,
				CredentialReaderNotSetCode,
				CredentialReaderNotSetError,
				fmt.Errorf("credentials array or credential reader option must be set"))
		}

		var err error

		credentials, err = qOpts.credentialReader.GetAll()
		if err != nil {
			return nil,
				walleterror.NewValidationError(
					module,
					CredentialReaderReadFailedCode,
					CredentialReaderReadFailedError,
					fmt.Errorf("credential reader failed: %w", err))
		}
	}

	return credentials, nil
}
