/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package walletsdk implements a simplified interface to interop with JS.
package walletsdk

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/component/models/did"
	arieskms "github.com/hyperledger/aries-framework-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/did/creator"
	"github.com/trustbloc/wallet-sdk/pkg/did/resolver"
	"github.com/trustbloc/wallet-sdk/pkg/localkms"
	"github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

// Agent is a facade around Wallet-SDK functionality. It provides a simplified interface to interop with JS.
type Agent struct {
	keyWriter   api.KeyWriter
	crypto      api.Crypto
	didResolver api.DIDResolver
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

// CreateOpenID4CIIssuerInitiatedInteraction creates and starts openid4ci issuer-initiated interaction.
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
		interaction: interaction,
		crypto:      a.crypto,
	}, nil
}
