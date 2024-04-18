/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4vp implements the OpenID4VP presentation flow.
package openid4vp

import (
	"bytes"
	"crypto/rand"
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
	wrapperapi "github.com/trustbloc/kms-go/wrapper/api"
	"github.com/trustbloc/vc-go/dataintegrity"
	"github.com/trustbloc/vc-go/dataintegrity/suite/ecdsa2019"
	"github.com/trustbloc/vc-go/jwt"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/proof/defaults"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	"github.com/trustbloc/wallet-sdk/pkg/did/wellknown"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
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

type didResolverWrapper struct {
	didResolver api.DIDResolver
}

func (d *didResolverWrapper) Resolve(did string, _ ...vdrapi.DIDMethodOption) (*diddoc.DocResolution, error) {
	return d.didResolver.Resolve(did)
}

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
	signer         wrapperapi.KMSCryptoSigner
}

type authorizedResponse struct {
	IDTokenJWS string
	VPTokenJWS string
	State      string
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
	client, activityLogger, metricsLogger, signer := processOpts(opts)

	var rawRequestObject string

	if strings.HasPrefix(authorizationRequest, "openid-vc://") {
		var err error

		rawRequestObject, err = fetchRequestObject(authorizationRequest, client, metricsLogger)
		if err != nil {
			return nil, err
		}
	} else {
		rawRequestObject = authorizationRequest
	}

	requestObject, err := verifyRequestObjectAndDecodeClaims(rawRequestObject, signatureVerifier)
	if err != nil {
		return nil, walleterror.NewValidationError(
			ErrorModule,
			InvalidAuthorizationRequestErrorCode,
			InvalidAuthorizationRequestError,
			fmt.Errorf("verify request object: %w", err))
	}

	return &Interaction{
		requestObject:  requestObject,
		httpClient:     client,
		activityLogger: activityLogger,
		metricsLogger:  metricsLogger,
		didResolver:    didResolver,
		crypto:         crypto,
		documentLoader: documentLoader,
		signer:         signer,
	}, nil
}

// GetQuery creates query based on authorization request data.
func (o *Interaction) GetQuery() *presexch.PresentationDefinition {
	return o.requestObject.Claims.VPToken.PresentationDefinition
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
		Name:    o.requestObject.Registration.ClientName,
		Purpose: o.requestObject.Registration.ClientPurpose,
		LogoURI: o.requestObject.Registration.ClientLogoURI,
	}
}

// TrustInfo return verifier trust info.
func (o *Interaction) TrustInfo() (*VerifierTrustInfo, error) {
	// Verifier is issuer of request object.
	verifier := o.requestObject.Issuer

	verifierDID := strings.Split(verifier, "#")[0]

	valid, linkedDomain, err := wellknown.ValidateLinkedDomains(verifierDID, o.didResolver, o.httpClient)
	if err != nil {
		return nil, err
	}

	return &VerifierTrustInfo{
		DID:         verifierDID,
		Domain:      linkedDomain,
		DomainValid: valid,
	}, nil
}

type presentOpts struct {
	ignoreConstraints bool
	signer            wrapperapi.KMSCryptoSigner

	attestationVPSigner api.JWTSigner
	attestationVC       string
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

// PresentCredential presents credentials to redirect uri from request object.
func (o *Interaction) PresentCredential(
	credentials []*verifiable.Credential,
	customClaims CustomClaims,
	opts ...PresentOpt,
) error {
	resolveOpts := &presentOpts{signer: o.signer}

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
func (o *Interaction) presentCredentials(
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
	data.Set("id_token", response.IDTokenJWS)
	data.Set("vp_token", response.VPTokenJWS)
	data.Set("state", response.State)

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
			Client:    o.requestObject.Registration.ClientName,
			Operation: activityLogOperation,
			Status:    api.ActivityLogStatusSuccess,
		},
	})
}

func (o *Interaction) sendAuthorizedResponse(responseBody string) error {
	_, err := httprequest.New(o.httpClient, o.metricsLogger).Do(http.MethodPost,
		o.requestObject.RedirectURI, "application/x-www-form-urlencoded",
		bytes.NewBufferString(responseBody),
		fmt.Sprintf(sendAuthorizedResponseEventText, o.requestObject.RedirectURI),
		presentCredentialEventText, processAuthorizationErrorResponse)

	return err
}

