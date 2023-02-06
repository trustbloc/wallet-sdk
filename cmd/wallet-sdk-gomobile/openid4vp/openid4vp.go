/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4vp contains functionality for doing OpenID4VP operations.
package openid4vp

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/jwt"
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/walleterror"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/openid4vp"
	gowalleterror "github.com/trustbloc/wallet-sdk/pkg/walleterror"
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

// ClientConfig contains various parameters for an OpenID4VP Interaction.
// ActivityLogger is optional, but if provided then activities will be logged there.
// If not provided, then no activities will be logged.
type ClientConfig struct {
	KeyHandleReader api.KeyReader
	Crypto          api.Crypto
	DIDRes          api.DIDResolver
	DocumentLoader  api.LDDocumentLoader
	ActivityLogger  api.ActivityLogger
}

// NewClientConfig creates the client config object.
// ActivityLogger is optional, but if provided then activities will be logged there.
// If not provided, then no activities will be logged.
func NewClientConfig(keyHandleReader api.KeyReader, crypto api.Crypto,
	didResolver api.DIDResolver, ldDocumentLoader api.LDDocumentLoader, activityLogger api.ActivityLogger,
) *ClientConfig {
	return &ClientConfig{
		KeyHandleReader: keyHandleReader,
		Crypto:          crypto,
		DIDRes:          didResolver,
		DocumentLoader:  ldDocumentLoader,
		ActivityLogger:  activityLogger,
	}
}

// NewInteraction creates a new OpenID4VP Interaction.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// If activityLogger is nil, then no activity logging will take place.
func NewInteraction(authorizationRequest string, config *ClientConfig) *Interaction {
	jwtVerifier := jwt.NewVerifier(jwt.KeyResolverFunc(
		common.NewVDRKeyResolver(&wrapper.VDRResolverWrapper{
			DIDResolver: config.DIDRes,
		}).PublicKeyFetcher()))

	opts := []openid4vp.Opt{openid4vp.WithHTTPClient(common.DefaultHTTPClient())}

	if config.ActivityLogger != nil {
		mobileActivityLoggerWrapper := &wrapper.MobileActivityLoggerWrapper{MobileAPIActivityLogger: config.ActivityLogger}

		opts = append(opts, openid4vp.WithActivityLogger(mobileActivityLoggerWrapper))
	}

	return &Interaction{
		keyHandleReader:  config.KeyHandleReader,
		ldDocumentLoader: config.DocumentLoader,
		crypto:           config.Crypto,
		goAPIOpenID4VP:   openid4vp.New(authorizationRequest, jwtVerifier, opts...),
	}
}

// GetQuery creates query based on authorization request data.
func (o *Interaction) GetQuery() ([]byte, error) {
	presentationDefinition, err := o.goAPIOpenID4VP.GetQuery()
	if err != nil {
		return nil, walleterror.ToMobileError(err)
	}

	pdBytes, err := json.Marshal(presentationDefinition)
	if err != nil {
		return nil, walleterror.ToMobileError(
			fmt.Errorf("presentation definition marshal: %w", err))
	}

	return pdBytes, nil
}

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredential(presentation []byte, vm *api.VerificationMethod) error {
	signer, err := common.NewJWSSigner(vm.ToSDKVerificationMethod(), o.crypto)
	if err != nil {
		return walleterror.ToMobileError(err)
	}

	parsedPresentation, err := verifiable.ParsePresentation(
		presentation,
		verifiable.WithPresDisabledProofCheck(),
		verifiable.WithPresJSONLDDocumentLoader(
			&wrapper.DocumentLoaderWrapper{
				DocumentLoader: o.ldDocumentLoader,
			}))
	if err != nil {
		return walleterror.ToMobileError(
			gowalleterror.NewValidationError(module,
				CredentialParseFailedCode,
				CredentialParseFailedError,
				err,
			),
		)
	}

	return walleterror.ToMobileError(o.goAPIOpenID4VP.PresentCredential(parsedPresentation, signer))
}
