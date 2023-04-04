/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credential contains a type that can be used to query for credentials using a presentation definition.
// It also contains a credential storage implementation using in-memory storage only.
package credential

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	afgoverifiable "github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/credentialquery"
)

// Inquirer implements querying credentials using presentation definition.
type Inquirer struct {
	goAPICredentialQuery *credentialquery.Instance
}

// Credentials returns marshaled representation of credentials from this verifiable presentation.
func (vp *VerifiablePresentation) Credentials() (*verifiable.CredentialsArray, error) {
	result := verifiable.NewCredentialsArray()

	vp.wrapped.Credentials()

	credentialsRaw := vp.wrapped.Credentials()

	for i := range credentialsRaw {
		cred, ok := credentialsRaw[i].(*afgoverifiable.Credential)
		if !ok {
			return nil, fmt.Errorf("credential at index %d could not be asserted as a *verifiable.Credential", i)
		}

		result.Add(verifiable.NewCredential(cred))
	}

	return result, nil
}

// NewInquirer returns a new Inquirer.
func NewInquirer(opts *InquirerOpts) *Inquirer {
	if opts == nil {
		opts = &InquirerOpts{}
	}

	var goAPIDocumentLoader ld.DocumentLoader

	if opts.documentLoader != nil {
		goAPIDocumentLoader = &wrapper.DocumentLoaderWrapper{
			DocumentLoader: opts.documentLoader,
		}
	} else {
		goAPIDocumentLoader = ld.NewDefaultDocumentLoader(http.DefaultClient)
	}

	return &Inquirer{
		goAPICredentialQuery: credentialquery.NewInstance(goAPIDocumentLoader),
	}
}

// Query returns credentials that match PresentationDefinition.
func (c *Inquirer) Query(query []byte, credentials *CredentialsArg) (*VerifiablePresentation, error) {
	pdQuery, vcs, err := unwrapInputs(query, credentials)
	if err != nil {
		return nil, err
	}

	presentation, err := c.goAPICredentialQuery.Query(pdQuery,
		credentialquery.WithCredentialsArray(vcs),
		credentialquery.WithCredentialReader(&ReaderWrapper{
			CredentialReader: credentials.reader,
		}),
	)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return wrapVerifiablePresentation(presentation), err
}

// GetSubmissionRequirements returns information about VCs matching requirements.
func (c *Inquirer) GetSubmissionRequirements(query []byte, credentials *CredentialsArg,
) (*SubmissionRequirementArray, error) {
	pdQuery, vcs, err := unwrapInputs(query, credentials)
	if err != nil {
		return nil, err
	}

	requirements, err := c.goAPICredentialQuery.GetSubmissionRequirements(pdQuery,
		credentialquery.WithCredentialsArray(vcs),
		credentialquery.WithCredentialReader(&ReaderWrapper{
			CredentialReader: credentials.reader,
		}),
	)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &SubmissionRequirementArray{wrapped: requirements}, nil
}

func unwrapInputs(query []byte, credentials *CredentialsArg,
) (*presexch.PresentationDefinition, []*afgoverifiable.Credential, error) {
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

	if credentials.reader == nil && credentials.vcs == nil {
		return nil, nil, fmt.Errorf("either credential reader or vc array must be set")
	}

	var vcs []*afgoverifiable.Credential

	if credentials.vcs != nil {
		vcs = unwrapVCs(credentials.vcs)
	}

	return pdQuery, vcs, nil
}

func unwrapVCs(vcs *verifiable.CredentialsArray) []*afgoverifiable.Credential {
	var credentials []*afgoverifiable.Credential

	for i := 0; i < vcs.Length(); i++ {
		credentials = append(credentials, vcs.AtIndex(i).VC)
	}

	return credentials
}
