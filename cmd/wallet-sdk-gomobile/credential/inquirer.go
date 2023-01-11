/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credential contains a type that can be used to query for credentials using a presentation definition.
// It also contains a credential storage implementation using in-memory storage only.
package credential

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/walleterror"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/credentialquery"
)

// Inquirer implements querying credentials using presentation definition.
type Inquirer struct {
	documentLoader       ld.DocumentLoader
	goAPICredentialQuery *credentialquery.Instance
}

// CredentialsOpt represents the different ways that credentials can be passed in to the Query method.
// At most one out of VCs and CredentialReader should be used for a given call to Resolve. If both are specified,
// then VCs will take precedence.
type CredentialsOpt struct {
	// VCs is an array of Verifiable CredentialsOpt. If specified, this takes precedence over the CredentialReader
	// used in the constructor (NewResolver).
	VCs *api.VerifiableCredentialsArray
	// CredentialReader allows for access to a VC storage mechanism.
	CredentialReader api.CredentialReader
}

// NewCredentialsOpt creates CredentialsOpt from VerifiableCredentialsArray.
func NewCredentialsOpt(vcArr *api.VerifiableCredentialsArray) *CredentialsOpt {
	return &CredentialsOpt{
		VCs: vcArr,
	}
}

// NewCredentialsOptFromReader creates CredentialsOpt from CredentialReader.
func NewCredentialsOptFromReader(credentialReader api.CredentialReader) *CredentialsOpt {
	return &CredentialsOpt{
		CredentialReader: credentialReader,
	}
}

// VerifiablePresentation typed wrapper around go implementation of verifiable presentation.
type VerifiablePresentation struct {
	wrapped *verifiable.Presentation
}

// wrapVerifiablePresentation wraps go implementation of verifiable presentation into gomobile compatible struct.
func wrapVerifiablePresentation(vp *verifiable.Presentation) *VerifiablePresentation {
	return &VerifiablePresentation{wrapped: vp}
}

// Content return marshaled representation of verifiable presentation.
func (vp *VerifiablePresentation) Content() ([]byte, error) {
	return vp.wrapped.MarshalJSON()
}

// Credentials returns marshaled representation of credentials from this verifiable presentation.
func (vp *VerifiablePresentation) Credentials() (*api.VerifiableCredentialsArray, error) {
	result := api.NewVerifiableCredentialsArray()

	credentialsRaw, err := vp.wrapped.MarshalledCredentials()
	if err != nil {
		return nil, err
	}

	for _, cred := range credentialsRaw {
		result.Add(api.NewVerifiableCredential(unQuote(string(cred))))
	}

	return result, nil
}

// NewInquirer returns a new Inquirer.
func NewInquirer(documentLoader api.LDDocumentLoader) *Inquirer {
	wrappedLoader := &wrapper.DocumentLoaderWrapper{
		DocumentLoader: documentLoader,
	}

	return &Inquirer{
		documentLoader:       wrappedLoader,
		goAPICredentialQuery: credentialquery.NewInstance(wrappedLoader),
	}
}

// Query returns credentials that match PresentationDefinition.
func (c *Inquirer) Query(query []byte, contents *CredentialsOpt) (*VerifiablePresentation, error) {
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
		credentials, err = c.parseVC(contents.VCs)
		if err != nil {
			return nil, err
		}
	}

	presentation, err := c.goAPICredentialQuery.Query(pdQuery,
		credentialquery.WithCredentialsArray(credentials),
		credentialquery.WithCredentialReader(&wrapper.CredentialReaderWrapper{
			CredentialReader: contents.CredentialReader,
			DocumentLoader:   c.documentLoader,
		}),
	)
	if err != nil {
		return nil, walleterror.ToMobileError(err)
	}

	return wrapVerifiablePresentation(presentation), err
}

func (c *Inquirer) parseVC(rawCreds *api.VerifiableCredentialsArray) ([]*verifiable.Credential, error) {
	var credentials []*verifiable.Credential

	for i := 0; i < rawCreds.Length(); i++ {
		content := rawCreds.AtIndex(i).Content

		cred, credErr := verifiable.ParseCredential([]byte(content), verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(c.documentLoader))
		if credErr != nil {
			return nil, fmt.Errorf("verifiable credential parse failed: %w", credErr)
		}

		credentials = append(credentials, cred)
	}

	return credentials, nil
}

func unQuote(s string) string {
	if len(s) <= 1 {
		return s
	}

	if s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}

	return s
}
