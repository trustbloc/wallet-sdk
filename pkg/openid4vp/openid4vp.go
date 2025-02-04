/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4vp implements the OpenID4VP presentation flow.
package openid4vp

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/bbs-signature-go/bbs12381g2pub"
	diddoc "github.com/trustbloc/did-go/doc/did"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	"github.com/trustbloc/vc-go/jwt"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/proof/defaults"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/did/wellknown"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	"github.com/trustbloc/wallet-sdk/pkg/ldproof"
	"github.com/trustbloc/wallet-sdk/pkg/models"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const (
	tokenLiveTimeSec = 600

	activityLogOperation = "oidc-presentation"

	newInteractionEventText         = "Instantiating OpenID4VP interaction object"
	fetchRequestObjectEventText     = "Fetch request object via an HTTP GET request to %s"
	presentCredentialEventText      = "Present credential" //nolint:gosec // false positive
	sendAuthorizedResponseEventText = "Send authorized response via an HTTP POST request to %s"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// VerifierTrustInfo represent verifier trust information.
type VerifierTrustInfo struct {
	DID         string
	Domain      string
	DomainValid bool
}

// Interaction is used to help with OpenID4VP operations.
type Interaction struct {
	requestObject  *requestObject
	httpClient     httpClient
	activityLogger api.ActivityLogger
	metricsLogger  api.MetricsLogger
	didResolver    api.DIDResolver
	crypto         api.Crypto
	documentLoader ld.DocumentLoader
}

type authorizedResponse struct {
	IDTokenJWS             string
	VPToken                string
	PresentationSubmission string
	State                  string
}

// NewInteraction creates a new OpenID4VP interaction object.
// If no ActivityLogger is provided (via an option), then no activity logging will take place.
func NewInteraction(
	authorizationRequest string,
	signatureVerifier jwt.ProofChecker,
	didResolver api.DIDResolver,
	crypto api.Crypto,
	documentLoader ld.DocumentLoader,
	opts ...Opt,
) (*Interaction, error) {
	client, activityLogger, metricsLogger := processOpts(opts)

	var (
		authorizationRequestClientID string
		rawRequestObject             string
	)

	if strings.HasPrefix(authorizationRequest, "openid-vc://") ||
		strings.HasPrefix(authorizationRequest, "openid4vp://") {
		var (
			authorizationRequestURL *url.URL
			err                     error
		)

		authorizationRequestURL, err = url.Parse(authorizationRequest)
		if err != nil {
			return nil, walleterror.NewValidationError(
				ErrorModule,
				InvalidAuthorizationRequestErrorCode,
				InvalidAuthorizationRequestError,
				err)
		}

		authorizationRequestClientID = authorizationRequestURL.Query().Get("client_id")

		rawRequestObject, err = fetchRequestObject(authorizationRequestURL, client, metricsLogger)
		if err != nil {
			return nil, err
		}
	} else {
		rawRequestObject = authorizationRequest
	}

	reqObject, err := parseRequestObject(authorizationRequestClientID, rawRequestObject, signatureVerifier)
	if err != nil {
		return nil, walleterror.NewValidationError(
			ErrorModule,
			InvalidAuthorizationRequestErrorCode,
			InvalidAuthorizationRequestError,
			fmt.Errorf("verify request object: %w", err))
	}

	return &Interaction{
		requestObject:  reqObject,
		httpClient:     client,
		activityLogger: activityLogger,
		metricsLogger:  metricsLogger,
		didResolver:    didResolver,
		crypto:         crypto,
		documentLoader: documentLoader,
	}, nil
}

// GetQuery creates query based on authorization request data.
func (o *Interaction) GetQuery() *presexch.PresentationDefinition {
	return o.requestObject.PresentationDefinition
}

// CustomScope returns vp integration scope.
func (o *Interaction) CustomScope() []string {
	var customScopes []string

	for _, scope := range strings.Split(o.requestObject.Scope, "+") {
		if scope != "openid" {
			customScopes = append(customScopes, scope)
		}
	}

	return customScopes
}

// VerifierDisplayData returns display information about verifier.
func (o *Interaction) VerifierDisplayData() *VerifierDisplayData {
	return &VerifierDisplayData{
		DID:     o.requestObject.ClientID,
		Name:    o.requestObject.ClientMetadata.ClientName,
		Purpose: o.requestObject.ClientMetadata.ClientPurpose,
		LogoURI: o.requestObject.ClientMetadata.ClientLogoURI,
	}
}

// TrustInfo return verifier trust info.
func (o *Interaction) TrustInfo() (*VerifierTrustInfo, error) {
	trustInfo := &VerifierTrustInfo{}

	if o.requestObject.ClientIDScheme == redirectURIScheme {
		verifierURI, err := url.Parse(o.requestObject.ResponseURI)
		if err != nil {
			return nil, err
		}

		trustInfo.Domain = verifierURI.Host
	} else {
		// Verifier is issuer of request object.
		verifier := o.requestObject.Issuer

		verifierDID := strings.Split(verifier, "#")[0]

		valid, linkedDomain, err := wellknown.ValidateLinkedDomains(verifierDID, o.didResolver, o.httpClient)
		if err != nil {
			return nil, err
		}

		trustInfo.DID = verifierDID
		trustInfo.Domain = linkedDomain
		trustInfo.DomainValid = valid
	}

	return trustInfo, nil
}

// Acknowledgment returns acknowledgment object for the current interaction.
func (o *Interaction) Acknowledgment() *Acknowledgment {
	return &Acknowledgment{
		ResponseURI: o.requestObject.ResponseURI,
		State:       o.requestObject.State,
	}
}

type presentOpts struct {
	ignoreConstraints bool

	attestationVPSigner api.JWTSigner
	attestationVC       string

	interactionDetails map[string]interface{}
}

// PresentOpt is an option for the RequestCredentialWithPreAuth method.
type PresentOpt func(opts *presentOpts)

// WithAttestationVC is an option for the RequestCredentialWithPreAuth method that allows you to specify
// attestation VC, which may be required by the verifier.
func WithAttestationVC(
	attestationVPSigner api.JWTSigner, vc string,
) PresentOpt {
	return func(opts *presentOpts) {
		opts.attestationVPSigner = attestationVPSigner
		opts.attestationVC = vc
	}
}

// WithInteractionDetails extends authorization response with interaction details.
func WithInteractionDetails(
	interactionDetails map[string]interface{},
) PresentOpt {
	return func(opts *presentOpts) {
		opts.interactionDetails = interactionDetails
	}
}

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredential(
	credentials []*verifiable.Credential,
	customClaims CustomClaims,
	opts ...PresentOpt,
) error {
	resolveOpts := &presentOpts{}

	for _, opt := range opts {
		if opt != nil {
			opt(resolveOpts)
		}
	}

	return o.presentCredentials(
		credentials,
		customClaims,
		resolveOpts,
	)
}

// PresentCredentialUnsafe presents a single credential to redirect uri from request object.
// This skips presentation definition constraint validation.
func (o *Interaction) PresentCredentialUnsafe(credential *verifiable.Credential, customClaims CustomClaims) error {
	return o.presentCredentials(
		[]*verifiable.Credential{credential},
		customClaims,
		&presentOpts{
			ignoreConstraints: true,
		},
	)
}

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) presentCredentials( //nolint: funlen
	credentials []*verifiable.Credential,
	customClaims CustomClaims,
	opts *presentOpts,
) error {
	timeStartPresentCredential := time.Now()

	response, err := createAuthorizedResponse(
		credentials,
		o.requestObject,
		customClaims,
		o.didResolver,
		o.crypto,
		o.documentLoader,
		opts,
	)
	if err != nil {
		return walleterror.NewExecutionError(
			ErrorModule,
			CreateAuthorizedResponseFailedCode,
			CreateAuthorizedResponseFailedError,
			fmt.Errorf("create authorized response failed: %w", err))
	}

	data := url.Values{}
	data.Set("presentation_submission", response.PresentationSubmission)
	data.Set("vp_token", response.VPToken)

	if response.IDTokenJWS != "" {
		data.Set("id_token", response.IDTokenJWS)
	}

	if response.State != "" {
		data.Set("state", response.State)
	}

	if opts.interactionDetails != nil {
		interactionDetailsBytes, e := json.Marshal(opts.interactionDetails)
		if e != nil {
			return fmt.Errorf("encode interaction details: %w", e)
		}

		data.Add("interaction_details", base64.StdEncoding.EncodeToString(interactionDetailsBytes))
	}

	err = o.sendAuthorizedResponse(data.Encode())
	if err != nil {
		return fmt.Errorf("send authorized response failed: %w", err)
	}

	err = o.metricsLogger.Log(&api.MetricsEvent{
		Event:    presentCredentialEventText,
		Duration: time.Since(timeStartPresentCredential),
	})
	if err != nil {
		return err
	}

	return o.activityLogger.Log(&api.Activity{
		ID:   uuid.New(),
		Type: api.LogTypeCredentialActivity,
		Time: time.Now(),
		Data: api.Data{
			Client:    o.requestObject.ClientMetadata.ClientName,
			Operation: activityLogOperation,
			Status:    api.ActivityLogStatusSuccess,
		},
	})
}

