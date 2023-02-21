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

// NewCredentialsOpt creates CredentialsOpt from VCs.
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

	vp.wrapped.Credentials()

	credentialsRaw := vp.wrapped.Credentials()

	for i := range credentialsRaw {
		cred, ok := credentialsRaw[i].(*verifiable.Credential)
		if !ok {
			return nil, fmt.Errorf("credential at index %d could not be asserted as a *verifiable.Credential", i)
		}

		result.Add(api.NewVerifiableCredential(cred))
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
	pdQuery, credentials, err := unwrapInputs(query, contents)
	if err != nil {
		return nil, err
	}

	presentation, err := c.goAPICredentialQuery.Query(pdQuery,
		credentialquery.WithCredentialsArray(credentials),
		credentialquery.WithCredentialReader(&wrapper.CredentialReaderWrapper{
			CredentialReader: contents.CredentialReader,
		}),
	)
	if err != nil {
		return nil, walleterror.ToMobileError(err)
	}

	return wrapVerifiablePresentation(presentation), err
}

// GetSubmissionRequirements returns information about VCs matching requirements.
func (c *Inquirer) GetSubmissionRequirements(query []byte, contents *CredentialsOpt,
) (*SubmissionRequirementArray, error) {
	pdQuery, credentials, err := unwrapInputs(query, contents)
	if err != nil {
		return nil, err
	}

	requirements, err := c.goAPICredentialQuery.GetSubmissionRequirements(pdQuery,
		credentialquery.WithCredentialsArray(credentials),
		credentialquery.WithCredentialReader(&wrapper.CredentialReaderWrapper{
			CredentialReader: contents.CredentialReader,
		}),
	)
	if err != nil {
		return nil, walleterror.ToMobileError(err)
	}

	return &SubmissionRequirementArray{wrapped: requirements}, nil
}

func unwrapInputs(query []byte, contents *CredentialsOpt,
) (*presexch.PresentationDefinition, []*verifiable.Credential, error) {
	pdQuery := &presexch.PresentationDefinition{}

	err := json.Unmarshal(query, pdQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("unmarshal of presentation definition failed: %w", err)
	}

	if pdQuery.Format == nil {
		pdQuery.Format = &presexch.Format{
			JwtVP: &presexch.JwtType{
				Alg: []string{"RS256", "EdDSA", "PS256", "ES256K", "ES256", "ES384", "ES521"},
			},
		}
	}

	err = pdQuery.ValidateSchema()
	if err != nil {
		return nil, nil, fmt.Errorf("validation of presentation definition failed: %w", err)
	}

	if contents.CredentialReader == nil && contents.VCs == nil {
		return nil, nil, fmt.Errorf("either credential reader or vc array should be set")
	}

	var credentials []*verifiable.Credential

	if contents.VCs != nil {
		credentials = unwrapVCs(contents.VCs)
	}

	return pdQuery, credentials, nil
}

func unwrapVCs(vcs *api.VerifiableCredentialsArray) []*verifiable.Credential {
	var credentials []*verifiable.Credential

	for i := 0; i < vcs.Length(); i++ {
		credentials = append(credentials, vcs.AtIndex(i).VC)
	}

	return credentials
}
