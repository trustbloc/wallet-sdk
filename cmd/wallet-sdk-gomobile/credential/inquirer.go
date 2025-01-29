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

	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/vc-go/presexch"
	afgoverifiable "github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/credentialquery"
	"github.com/trustbloc/wallet-sdk/pkg/memstorage/legacy"
)

// Inquirer implements querying credentials using presentation definition.
type Inquirer struct {
	goAPICredentialQuery *credentialquery.Instance
	goDIDResolver        goapi.DIDResolver
}

// NewInquirer returns a new Inquirer.
func NewInquirer(opts *InquirerOpts) (*Inquirer, error) {
	if opts == nil {
		opts = &InquirerOpts{}
	}

	var goAPIDocumentLoader ld.DocumentLoader

	if opts.documentLoader != nil {
		goAPIDocumentLoader = &wrapper.DocumentLoaderWrapper{
			DocumentLoader: opts.documentLoader,
		}
	} else {
		httpClient := &http.Client{}

		if opts.httpTimeout != nil {
			httpClient.Timeout = *opts.httpTimeout
		} else {
			httpClient.Timeout = goapi.DefaultHTTPTimeout
		}

		var err error

		goAPIDocumentLoader, err = common.CreateJSONLDDocumentLoader(httpClient, legacy.NewProvider())
		if err != nil {
			return nil, wrapper.ToMobileError(err)
		}
	}

	var goDIDResolver goapi.DIDResolver
	if opts.didResolver != nil {
		goDIDResolver = &wrapper.VDRResolverWrapper{DIDResolver: opts.didResolver}
	}

	return &Inquirer{
		goAPICredentialQuery: credentialquery.NewInstance(goAPIDocumentLoader),
		goDIDResolver:        goDIDResolver,
	}, nil
}

// GetSubmissionRequirements returns information about VCs matching requirements.
func (c *Inquirer) GetSubmissionRequirements(query []byte, credentials *verifiable.CredentialsArray,
) (*SubmissionRequirementArray, error) {
	if credentials == nil {
		credentials = verifiable.NewCredentialsArray()
	}

	pdQuery, err := unwrapQuery(query)
	if err != nil {
		return nil, err
	}

	requirements, err := c.goAPICredentialQuery.GetSubmissionRequirements(pdQuery,
		credentialquery.WithCredentialsArray(unwrapVCs(credentials)),
		credentialquery.WithSelectiveDisclosure(c.goDIDResolver))
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &SubmissionRequirementArray{wrapped: requirements}, nil
}

func unwrapQuery(query []byte) (*presexch.PresentationDefinition, error) {
	pdQuery := &presexch.PresentationDefinition{}

	err := json.Unmarshal(query, pdQuery)
	if err != nil {
		return nil, fmt.Errorf("unmarshal of presentation definition failed: %w", err)
	}

	err = pdQuery.ValidateSchema()
	if err != nil {
		return nil, fmt.Errorf("validation of presentation definition failed: %w", err)
	}

	return pdQuery, nil
}

func unwrapVCs(vcs *verifiable.CredentialsArray) []*afgoverifiable.Credential {
	var credentials []*afgoverifiable.Credential

	for i := range vcs.Length() {
		credentials = append(credentials, vcs.AtIndex(i).VC)
	}

	return credentials
}
