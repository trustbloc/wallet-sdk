/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package walletsdk implements a simplified interface to interop with JS.
package walletsdk

import (
	"encoding/json"
	"fmt"
	"net/http"

	jsonld "github.com/piprate/json-gold/ld"
	arieskms "github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/did"
	"github.com/trustbloc/vc-go/jwt"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/credentialquery"
	"github.com/trustbloc/wallet-sdk/pkg/credentialschema"
	"github.com/trustbloc/wallet-sdk/pkg/credentialstatus"
	"github.com/trustbloc/wallet-sdk/pkg/did/creator"
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
	"github.com/trustbloc/wallet-sdk/pkg/did/wellknown"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
	"github.com/trustbloc/wallet-sdk/pkg/memstorage/legacy"
	"github.com/trustbloc/wallet-sdk/pkg/openid4ci"
	"github.com/trustbloc/wallet-sdk/pkg/openid4vp"
)

// Agent is a facade around Wallet-SDK functionality. It provides a simplified interface to interop with JS.
type Agent struct {
	keyWriter   api.KeyWriter
	crypto      api.Crypto
	didResolver api.DIDResolver
	docLoader   jsonld.DocumentLoader
}

// NewAgent creates a new Agent.
func NewAgent(didResolverURI string, keyStore arieskms.Store) (*Agent, error) {
	localKMS, err := localkms.NewLocalKMS(localkms.Config{Storage: keyStore})
	if err != nil {
		return nil, fmt.Errorf("failed to create local kms: %w", err)
	}

	agent := &Agent{
		keyWriter: localKMS,
		crypto:    localKMS.GetCrypto(),
	}

	didResolver, err := resolver.NewDIDResolver(resolver.WithResolverServerURI(didResolverURI))
	if err != nil {
		return nil, fmt.Errorf("failed to create a did resolver: %w", err)
	}

	agent.didResolver = didResolver

	docLoader, err := common.CreateJSONLDDocumentLoader(&http.Client{}, legacy.NewProvider())
	if err != nil {
		return nil, fmt.Errorf("failed to create a did resolver: %w", err)
	}

	agent.docLoader = docLoader

	return agent, nil
}