func (o *Interaction) PresentedClaims(credential *verifiable.Credential) (interface{}, error) {
	pd := o.requestObject.PresentationDefinition

	bbsProofCreator := &verifiable.BBSProofCreator{
		ProofDerivation:            bbs12381g2pub.New(),
		VerificationMethodResolver: common.NewVDRKeyResolver(o.didResolver),
	}

	presentation, err := pd.CreateVP(
		[]*verifiable.Credential{credential},
		o.documentLoader,
		presexch.WithSDBBSProofCreator(bbsProofCreator),
		presexch.WithSDCredentialOptions(
			verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(o.documentLoader),
			verifiable.WithProofChecker(defaults.NewDefaultProofChecker(common.NewVDRKeyResolver(o.didResolver))),
		),
		presexch.WithDefaultPresentationFormat("jwt_vp"),
	)
	if err != nil {
		return nil, err
	}

	vcContent := presentation.Credentials()[0].Contents()
	if len(vcContent.Subject) == 0 {
		return nil, errors.New("no subject in presentation VC")
	}

	return copyJSONKeysOnly(vcContent.Subject[0].CustomFields), nil
}

func (o *Interaction) sendAuthorizedResponse(responseBody string) error {
	_, err := httprequest.New(o.httpClient, o.metricsLogger).Do(http.MethodPost,
		o.requestObject.ResponseURI, "application/x-www-form-urlencoded",
		bytes.NewBufferString(responseBody),
		fmt.Sprintf(sendAuthorizedResponseEventText, o.requestObject.ResponseURI),
		presentCredentialEventText, processAuthorizationErrorResponse)

	return err
}

