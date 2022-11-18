/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4vp contains functionality for doing OpenID4VP operations.
package openid4vp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/doc/jwt"
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	gomobilewrappers2 "github.com/trustbloc/wallet-sdk/internal/gomobilewrappers"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/openid4vp"
)

type goAPIOpenID4VP interface {
	GetQuery() (*presexch.PresentationDefinition, error)
	PresentCredential(presentation *verifiable.Presentation, jwtSigner goapi.JWTSigner) error
}

// Interaction represents a single OpenID4VP interaction between a wallet and a verifier. The methods defined on this
// object are used to help guide the calling code through the OpenID4VP flow.
type Interaction struct {
	keyHandleReader  api.KeyReader
	crypto           api.Crypto
	ldDocumentLoader api.LDDocumentLoader
	goAPIOpenID4VP   goAPIOpenID4VP
}

// NewInteraction creates a new OpenID4VP Interaction.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
func NewInteraction(authorizationRequest string, keyHandleReader api.KeyReader, crypto api.Crypto,
	didResolver api.DIDResolver, ldDocumentLoader api.LDDocumentLoader,
) *Interaction {
	jwtVerifier := jwt.NewVerifier(jwt.KeyResolverFunc(
		common.NewVDRKeyResolver(&gomobilewrappers2.VDRKeyResolverWrapper{
			DIDResolver: didResolver,
		}).PublicKeyFetcher()))

	return &Interaction{
		keyHandleReader:  keyHandleReader,
		ldDocumentLoader: ldDocumentLoader,
		crypto:           crypto,
		goAPIOpenID4VP:   openid4vp.New(authorizationRequest, jwtVerifier, &http.Client{}),
	}
}

// GetQuery creates query based on authorization request data.
func (o *Interaction) GetQuery() ([]byte, error) {
	presentationDefinition, err := o.goAPIOpenID4VP.GetQuery()
	if err != nil {
		return nil, err
	}

	pdBytes, err := json.Marshal(presentationDefinition)
	if err != nil {
		return nil, fmt.Errorf("presentation definition marshal: %w", err)
	}

	return pdBytes, nil
}

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredential(presentation []byte, kid string) error {
	signAlg, err := o.keyHandleReader.GetSigningAlgorithm(kid)
	if err != nil {
		return fmt.Errorf("get sign algorithm failed: %w", err)
	}

	parsedPresentation, err := verifiable.ParsePresentation(
		presentation,
		verifiable.WithPresDisabledProofCheck(),
		verifiable.WithPresJSONLDDocumentLoader(
			&gomobilewrappers2.DocumentLoaderWrapper{
				DocumentLoader: o.ldDocumentLoader,
			}))
	if err != nil {
		return fmt.Errorf("parse presentation failed: %w", err)
	}

	return o.goAPIOpenID4VP.PresentCredential(parsedPresentation, common.NewJWSSigner(kid, signAlg, o.crypto))
}
