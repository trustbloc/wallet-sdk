/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialquery allows querying credentials using presentation definition.
package credentialquery

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/internal/gomobilewrappers"
	"github.com/trustbloc/wallet-sdk/pkg/credentialquery"
)

// Query implements querying credentials using presentation definition.
type Query struct {
	documentLoader       ld.DocumentLoader
	goAPICredentialQuery *credentialquery.Instance
}

// Credentials represents the different ways that credentials can be passed in to the Query method.
// At most one out of VCs and CredentialReader should be used for a given call to Resolve. If both are specified,
// then VCs will take precedence.
type Credentials struct {
	// VCs is a JSON array of Verifiable Credentials. If specified, this takes precedence over the CredentialReader
	// used in the constructor (NewResolver).
	VCs *api.JSONArray
	// CredentialReader allows for access to a VC storage mechanism.
	CredentialReader api.CredentialReader
}

// NewQuery returns new Query.
func NewQuery(documentLoader api.LDDocumentLoader) *Query {
	wrappedLoader := &gomobilewrappers.DocumentLoaderWrapper{
		DocumentLoader: documentLoader,
	}

	return &Query{
		documentLoader:       wrappedLoader,
		goAPICredentialQuery: credentialquery.NewInstance(wrappedLoader),
	}
}

// Query returns credentials that match PresentationDefinition.
func (c *Query) Query(query []byte, contents *Credentials) ([]byte, error) {
	pdQuery := &presexch.PresentationDefinition{}

	err := json.Unmarshal(query, pdQuery)
	if err != nil {
		return nil, fmt.Errorf("unmarshal of presentation definition failed: %w", err)
	}

	err = pdQuery.ValidateSchema()
	if err != nil {
		return nil, fmt.Errorf("validation of presentation definition failed: %w", err)
	}

	if contents.CredentialReader == nil && contents.VCs == nil {
		return nil, fmt.Errorf("either credential reader or vc array should be set")
	}

	var credentials []*verifiable.Credential

	if contents.VCs != nil {
		credentials, err = c.parseVC(contents.VCs.Data)
		if err != nil {
			return nil, err
		}
	}

	presentation, err := c.goAPICredentialQuery.Query(pdQuery,
		credentialquery.WithCredentialsArray(credentials),
		credentialquery.WithCredentialReader(&gomobilewrappers.CredentialReaderWrapper{
			CredentialReader: contents.CredentialReader,
			DocumentLoader:   c.documentLoader,
		}),
	)
	if err != nil {
		if errors.Is(err, presexch.ErrNoCredentials) {
			return nil, err
		}

		return nil, fmt.Errorf("query is failed: %w", err)
	}

	result, err := presentation.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal presentation: %w", err)
	}

	return result, err
}

func (c *Query) parseVC(data []byte) ([]*verifiable.Credential, error) {
	var credentials []*verifiable.Credential

	var credsJWTsStrs []string

	err := json.Unmarshal(data, &credsJWTsStrs)
	if err != nil || len(credsJWTsStrs) == 0 {
		return nil, fmt.Errorf("unmarshal of credentials array failed, "+
			"should be json array of jwt strings: %w", err)
	}

	for _, credContent := range credsJWTsStrs {
		cred, credErr := verifiable.ParseCredential([]byte(credContent), verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(c.documentLoader))
		if credErr != nil {
			return nil, fmt.Errorf("verifiable credential parse failed: %w", credErr)
		}

		credentials = append(credentials, cred)
	}

	return credentials, nil
}