func fetchRequestObject(authorizationRequestURL *url.URL, client httpClient,
	metricsLogger api.MetricsLogger,
) (string, error) {
	if !authorizationRequestURL.Query().Has("request_uri") {
		return "", walleterror.NewValidationError(
			ErrorModule,
			InvalidAuthorizationRequestErrorCode,
			InvalidAuthorizationRequestError,
			errors.New("request_uri missing from authorization request URI"))
	}

	requestURI := authorizationRequestURL.Query().Get("request_uri")

	respBytes, err := httprequest.New(client, metricsLogger).Do(http.MethodGet, requestURI, "", nil,
		fmt.Sprintf(fetchRequestObjectEventText, requestURI), newInteractionEventText, nil)
	if err != nil {
		return "", walleterror.NewExecutionError(
			ErrorModule,
			RequestObjectFetchFailedCode,
			RequestObjectFetchFailedError,
			fmt.Errorf("fetch request object: %w", err))
	}

	return string(respBytes), nil
}

//nolint:gocyclo
func parseRequestObject(
	authorizationRequestClientID string,
	rawRequestObject string,
	signatureVerifier jwt.ProofChecker,
) (*requestObject, error) {
	reqObject := &requestObject{}

	_, _, err := jwt.Parse(rawRequestObject,
		jwt.DecodeClaimsTo(reqObject),
		jwt.WithIgnoreClaimsMapDecoding(true),
	)
	if err != nil {
		return nil, fmt.Errorf("parse jwt: %w", err)
	}

	switch reqObject.ClientIDScheme {
	case "", didScheme:
		if reqObject.Issuer == "" {
			return nil, errors.New("iss claim in request object is required")
		}

		err = jwt.CheckProof(rawRequestObject, signatureVerifier, &reqObject.Issuer, nil)
		if err != nil {
			return nil, fmt.Errorf("check proof: %w", err)
		}
	case redirectURIScheme:
		if !matchClientIDAndResponseURI(authorizationRequestClientID, reqObject.ResponseURI) {
			return nil, errors.New("client_id mismatch between authorization request and request object")
		}
	default:
		return nil, fmt.Errorf("unsupported client_id_scheme: %s", reqObject.ClientIDScheme)
	}

	// temporary solution for backward compatibility
	if reqObject.PresentationDefinition == nil && reqObject.Claims.VPToken.PresentationDefinition != nil {
		reqObject.PresentationDefinition = reqObject.Claims.VPToken.PresentationDefinition
	}

	if reqObject.ClientMetadata.VPFormats == nil && reqObject.Registration.VPFormats != nil {
		reqObject.ClientMetadata.ClientName = reqObject.Registration.ClientName
		reqObject.ClientMetadata.ClientPurpose = reqObject.Registration.ClientPurpose
		reqObject.ClientMetadata.ClientLogoURI = reqObject.Registration.LogoURI
		reqObject.ClientMetadata.VPFormats = reqObject.Registration.VPFormats
		reqObject.ClientMetadata.SubjectSyntaxTypesSupported = reqObject.Registration.SubjectSyntaxTypesSupported
	}

	if reqObject.ResponseURI == "" && reqObject.RedirectURI != "" {
		reqObject.ResponseURI = reqObject.RedirectURI
	}

	return reqObject, nil
}