func fetchRequestObject(authorizationRequest string, client httpClient,
	metricsLogger api.MetricsLogger,
) (string, error) {
	authorizationRequestURL, err := url.Parse(authorizationRequest)
	if err != nil {
		return "", walleterror.NewValidationError(
			ErrorModule,
			InvalidAuthorizationRequestErrorCode,
			InvalidAuthorizationRequestError,
			err)
	}

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

func verifyRequestObjectAndDecodeClaims(
	rawRequestObject string,
	signatureVerifier jwt.ProofChecker,
) (*requestObject, error) {
	requestObject := &requestObject{}

	err := verifyTokenSignatureAndDecodeClaims(rawRequestObject, requestObject, signatureVerifier)
	if err != nil {
		return nil, err
	}

	return requestObject, nil
}

func verifyTokenSignatureAndDecodeClaims(rawJwt string, claims interface{}, proofChecker jwt.ProofChecker) error {
	jsonWebToken, _, err := jwt.ParseAndCheckProof(rawJwt, proofChecker, true,
		jwt.DecodeClaimsTo(claims),
		jwt.WithIgnoreClaimsMapDecoding(true))
	if err != nil {
		return fmt.Errorf("parse JWT: %w", err)
	}

	err = jsonWebToken.DecodeClaims(claims)
	if err != nil {
		return fmt.Errorf("decode claims: %w", err)
	}

	return nil
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
			didResolver, crypto, documentLoader,
			opts.signer, opts)
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
	pd := requestObject.Claims.VPToken.PresentationDefinition

	if opts != nil && opts.ignoreConstraints {
		for i := range pd.InputDescriptors {
			pd.InputDescriptors[i].Constraints = nil
		}
	}

	bbsProofCreator := &verifiable.BBSProofCreator{
		ProofDerivation:            bbs12381g2pub.New(),
		VerificationMethodResolver: common.NewVDRKeyResolver(didResolver),
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
		presexch.WithDefaultPresentationFormat("jwt_vp"),
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

	signingVM, err := getSigningVM(did, didResolver)
	if err != nil {
		return nil, err
	}

	if opts != nil && opts.signer != nil {
		err = addDataIntegrityProof(
			fullVMID(did, signingVM.ID),
			didResolver,
			documentLoader,
			opts.signer,
			presentation,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add data integrity proof to VP: %w", err)
		}
	}

	jwtSigner, err := getHolderSigner(signingVM, crypto)
	if err != nil {
		return nil, err
	}

	presentationSubmission := presentation.CustomFields["presentation_submission"]

	presentation.CustomFields["presentation_submission"] = nil

	var attestationVP string
	if opts != nil && opts.attestationVC != "" {
		attestationVP, err = createAttestationVP(
			opts.attestationVC, opts.attestationVPSigner, documentLoader)
		if err != nil {
			return nil, err
		}
	}

	idTokenJWS, err := createIDToken(requestObject, presentationSubmission, did, customClaims, jwtSigner, attestationVP)
	if err != nil {
		return nil, err
	}

	vpTok := vpTokenClaims{
		VP:    presentation,
		Nonce: requestObject.Nonce,
		Exp:   time.Now().Unix() + tokenLiveTimeSec,
		Iss:   did,
		Aud:   requestObject.ClientID,
		Nbf:   time.Now().Unix(),
		Iat:   time.Now().Unix(),
		Jti:   uuid.NewString(),
	}

	vpTokenJWS, err := signToken(vpTok, jwtSigner)
	if err != nil {
		return nil, fmt.Errorf("sign vp_token: %w", err)
	}

	return &authorizedResponse{IDTokenJWS: idTokenJWS, VPTokenJWS: vpTokenJWS, State: requestObject.State}, nil
}

func createAuthorizedResponseMultiCred( //nolint:funlen,gocyclo // Unable to decompose without a major reworking
	credentials []*verifiable.Credential,
	requestObject *requestObject,
	customClaims CustomClaims,
	didResolver api.DIDResolver,
	crypto api.Crypto,
	documentLoader ld.DocumentLoader,
	signer wrapperapi.KMSCryptoSigner,
	opts *presentOpts,
) (*authorizedResponse, error) {
	pd := requestObject.Claims.VPToken.PresentationDefinition

	bbsProofCreator := &verifiable.BBSProofCreator{
		ProofDerivation:            bbs12381g2pub.New(),
		VerificationMethodResolver: common.NewVDRKeyResolver(didResolver),
	}

	presentations, submission, err := pd.CreateVPArray(
		credentials,
		documentLoader,
		presexch.WithSDBBSProofCreator(bbsProofCreator),
		presexch.WithSDCredentialOptions(
			verifiable.WithDisabledProofCheck(),
			verifiable.WithJSONLDDocumentLoader(documentLoader),
		),
		presexch.WithDefaultPresentationFormat("jwt_vp"),
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

		signingVM, e := getSigningVM(holderDID, didResolver)
		if e != nil {
			return nil, e
		}

		if signer != nil {
			e = addDataIntegrityProof(
				fullVMID(holderDID, signingVM.ID),
				didResolver,
				documentLoader,
				signer,
				presentation,
			)
			if e != nil {
				return nil, fmt.Errorf("failed to add data integrity proof to VP: %w", e)
			}
		}

		signer, e := getHolderSigner(signingVM, crypto)
		if e != nil {
			return nil, e
		}

		signers[holderDID] = signer

		// TODO: Fix this issue: the vpToken always uses the last presentation from the loop above
		vpTok := vpTokenClaims{
			VP:    presentation,
			Nonce: requestObject.Nonce,
			Exp:   time.Now().Unix() + tokenLiveTimeSec,
			Iss:   holderDID,
			Aud:   requestObject.ClientID,
			Nbf:   time.Now().Unix(),
			Iat:   time.Now().Unix(),
			Jti:   uuid.NewString(),
		}

		vpTokJWS, e := signToken(vpTok, signer)
		if e != nil {
			return nil, fmt.Errorf("sign vp_token: %w", e)
		}

		vpTokens = append(vpTokens, vpTokJWS)
	}

	vpTokenListJSON, err := json.Marshal(vpTokens)
	if err != nil {
		return nil, err
	}

	idTokenSigningDID, err := pickRandomElement(mapKeys(signers))
	if err != nil {
		return nil, err
	}

	var attestationVP string
	if opts.attestationVC != "" {
		attestationVP, err = createAttestationVP(
			opts.attestationVC, opts.attestationVPSigner, documentLoader)
		if err != nil {
			return nil, err
		}
	}

	idTokenJWS, err := createIDToken(requestObject, submission, idTokenSigningDID,
		customClaims, signers[idTokenSigningDID], attestationVP)
	if err != nil {
		return nil, err
	}

	return &authorizedResponse{
		IDTokenJWS: idTokenJWS,
		VPTokenJWS: string(vpTokenListJSON),
		State:      requestObject.State,
	}, nil
}

func addDataIntegrityProof(did string, didResolver api.DIDResolver, documentLoader ld.DocumentLoader,
	signer wrapperapi.KMSCryptoSigner, presentation *verifiable.Presentation,
) error {
	context := &verifiable.DataIntegrityProofContext{
		SigningKeyID: did,
		CryptoSuite:  ecdsa2019.SuiteType,
	}

	signerOpts := dataintegrity.Options{DIDResolver: &didResolverWrapper{didResolver: didResolver}}

	dataIntegritySigner, err := dataintegrity.NewSigner(&signerOpts,
		ecdsa2019.NewSignerInitializer(&ecdsa2019.SignerInitializerOptions{
			LDDocumentLoader: documentLoader,
			SignerGetter:     ecdsa2019.WithKMSCryptoWrapper(signer),
		}))
	if err != nil {
		return err
	}

	err = presentation.AddDataIntegrityProof(context, dataIntegritySigner)
	if err != nil {
		return err
	}

	return nil
}

func createIDToken(
	req *requestObject,
	submission interface{},
	signingDID string,
	customClaims CustomClaims,
	signer api.JWTSigner,
	attestationVP string,
) (string, error) {
	idToken := &idTokenClaims{
		VPToken: idTokenVPToken{
			PresentationSubmission: submission,
		},
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

	idTokenJWS, err := signToken(idToken, signer)
	if err != nil {
		return "", fmt.Errorf("sign id_token: %w", err)
	}

	return idTokenJWS, nil
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

func getSigningVM(holderDID string, didResolver api.DIDResolver) (*diddoc.VerificationMethod, error) {
	docRes, err := didResolver.Resolve(holderDID)
	if err != nil {
		return nil, fmt.Errorf("resolve holder DID for signing: %w", err)
	}

	verificationMethods := docRes.DIDDocument.VerificationMethods(diddoc.AssertionMethod)

	if len(verificationMethods[diddoc.AssertionMethod]) == 0 {
		return nil, fmt.Errorf("holder DID has no assertion method for signing")
	}

	signingVM := verificationMethods[diddoc.AssertionMethod][0].VerificationMethod

	return &signingVM, nil
}

func getHolderSigner(signingVM *diddoc.VerificationMethod, crypto api.Crypto) (api.JWTSigner, error) {
	return common.NewJWSSigner(models.VerificationMethodFromDoc(signingVM), crypto)
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

func fullVMID(did, vmID string) string {
	if vmID == "" {
		return did
	}

	if vmID[0] == '#' {
		return did + vmID
	}

	if strings.HasPrefix(vmID, "did:") {
		return vmID
	}

	return did + "#" + vmID
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
			detailedErr)
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
