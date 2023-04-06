/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4vp contains functionality for doing OpenID4VP operations.
package openid4vp

import (
	"encoding/json"
	"fmt"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/otel"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"

	"github.com/hyperledger/aries-framework-go/pkg/doc/jwt"
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	afgoverifiable "github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/openid4vp"
)

type goAPIOpenID4VP interface {
	GetQuery() (*presexch.PresentationDefinition, error)
	PresentCredential(credentials []*afgoverifiable.Credential) error
}

// Interaction represents a single OpenID4VP interaction between a wallet and a verifier. The methods defined on this
// object are used to help guide the calling code through the OpenID4VP flow.
type Interaction struct {
	crypto           api.Crypto
	ldDocumentLoader api.LDDocumentLoader
	goAPIOpenID4VP   goAPIOpenID4VP
	didResolver      api.DIDResolver
	inquirer         *credential.Inquirer
	oTel             *otel.Trace
}

// NewInteraction creates a new OpenID4VP Interaction.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
func NewInteraction(args *Args, opts *Opts) (*Interaction, error) { //nolint:funlen
	if opts == nil {
		opts = NewOpts()
	}

	var oTel *otel.Trace

	if !opts.disableOpenTelemetry {
		var err error

		oTel, err = otel.NewTrace()
		if err != nil {
			return nil, wrapper.ToMobileError(err)
		}

		opts.AddHeader(oTel.TraceHeader())
	}

	jwtVerifier := jwt.NewVerifier(jwt.KeyResolverFunc(
		common.NewVDRKeyResolver(&wrapper.VDRResolverWrapper{
			DIDResolver: args.didRes,
		}).PublicKeyFetcher()))

	httpClient := wrapper.NewHTTPClient()
	httpClient.AddHeaders(&opts.additionalHeaders)
	httpClient.DisableTLSVerification = opts.disableHTTPClientTLSVerification
	httpClient.Timeout = opts.httpTimeout

	goAPIOpts := []openid4vp.Opt{openid4vp.WithHTTPClient(httpClient)}

	if opts.activityLogger != nil {
		mobileActivityLoggerWrapper := &wrapper.MobileActivityLoggerWrapper{
			MobileAPIActivityLogger: opts.activityLogger,
		}

		goAPIOpts = append(goAPIOpts, openid4vp.WithActivityLogger(mobileActivityLoggerWrapper))
	}

	if opts.metricsLogger != nil {
		mobileMetricsLoggerWrapper := &wrapper.MobileMetricsLoggerWrapper{MobileAPIMetricsLogger: opts.metricsLogger}

		goAPIOpts = append(goAPIOpts, openid4vp.WithMetricsLogger(mobileMetricsLoggerWrapper))
	}

	var goAPIDocumentLoader ld.DocumentLoader

	if opts.documentLoader != nil {
		goAPIDocumentLoader = &wrapper.DocumentLoaderWrapper{
			DocumentLoader: opts.documentLoader,
		}
	}

	inquirerOpts := credential.NewInquirerOpts()
	inquirerOpts.SetDocumentLoader(opts.documentLoader)
	inquirer := credential.NewInquirer(inquirerOpts)

	return &Interaction{
		ldDocumentLoader: opts.documentLoader,
		crypto:           args.crypto,
		goAPIOpenID4VP: openid4vp.New(
			args.authorizationRequest,
			jwtVerifier,
			&wrapper.VDRResolverWrapper{DIDResolver: args.didRes},
			args.crypto,
			goAPIDocumentLoader,
			goAPIOpts...,
		),
		didResolver: args.didRes,
		inquirer:    inquirer,
		oTel:        oTel,
	}, nil
}

// GetQuery creates query based on authorization request data.
func (o *Interaction) GetQuery() ([]byte, error) {
	presentationDefinition, err := o.goAPIOpenID4VP.GetQuery()
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, o.oTel)
	}

	pdBytes, err := json.Marshal(presentationDefinition)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(
			fmt.Errorf("presentation definition marshal: %w", err), o.oTel)
	}

	return pdBytes, nil
}

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredential(credentials *verifiable.CredentialsArray) error {
	return wrapper.ToMobileErrorWithTrace(o.goAPIOpenID4VP.PresentCredential(unwrapVCs(credentials)), o.oTel)
}

// OTelTraceID returns open telemetry trace id.
func (o *Interaction) OTelTraceID() string {
	traceID := ""
	if o.oTel != nil {
		traceID = o.oTel.TraceID()
	}

	return traceID
}

func unwrapVCs(vcs *verifiable.CredentialsArray) []*afgoverifiable.Credential {
	var credentials []*afgoverifiable.Credential

	for i := 0; i < vcs.Length(); i++ {
		credentials = append(credentials, vcs.AtIndex(i).VC)
	}

	return credentials
}