func matchClientIDAndResponseURI(clientID, responseURI string) bool {
	clientIDURL, err := url.Parse(clientID)
	if err != nil {
		return false
	}

	responseURIURL, err := url.Parse(responseURI)
	if err != nil {
		return false
	}

	return strings.EqualFold(clientIDURL.Scheme, responseURIURL.Scheme) &&
		strings.EqualFold(clientIDURL.Host, responseURIURL.Host) &&
		strings.EqualFold(clientIDURL.Path, responseURIURL.Path)
}

func createAuthorizedResponse(
	credentials []*verifiable.Credential,
	requestObject *requestObject,
	customClaims CustomClaims,
	didResolver api.DIDResolver,
	crypto api.Crypto,
	documentLoader ld.DocumentLoader,
	opts *presentOpts,
) (*authorizedResponse, error) {
	switch len(credentials) {
	case 0:
		return nil, fmt.Errorf("expected at least one credential to present to verifier")
	case 1:
		return createAuthorizedResponseOneCred(credentials[0], requestObject, customClaims,
			didResolver, crypto, documentLoader, opts)
	default:
		return createAuthorizedResponseMultiCred(credentials, requestObject, customClaims,
			didResolver, crypto, documentLoader, opts)
	}
}

func createAuthorizedResponseOneCred( //nolint:funlen,gocyclo // Unable to decompose without a major reworking
	credential *verifiable.Credential,
	requestObject *requestObject,
	customClaims CustomClaims,
	didResolver api.DIDResolver,
	crypto api.Crypto,
	documentLoader ld.DocumentLoader,
	opts *presentOpts,
) (*authorizedResponse, error) {
	pd := requestObject.PresentationDefinition

	if opts != nil && opts.ignoreConstraints {
		for i := range pd.InputDescriptors {
			pd.InputDescriptors[i].Constraints = nil
		}
	}

	bbsProofCreator := &verifiable.BBSProofCreator{
		ProofDerivation:            bbs12381g2pub.New(),
		VerificationMethodResolver: common.NewVDRKeyResolver(didResolver),
	}

	vpFormat := presexch.FormatJWTVP

	if vpFormats := requestObject.ClientMetadata.VPFormats; vpFormats != nil {
		if vpFormats.LdpVP != nil && !credential.IsJWT() {
			vpFormat = presexch.FormatLDPVP
		}
	}

	presentation, err := pd.CreateVP(
		[]*verifiable.Credential{credential},
		documentLoader,
		presexch.WithSDBBSProofCreator(bbsProofCreator),
		presexch.WithSDCredentialOptions(
			verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(documentLoader),
			verifiable.WithProofChecker(defaults.NewDefaultProofChecker(common.NewVDRKeyResolver(didResolver))),
		),
		presexch.WithDefaultPresentationFormat(vpFormat),
	)
	if err != nil {
		return nil, err
	}

	did, err := verifiable.SubjectID(credential.Contents().Subject)
	if err != nil {
		return nil, fmt.Errorf("presentation VC does not have a subject ID: %w", err)
	}

	if did == "" {
		return nil, fmt.Errorf("presentation VC does not have a subject ID")
	}

	assertionVM, err := getAssertionVM(did, didResolver)
	if err != nil {
		return nil, err
	}

	jwtSigner, err := getHolderSigner(assertionVM, crypto)
	if err != nil {
		return nil, err
	}

	presentationSubmission := presentation.CustomFields["presentation_submission"]

	presentationSubmissionBytes, err := json.Marshal(presentationSubmission)
	if err != nil {
		return nil, fmt.Errorf("marshal presentation submission: %w", err)
	}

	presentation.CustomFields = nil

	// TODO: Resolve this properly in the vc-go
	presentation.ID = "urn:uuid:" + presentation.ID

	var vpToken string

	switch vpFormat { //nolint:dupl
	case presexch.FormatJWTVP:
		claims := vpTokenClaims{
			VP:    presentation,
			Nonce: requestObject.Nonce,
			Exp:   time.Now().Unix() + tokenLiveTimeSec,
			Iss:   did,
			Aud:   requestObject.ClientID,
			Nbf:   time.Now().Unix(),
			Iat:   time.Now().Unix(),
			Jti:   uuid.NewString(),
		}

		vpToken, err = signToken(claims, jwtSigner)
		if err != nil {
			return nil, fmt.Errorf("sign vp token: %w", err)
		}
	case presexch.FormatLDPVP:
		vpToken, err = createLdpVPToken(crypto, documentLoader, didResolver, did, assertionVM, requestObject, presentation)
		if err != nil {
			return nil, fmt.Errorf("create ldp vp token: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported presentation exchange format: %s", vpFormat)
	}

	var idTokenJWS string

	if strings.Contains(requestObject.ResponseType, "id_token") {
		var attestationVP string

		if opts != nil && opts.attestationVC != "" {
			attestationVP, err = createAttestationVP(
				opts.attestationVC, opts.attestationVPSigner, documentLoader)
			if err != nil {
				return nil, err
			}
		}

		idTokenJWS, err = createIDToken(requestObject, did, customClaims, jwtSigner, attestationVP, presentationSubmission)
		if err != nil {
			return nil, err
		}
	}

	return &authorizedResponse{
		PresentationSubmission: string(presentationSubmissionBytes),
		VPToken:                vpToken,
		IDTokenJWS:             idTokenJWS,
		State:                  requestObject.State,
	}, nil
}

func createAuthorizedResponseMultiCred( //nolint:funlen,gocyclo,gocognit
	credentials []*verifiable.Credential,
	requestObject *requestObject,
	customClaims CustomClaims,
	didResolver api.DIDResolver,
	crypto api.Crypto,
	documentLoader ld.DocumentLoader,
	opts *presentOpts,
) (*authorizedResponse, error) {
	pd := requestObject.PresentationDefinition

	bbsProofCreator := &verifiable.BBSProofCreator{
		ProofDerivation:            bbs12381g2pub.New(),
		VerificationMethodResolver: common.NewVDRKeyResolver(didResolver),
	}

	vpFormat := presexch.FormatJWTVP

	if vpFormats := requestObject.ClientMetadata.VPFormats; vpFormats != nil {
		isJWTCredential := false

		for _, credential := range credentials {
			if credential.IsJWT() {
				isJWTCredential = true

				break
			}
		}

		if vpFormats.LdpVP != nil && !isJWTCredential {
			vpFormat = presexch.FormatLDPVP
		}
	}

	presentations, presentationSubmission, err := pd.CreateVPArray(
		credentials,
		documentLoader,
		presexch.WithSDBBSProofCreator(bbsProofCreator),
		presexch.WithSDCredentialOptions(
			verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(documentLoader),
		),
		presexch.WithDefaultPresentationFormat(vpFormat),
	)
	if err != nil {
		return nil, err
	}

	var vpTokens []string

	signers := map[string]api.JWTSigner{}

	for _, presentation := range presentations {
		holderDID, e := getSubjectID(presentation.Credentials()[0])
		if e != nil {
			return nil, fmt.Errorf("presentation VC does not have a subject ID: %w", e)
		}

		assertionVM, e := getAssertionVM(holderDID, didResolver)
		if e != nil {
			return nil, e
		}

		jwtSigner, e := getHolderSigner(assertionVM, crypto)
		if e != nil {
			return nil, e
		}

		signers[holderDID] = jwtSigner

		var vpToken string

		switch vpFormat { //nolint:dupl
		case presexch.FormatJWTVP:
			// TODO: Fix this issue: the vpToken always uses the last presentation from the loop above
			claims := vpTokenClaims{
				VP:    presentation,
				Nonce: requestObject.Nonce,
				Exp:   time.Now().Unix() + tokenLiveTimeSec,
				Iss:   holderDID,
				Aud:   requestObject.ClientID,
				Nbf:   time.Now().Unix(),
				Iat:   time.Now().Unix(),
				Jti:   uuid.NewString(),
			}

			vpToken, err = signToken(claims, jwtSigner)
			if err != nil {
				return nil, fmt.Errorf("sign vp token: %w", err)
			}
		case presexch.FormatLDPVP:
			vpToken, err = createLdpVPToken(crypto, documentLoader, didResolver, holderDID, assertionVM, requestObject,
				presentation)
			if err != nil {
				return nil, fmt.Errorf("create ldp vp token: %w", err)
			}
		default:
			return nil, fmt.Errorf("unsupported presentation exchange format: %s", vpFormat)
		}

		vpTokens = append(vpTokens, vpToken)
	}

	vpTokenListJSON, err := json.Marshal(vpTokens)
	if err != nil {
		return nil, err
	}

	var idTokenJWS string

	//nolint:nestif
	if strings.Contains(requestObject.ResponseType, "id_token") {
		var idTokenSigningDID string

		idTokenSigningDID, err = pickRandomElement(mapKeys(signers))
		if err != nil {
			return nil, err
		}

		var attestationVP string

		if opts.attestationVC != "" {
			attestationVP, err = createAttestationVP(opts.attestationVC, opts.attestationVPSigner, documentLoader)
			if err != nil {
				return nil, err
			}
		}

		idTokenJWS, err = createIDToken(requestObject, idTokenSigningDID, customClaims,
			signers[idTokenSigningDID], attestationVP, presentationSubmission)
		if err != nil {
			return nil, err
		}
	}

	presentationSubmissionJSON, err := json.Marshal(presentationSubmission)
	if err != nil {
		return nil, fmt.Errorf("marshal presentation submission: %w", err)
	}

	return &authorizedResponse{
		IDTokenJWS:             idTokenJWS,
		VPToken:                string(vpTokenListJSON),
		PresentationSubmission: string(presentationSubmissionJSON),
		State:                  requestObject.State,
	}, nil
}

func createLdpVPToken(
	crypto api.Crypto,
	documentLoader ld.DocumentLoader,
	didResolver api.DIDResolver,
	did string,
	assertionVM *diddoc.VerificationMethod,
	requestObject *requestObject,
	presentation *verifiable.Presentation,
) (string, error) {
	vpFormats := requestObject.ClientMetadata.VPFormats

	if vpFormats == nil || vpFormats.LdpVP == nil {
		return "", errors.New("client does not support ldp_vp format")
	}

	ldProof := ldproof.New(
		crypto,
		documentLoader,
		didResolver,
	)

	err := ldProof.Add(
		presentation,
		ldproof.WithLdpType(vpFormats.LdpVP),
		ldproof.WithVerificationMethod(assertionVM),
		ldproof.WithDID(did),
		ldproof.WithChallenge(requestObject.Nonce),
		ldproof.WithDomain(requestObject.ClientID),
	)
	if err != nil {
		return "", fmt.Errorf("add ld proof to vp: %w", err)
	}

	var vpBytes []byte

	vpBytes, err = presentation.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("marshal vp into vp token: %w", err)
	}

	return string(vpBytes), nil
}

