/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4vp contains functionality for doing OpenID4VP operations.
package openid4vp

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/proof/defaults"
	afgoverifiable "github.com/trustbloc/vc-go/verifiable"

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
	CustomScope() []string
	PresentCredential(
		credentials []*afgoverifiable.Credential,
		customClaims openid4vp.CustomClaims,
		opts ...openid4vp.PresentOpt,
	) error
	PresentedClaims(credential *afgoverifiable.Credential) (interface{}, error)
	PresentCredentialUnsafe(credential *afgoverifiable.Credential, customClaims openid4vp.CustomClaims) error
	VerifierDisplayData() *openid4vp.VerifierDisplayData
	TrustInfo() (*openid4vp.VerifierTrustInfo, error)
	Acknowledgment() *openid4vp.Acknowledgment
}

// VerifierTrustInfo represent verifier trust information.
type VerifierTrustInfo struct {
	DID         string
	Domain      string
	DomainValid bool
}

// CredentialClaimKeys represent credential claim keys.
type CredentialClaimKeys struct {
	ContentJSON interface{}
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

	goAPIOpts, err := toGoAPIOpts(opts)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	var goAPIDocumentLoader ld.DocumentLoader

	if opts.documentLoader != nil {
		goAPIDocumentLoader = &wrapper.DocumentLoaderWrapper{
			DocumentLoader: opts.documentLoader,
		}
	} else {
		dlHTTPClient := wrapper.NewHTTPClient(opts.httpTimeout, api.Headers{}, opts.disableHTTPClientTLSVerification)

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

	jwtVerifier := defaults.NewDefaultProofChecker(
		common.NewVDRKeyResolver(&wrapper.VDRResolverWrapper{
			DIDResolver: args.didRes,
		}))

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

// CustomScope returns vp integration scope.
func (o *Interaction) CustomScope() *Scope {
	return NewScope(o.goAPIOpenID4VP.CustomScope())
}

// TrustInfo return verifier trust info.
func (o *Interaction) TrustInfo() (*VerifierTrustInfo, error) {
	info, err := o.goAPIOpenID4VP.TrustInfo()
	if err != nil {
		return nil, err
	}

	return &VerifierTrustInfo{
		DID:         info.DID,
		Domain:      info.Domain,
		DomainValid: info.DomainValid,
	}, nil
}

// Acknowledgment returns acknowledgment object.
func (o *Interaction) Acknowledgment() *Acknowledgment {
	return &Acknowledgment{acknowledgment: o.goAPIOpenID4VP.Acknowledgment()}
}

// VerifierDisplayData returns display information about verifier.
func (o *Interaction) VerifierDisplayData() *VerifierDisplayData {
	displayData := o.goAPIOpenID4VP.VerifierDisplayData()

	return &VerifierDisplayData{displayData: displayData}
}

// PresentedClaims returns vc presented claims.
func (o *Interaction) PresentedClaims(credential *verifiable.Credential) (*CredentialClaimKeys, error) {
	claims, err := o.goAPIOpenID4VP.PresentedClaims(credential.VC)
	if err != nil {
		return nil, wrapper.ToMobileErrorWithTrace(err, o.oTel)
	}

	return &CredentialClaimKeys{ContentJSON: claims}, nil
}

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredential(credentials *verifiable.CredentialsArray) error {
	vcs, err := unwrapVCs(credentials)
	if err != nil {
		return wrapper.ToMobileErrorWithTrace(err, o.oTel)
	}

	return wrapper.ToMobileErrorWithTrace(o.goAPIOpenID4VP.PresentCredential(vcs, openid4vp.CustomClaims{}), o.oTel)
}

// PresentCredentialOpts presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredentialOpts(
	credentials *verifiable.CredentialsArray,
	opts *PresentCredentialOpts,
) error {
	vcs, err := unwrapVCs(credentials)
	if err != nil {
		return wrapper.ToMobileErrorWithTrace(err, o.oTel)
	}

	claims, err := getCustomClaims(opts)
	if err != nil {
		return err
	}

	var presentOpts []openid4vp.PresentOpt

	if opts != nil { //nolint:nestif
		if opts.serializedInteractionDetails != "" {
			var interactionDetails map[string]interface{}
			if err = json.Unmarshal([]byte(opts.serializedInteractionDetails), &interactionDetails); err != nil {
				return fmt.Errorf("decode vp interaction details: %w", err)
			}

			presentOpts = append(presentOpts, openid4vp.WithInteractionDetails(interactionDetails))
		}

		if opts.attestationVM != nil {
			attestationSigner, attErr := common.NewJWSSigner(opts.attestationVM.ToSDKVerificationMethod(), o.crypto)
			if attErr != nil {
				return wrapper.ToMobileErrorWithTrace(attErr, o.oTel)
			}

			presentOpts = append(presentOpts, openid4vp.WithAttestationVC(attestationSigner, opts.attestationVC))
		}
	}

	return wrapper.ToMobileErrorWithTrace(o.goAPIOpenID4VP.PresentCredential(vcs, claims, presentOpts...), o.oTel)
}

// PresentCredentialUnsafe presents a single credential to redirect uri from
// request object.
//
// Note: this variant of PresentCredential will skip client-side presentation
// definition constraint validation. All input descriptors will accept the
// provided credential, at least in terms of issuer fields, and subject data
// fields.
func (o *Interaction) PresentCredentialUnsafe(credential *verifiable.Credential) error {
	return wrapper.ToMobileErrorWithTrace(o.goAPIOpenID4VP.PresentCredentialUnsafe(credential.VC,
		openid4vp.CustomClaims{}), o.oTel)
}

// OTelTraceID returns open telemetry trace id.
func (o *Interaction) OTelTraceID() string {
	traceID := ""
	if o.oTel != nil {
		traceID = o.oTel.TraceID()
	}

	return traceID
}

//nolint:unparam
func toGoAPIOpts(opts *Opts) ([]openid4vp.Opt, error) {
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

	return goAPIOpts, nil
}

func unwrapVCs(vcs *verifiable.CredentialsArray) ([]*afgoverifiable.Credential, error) {
	if vcs == nil {
		return nil, errors.New("credentialsArray object cannot be nil")
	}

	var credentials []*afgoverifiable.Credential

	for i := range vcs.Length() {
		vc := vcs.AtIndex(i)
		if vc == nil {
			return nil, fmt.Errorf("credential objects cannot be nil (credential at index %d is nil)", i)
		}

		credentials = append(credentials, vc.VC)
	}

	return credentials, nil
}

func getCustomClaims(opts *PresentCredentialOpts) (openid4vp.CustomClaims, error) {
	if opts == nil {
		return openid4vp.CustomClaims{}, nil
	}

	claims := openid4vp.CustomClaims{
		ScopeClaims: map[string]interface{}{},
	}

	for key, value := range opts.scopeClaims {
		var jsonValue interface{}

		err := json.Unmarshal([]byte(value), &jsonValue)
		if err != nil {
			return openid4vp.CustomClaims{}, fmt.Errorf("fail to parse %q claim json: %w", key, err)
		}

		claims.ScopeClaims[key] = jsonValue
	}

	return claims, nil
}
