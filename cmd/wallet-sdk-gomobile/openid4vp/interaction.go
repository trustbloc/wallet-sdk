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

	"github.com/hyperledger/aries-framework-go/component/models/jwt"
	"github.com/hyperledger/aries-framework-go/component/models/presexch"
	afgoverifiable "github.com/hyperledger/aries-framework-go/component/models/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	credentialInquirer "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/otel"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/memstorage/legacy"
	"github.com/trustbloc/wallet-sdk/pkg/openid4vp"
)

type goAPIOpenID4VP interface {
	GetQuery() *presexch.PresentationDefinition
	PresentCredential(credentials []*afgoverifiable.Credential) error
	PresentCredentialUnsafe(credential *afgoverifiable.Credential) error
	VerifierDisplayData() *openid4vp.VerifierDisplayData
}

// Interaction represents a single OpenID4VP interaction between a wallet and a verifier. The methods defined on this
// object are used to help guide the calling code through the OpenID4VP flow.
type Interaction struct {
	crypto           api.Crypto
	ldDocumentLoader api.LDDocumentLoader
	goAPIOpenID4VP   goAPIOpenID4VP
	didResolver      api.DIDResolver
	inquirer         *credentialInquirer.Inquirer
	oTel             *otel.Trace
}

// NewInteraction creates a new OpenID4VP Interaction.
// The methods defined on this object are used to help guide the calling code through the OpenID4VP flow.
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

	goAPIOpts := toGoAPIOpts(opts)

	var goAPIDocumentLoader ld.DocumentLoader

	if opts.documentLoader != nil {
		goAPIDocumentLoader = &wrapper.DocumentLoaderWrapper{
			DocumentLoader: opts.documentLoader,
		}
	} else {
		dlHTTPClient := wrapper.NewHTTPClient(opts.httpTimeout, api.Headers{}, opts.disableHTTPClientTLSVerification)

		var err error
		goAPIDocumentLoader, err = common.CreateJSONLDDocumentLoader(dlHTTPClient, legacy.NewProvider())
		if err != nil {
			return nil, wrapper.ToMobileErrorWithTrace(err, oTel)
		}
	}

	inquirerOpts := credentialInquirer.NewInquirerOpts()
	inquirerOpts.SetDocumentLoader(opts.documentLoader)

	inquirer, err := credentialInquirer.NewInquirer(inquirerOpts)
	if err != nil {
		return nil, err
	}

	jwtVerifier := jwt.NewVerifier(jwt.KeyResolverFunc(
		common.NewVDRKeyResolver(&wrapper.VDRResolverWrapper{
			DIDResolver: args.didRes,
		}).PublicKeyFetcher()))

	goAPIInteraction, err := openid4vp.NewInteraction(
		args.authorizationRequest,
		jwtVerifier,
		&wrapper.VDRResolverWrapper{DIDResolver: args.didRes},
		args.crypto,
		goAPIDocumentLoader,
		goAPIOpts...,
	)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, oTel)
	}

	return &Interaction{
		ldDocumentLoader: opts.documentLoader,
		crypto:           args.crypto,
		goAPIOpenID4VP:   goAPIInteraction,
		didResolver:      args.didRes,
		inquirer:         inquirer,
		oTel:             oTel,
	}, nil
}

// GetQuery creates query based on authorization request data.
func (o *Interaction) GetQuery() ([]byte, error) {
	presentationDefinition := o.goAPIOpenID4VP.GetQuery()

	pdBytes, err := json.Marshal(presentationDefinition)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(
			fmt.Errorf("presentation definition marshal: %w", err), o.oTel)
	}

	return pdBytes, nil
}

// VerifierDisplayData returns display information about verifier.
func (o *Interaction) VerifierDisplayData() *VerifierDisplayData {
	displayData := o.goAPIOpenID4VP.VerifierDisplayData()

	return &VerifierDisplayData{displayData: displayData}
}

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredential(credentials *verifiable.CredentialsArray) error {
	return wrapper.ToMobileErrorWithTrace(o.goAPIOpenID4VP.PresentCredential(unwrapVCs(credentials)), o.oTel)
}

// PresentCredentialUnsafe presents a single credential to redirect uri from
// request object.
//
// Note: this variant of PresentCredential will skip client-side presentation
// definition constraint validation. All input descriptors will accept the
// provided credential, at least in terms of issuer fields, and subject data
// fields.
func (o *Interaction) PresentCredentialUnsafe(credential *verifiable.Credential) error {
	return wrapper.ToMobileErrorWithTrace(o.goAPIOpenID4VP.PresentCredentialUnsafe(credential.VC), o.oTel)
}

// OTelTraceID returns open telemetry trace id.
func (o *Interaction) OTelTraceID() string {
	traceID := ""
	if o.oTel != nil {
		traceID = o.oTel.TraceID()
	}

	return traceID
}

func toGoAPIOpts(opts *Opts) []openid4vp.Opt {
	httpClient := wrapper.NewHTTPClient(opts.httpTimeout, opts.additionalHeaders, opts.disableHTTPClientTLSVerification)

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

	if opts.kms != nil {
		goAPIOpts = append(goAPIOpts,
			openid4vp.WithDIProofs(opts.kms.GoAPILocalKMS.AriesCrypto, opts.kms.GoAPILocalKMS.AriesLocalKMS))
	}

	return goAPIOpts
}

func unwrapVCs(vcs *verifiable.CredentialsArray) []*afgoverifiable.Credential {
	var credentials []*afgoverifiable.Credential

	for i := 0; i < vcs.Length(); i++ {
		credentials = append(credentials, vcs.AtIndex(i).VC)
	}

	return credentials
}