func createIDToken(
	req *requestObject,
	signingDID string,
	customClaims CustomClaims,
	signer api.JWTSigner,
	attestationVP string,
	presentationSubmission interface{},
) (string, error) {
	idToken := &idTokenClaims{
		Scope:         customClaims.ScopeClaims,
		AttestationVP: attestationVP,
		Nonce:         req.Nonce,
		Exp:           time.Now().Unix() + tokenLiveTimeSec,
		Iss:           "https://self-issued.me/v2/openid-vc",
		Sub:           signingDID,
		Aud:           req.ClientID,
		Nbf:           time.Now().Unix(),
		Iat:           time.Now().Unix(),
		Jti:           uuid.NewString(),
	}

	if ps, ok := presentationSubmission.(*presexch.PresentationSubmission); ok {
		var psCopy presexch.PresentationSubmission

		if err := deepCopy(ps, &psCopy); err != nil {
			return "", fmt.Errorf("deep copy presentation submission: %w", err)
		}

		for _, descriptor := range psCopy.DescriptorMap {
			if descriptor.PathNested != nil {
				descriptor.PathNested.Path = strings.Replace(descriptor.PathNested.Path, "$.vp.", "$.", 1)
			}
		}

		idToken.VPToken = idTokenVPToken{
			PresentationSubmission: psCopy,
		}
	} else {
		idToken.VPToken = idTokenVPToken{
			PresentationSubmission: presentationSubmission,
		}
	}

	idTokenJWS, err := signToken(idToken, signer)
	if err != nil {
		return "", fmt.Errorf("sign id_token: %w", err)
	}

	return idTokenJWS, nil
}