// CreateDID creates a DID document using the given DID method.
func (a *Agent) CreateDID(didMethodType string, didKeyType arieskms.KeyType, verificationType string,
) (*did.DocResolution, error) {
	didCreator, err := creator.NewCreatorWithKeyWriter(a.keyWriter)
	if err != nil {
		return nil, fmt.Errorf("failed to create did creator: %w", err)
	}

	didDoc, err := didCreator.Create(didMethodType, &api.CreateDIDOpts{
		VerificationType: verificationType,
		KeyType:          didKeyType,
		MetricsLogger:    nil,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create did: %w", err)
	}

	return didDoc, nil
}

// CreateOpenID4CIIssuerInitiatedInteraction creates and starts openid4ci Interaction.
func (a *Agent) CreateOpenID4CIIssuerInitiatedInteraction(
	initiateIssuanceURI string,
) (*OpenID4CIIssuerInitiatedInteraction, error) {
	interaction, err := openid4ci.NewIssuerInitiatedInteraction(initiateIssuanceURI, &openid4ci.ClientConfig{
		DIDResolver: a.didResolver,
	})
	if err != nil {
		return nil, err
	}

	return &OpenID4CIIssuerInitiatedInteraction{
		Interaction: interaction,
		crypto:      a.crypto,
	}, nil
}

// CreateOpenID4VPInteraction creates and starts openid4vp Interaction.
func (a *Agent) CreateOpenID4VPInteraction(
	authorizationRequest string,
) (*OpenID4VPInteraction, error) {
	jwtVerifier := jwt.NewVerifier(jwt.KeyResolverFunc(
		common.NewVDRKeyResolver(
			a.didResolver,
		).PublicKeyFetcher()))

	interaction, err := openid4vp.NewInteraction(authorizationRequest, jwtVerifier, a.didResolver,
		a.crypto, a.docLoader)
	if err != nil {
		return nil, err
	}

	return &OpenID4VPInteraction{
		Interaction: interaction,
		DocLoader:   a.docLoader,
	}, nil
}

// ResolveDisplayData resolves display information for issued credentials based on an issuer's metadata,
// which is fetched using the issuer's (base) URI.
// The CredentialDisplays in the returned Data object correspond to the VCs passed in and are in the
// same order.
func (a *Agent) ResolveDisplayData(issuerURI string, credentials []string,
) (*credentialschema.ResolvedDisplayData, error) {
	var parsedCreds []*verifiable.Credential

	for _, cred := range credentials {
		verifiableCredential, err := verifiable.ParseCredential(
			[]byte(cred),
			verifiable.WithJSONLDDocumentLoader(a.docLoader),
			verifiable.WithDisabledProofCheck())
		if err != nil {
			return nil, fmt.Errorf("parse creds: %w", err)
		}

		parsedCreds = append(parsedCreds, verifiableCredential)
	}

	data, err := credentialschema.Resolve(
		credentialschema.WithIssuerURI(issuerURI),
		credentialschema.WithCredentials(parsedCreds),
	)
	if err != nil {
		return nil, fmt.Errorf("resolve data: %w", err)
	}

	return data, nil
}

// ParseResolvedDisplayData parses the given serialized display data into display data object.
func (a *Agent) ParseResolvedDisplayData(resolvedCredentialDisplayData string,
) (*credentialschema.ResolvedDisplayData, error) {
	var parsedDisplayData credentialschema.ResolvedDisplayData

	err := json.Unmarshal([]byte(resolvedCredentialDisplayData), &parsedDisplayData)
	if err != nil {
		return nil, err
	}

	return &parsedDisplayData, nil
}

// ParseCredential parses the given serialized VC into a VC object.
func (a *Agent) ParseCredential(credential string) (*verifiable.Credential, error) {
	verifiableCredential, err := verifiable.ParseCredential(
		[]byte(credential),
		verifiable.WithJSONLDDocumentLoader(a.docLoader),
		verifiable.WithDisabledProofCheck())
	if err != nil {
		return nil, fmt.Errorf("parse creds: %w", err)
	}

	return verifiableCredential, nil
}

// GetSubmissionRequirements returns information about VCs matching requirements.
func (a *Agent) GetSubmissionRequirements(query string, credentials []string,
) ([]*presexch.MatchedSubmissionRequirement, error) {
	credInquirer := credentialquery.NewInstance(a.docLoader)

	pdQuery, err := unwrapQuery([]byte(query))
	if err != nil {
		return nil, err
	}

	var parsedCreds []*verifiable.Credential

	for _, cred := range credentials {
		verifiableCredential, err := a.ParseCredential(cred)
		if err != nil {
			return nil, err
		}

		parsedCreds = append(parsedCreds, verifiableCredential)
	}

	return credInquirer.GetSubmissionRequirements(pdQuery, credentialquery.WithCredentialsArray(parsedCreds),
		credentialquery.WithSelectiveDisclosure(a.didResolver))
}

// VerifyCredentialsStatus checks the Credential Status, returning an error if the status field is invalid,
// the status is revoked, or if it isn't possible to verify the credential's status.
func (a *Agent) VerifyCredentialsStatus(credential string) error {
	parsedCred, err := a.ParseCredential(credential)
	if err != nil {
		return err
	}

	v, err := credentialstatus.NewVerifier(&credentialstatus.Config{
		HTTPClient:  &http.Client{},
		DIDResolver: a.didResolver,
	})
	if err != nil {
		return err
	}

	return v.Verify(parsedCred)
}

// ValidateLinkedDomains validates the given DID's Linked Domains service against its well-known DID configuration.
func (a *Agent) ValidateLinkedDomains(d string) (bool, string, error) {
	return wellknown.ValidateLinkedDomains(d, a.didResolver, &http.Client{})
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
