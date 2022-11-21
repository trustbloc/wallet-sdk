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

type goAPICredentialQuery interface {
	Query(query *presexch.PresentationDefinition, contents []*verifiable.Credential) (*verifiable.Presentation, error)
}

// Query implements querying credentials using presentation definition.
type Query struct {
	documentLoader       ld.DocumentLoader
	goAPICredentialQuery goAPICredentialQuery
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
func (c *Query) Query(query, contents []byte) ([]byte, error) {
	pdQuery := &presexch.PresentationDefinition{}

	err := json.Unmarshal(query, pdQuery)
	if err != nil {
		return nil, fmt.Errorf("unmarshal of presentation definition failed: %w", err)
	}

	err = pdQuery.ValidateSchema()
	if err != nil {
		return nil, fmt.Errorf("validation of presentation definition failed: %w", err)
	}

	var credsJWTsStrs []string

	err = json.Unmarshal(contents, &credsJWTsStrs)
	if err != nil || len(credsJWTsStrs) == 0 {
		return nil, fmt.Errorf("unmarshal of credentials array failed, "+
			"should be json array of jwt strings: %w", err)
	}

	var credentials []*verifiable.Credential

	for _, credContent := range credsJWTsStrs {
		cred, credErr := verifiable.ParseCredential([]byte(credContent), verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(c.documentLoader))
		if credErr != nil {
			return nil, fmt.Errorf("verifiable credential parse failed: %w", credErr)
		}

		credentials = append(credentials, cred)
	}

	presentation, err := c.goAPICredentialQuery.Query(pdQuery, credentials)
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