func deepCopy(src, dst interface{}) error {
	var b bytes.Buffer

	if err := gob.NewEncoder(&b).Encode(src); err != nil {
		return err
	}

	return gob.NewDecoder(bytes.NewBuffer(b.Bytes())).Decode(dst)
}

func createAttestationVP(
	attestationVCData string,
	attestationVPSigner api.JWTSigner,
	documentLoader ld.DocumentLoader,
) (string, error) {
	attestationVC, err := verifiable.ParseCredential([]byte(attestationVCData),
		verifiable.WithDisabledProofCheck(),
		verifiable.WithJSONLDDocumentLoader(documentLoader))
	if err != nil {
		return "", err
	}

	attestationVP, err := verifiable.NewPresentation(verifiable.WithCredentials(attestationVC))
	if err != nil {
		return "", err
	}

	attestationVP.ID = uuid.New().String()

	claims, err := attestationVP.JWTClaims([]string{}, false)
	if err != nil {
		return "", err
	}

	return signToken(claims, attestationVPSigner)
}

func signToken(claims interface{}, signer api.JWTSigner) (string, error) {
	token, err := jwt.NewSigned(claims, jwt.SignParameters{}, signer)
	if err != nil {
		return "", fmt.Errorf("sign token failed: %w", err)
	}

	tokenBytes, err := token.Serialize(false)
	if err != nil {
		return "", fmt.Errorf("serialize token failed: %w", err)
	}

	return tokenBytes, nil
}

func getAssertionVM(holderDID string, didResolver api.DIDResolver) (*diddoc.VerificationMethod, error) {
	docRes, err := didResolver.Resolve(holderDID)
	if err != nil {
		return nil, fmt.Errorf("resolve holder DID for assertion method: %w", err)
	}

	verificationMethods := docRes.DIDDocument.VerificationMethods(diddoc.AssertionMethod)

	if len(verificationMethods[diddoc.AssertionMethod]) == 0 {
		return nil, fmt.Errorf("holder DID has no assertion method for signing")
	}

	assertionVM := verificationMethods[diddoc.AssertionMethod][0].VerificationMethod

	return &assertionVM, nil
}

func getHolderSigner(vm *diddoc.VerificationMethod, crypto api.Crypto) (api.JWTSigner, error) {
	return common.NewJWSSigner(models.VerificationMethodFromDoc(vm), crypto)
}

func getSubjectID(vc *verifiable.Credential) (string, error) {
	return verifiable.SubjectID(vc.Contents().Subject)
}

func mapKeys(in map[string]api.JWTSigner) []string {
	var keys []string

	for s := range in {
		keys = append(keys, s)
	}

	return keys
}

func pickRandomElement(list []string) (string, error) {
	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(list))))
	if err != nil {
		return "", err
	}

	return list[idx.Int64()], nil
}

type resolverAdapter struct {
	didResolver api.DIDResolver
}

func (r *resolverAdapter) Resolve(did string, _ ...vdrapi.DIDMethodOption) (*diddoc.DocResolution, error) {
	return r.didResolver.Resolve(did)
}

func wrapResolver(didResolver api.DIDResolver) *resolverAdapter {
	return &resolverAdapter{didResolver: didResolver}
}

func processAuthorizationErrorResponse(statusCode int, respBytes []byte) error {
	detailedErr := fmt.Errorf(
		"received status code [%d] with body [%s] in response to the authorization request",
		statusCode, string(respBytes))

	var errResponse errorResponse

	err := json.Unmarshal(respBytes, &errResponse)
	if err != nil {
		// Try interpreting the response using the MS Entra error response format.
		return processUsingMSEntraErrorResponseFormat(respBytes, detailedErr)
	}

	switch errResponse.Error {
	case "invalid_scope":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidScopeErrorCode,
			InvalidScopeError,
			detailedErr)
	case "invalid_request":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidRequestErrorCode,
			InvalidRequestError,
			detailedErr)
	case "invalid_client":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidClientErrorCode,
			InvalidClientError,
			detailedErr)
	case "vp_formats_not_supported":
		return walleterror.NewExecutionError(ErrorModule,
			VPFormatsNotSupportedErrorCode,
			VPFormatsNotSupportedError,
			detailedErr)
	case "invalid_presentation_definition_uri":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidPresentationDefinitionURIErrorCode,
			InvalidPresentationDefinitionURIError,
			detailedErr)
	case "invalid_presentation_definition_reference":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidPresentationDefinitionReferenceErrorCode,
			InvalidPresentationDefinitionReferenceError,
			detailedErr)
	default:
		return walleterror.NewExecutionError(ErrorModule,
			OtherAuthorizationResponseErrorCode,
			OtherAuthorizationResponseError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	}
}

func processUsingMSEntraErrorResponseFormat(respBytes []byte, detailedErr error) error {
	var errorResponse msEntraErrorResponse

	err := json.Unmarshal(respBytes, &errorResponse)
	if err != nil {
		return walleterror.NewExecutionError(ErrorModule,
			OtherAuthorizationResponseErrorCode,
			OtherAuthorizationResponseError,
			detailedErr)
	}

	switch errorResponse.Error.InnerError.Code {
	case "badOrMissingField":
		return walleterror.NewExecutionErrorWithMessage(ErrorModule,
			MSEntraBadOrMissingFieldsErrorCode,
			MSEntraBadOrMissingFieldsError,
			errorResponse.Error.InnerError.Message,
			detailedErr)
	case "notFound":
		return walleterror.NewExecutionErrorWithMessage(ErrorModule,
			MSEntraNotFoundErrorCode,
			MSEntraNotFoundError, errorResponse.Error.InnerError.Message,
			detailedErr)
	case "tokenError":
		return walleterror.NewExecutionErrorWithMessage(ErrorModule,
			MSEntraTokenErrorCode,
			MSEntraTokenError,
			errorResponse.Error.InnerError.Message,
			detailedErr)
	case "transientError":
		return walleterror.NewExecutionErrorWithMessage(ErrorModule,
			MSEntraTransientErrorCode,
			MSEntraTransientError,
			errorResponse.Error.InnerError.Message,
			detailedErr)
	default:
		return walleterror.NewExecutionErrorWithMessage(ErrorModule,
			OtherAuthorizationResponseErrorCode,
			OtherAuthorizationResponseError,
			errorResponse.Error.InnerError.Message,
			detailedErr)
	}
}

type emptyObj struct{}

func copyJSONKeysOnly(obj interface{}) interface{} {
	empty := &emptyObj{}

	switch jsonObj := obj.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})

		for k, v := range jsonObj {
			newMap[k] = copyJSONKeysOnly(v)
		}

		return newMap
	case verifiable.CustomFields:
		newMap := make(map[string]interface{})
		populateClaimKeys(newMap, jsonObj)

		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(jsonObj))

		for i, v := range jsonObj {
			newSlice[i] = copyJSONKeysOnly(v)
		}

		return newSlice
	default:
		return empty
	}
}

func populateClaimKeys(claimKeys, doc map[string]interface{}) {
	for k, v := range doc {
		if k == "_sd" {
			continue
		}

		keys := make(map[string]interface{})

		claimKeys[k] = keys

		obj, ok := v.(map[string]interface{})
		if ok {
			populateClaimKeys(keys, obj)
		}
	}
}
