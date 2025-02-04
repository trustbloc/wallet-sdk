/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4ci provides APIs for wallets to receive verifiable credentials via OIDC for Credential Issuance.
package openid4ci

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/kms-go/doc/jose"
	"github.com/trustbloc/vc-go/jwt"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const (
	activityLogOperation        = "oidc-issuance"
	jwtVCJSONCredentialFormat   = "jwt_vc_json"    //nolint:gosec // false positive
	jwtVCJSONLDCredentialFormat = "jwt_vc_json-ld" //nolint:gosec // false positive
	ldpVCCredentialFormat       = "ldp_vc"

	newInteractionEventText = "Instantiating OpenID4CI interaction object"
	//nolint:gosec //false positive
	fetchCredOfferViaGETReqEventText = "Fetch credential offer via an HTTP GET request to %s"

	//nolint:gosec //false positive
	requestCredentialEventText = "Request credential(s) from issuer"
	//nolint:gosec //false positive
	fetchTokenViaPOSTReqEventText = "Fetch token via an HTTP POST request to %s"
	//nolint:gosec //false positive
	fetchCredentialViaGETReqEventText  = "Fetch credential %d of %d via an HTTP POST request to %s"
	parseAndCheckProofCheckVCEventText = "Parsing and checking proof for received credential %d of %d"

	preAuthorizedGrantType     = "urn:ietf:params:oauth:grant-type:pre-authorized_code"
	authorizationCodeGrantType = "authorization_code"
	attestJWTClientAuthType    = "attest_jwt_client_auth"
)

// IssuerInitiatedInteraction represents a single issuer-instantiated OpenID4CI interaction between a wallet and an
// issuer. This type can be used if you have received a credential offer from an issuer in some form.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// An IssuerInitiatedInteraction is a stateful object, and is intended for going through the full flow only once
// after which it should be discarded. Any new interactions should use a fresh IssuerInitiatedInteraction instance.
type IssuerInitiatedInteraction struct {
	interaction *interaction

	credentialTypes              [][]string
	credentialFormats            []string
	credentialConfigIDs          []string
	credentialContexts           [][]string
	preAuthorizedCodeGrantParams *PreAuthorizedCodeGrantParams
	authorizationCodeGrantParams *AuthorizationCodeGrantParams

	authToken *universalAuthToken
}

// NewIssuerInitiatedInteraction creates a new OpenID4CI IssuerInitiatedInteraction.
// If no ActivityLogger is provided (via the ClientConfig object), then no activity logging will take place.
//
//nolint:funlen
func NewIssuerInitiatedInteraction(
	initiateIssuanceURI string,
	config *ClientConfig,
) (*IssuerInitiatedInteraction, error) {
	timeStartNewInteraction := time.Now()

	err := validateRequiredParameters(config)
	if err != nil {
		return nil, walleterror.NewInvalidSDKUsageError(ErrorModule, err)
	}

	setDefaults(config)

	credentialOffer, err := getCredentialOffer(initiateIssuanceURI, config.HTTPClient, config.MetricsLogger)
	if err != nil {
		return nil, err
	}

	issuerInteraction := &interaction{
		issuerURI:            credentialOffer.CredentialIssuer,
		didResolver:          config.DIDResolver,
		activityLogger:       config.ActivityLogger,
		metricsLogger:        config.MetricsLogger,
		disableVCProofChecks: config.DisableVCProofChecks,
		documentLoader:       config.DocumentLoader,
		httpClient:           config.HTTPClient,
	}

	err = issuerInteraction.populateIssuerMetadata(getIssuerMetadataEventText)
	if err != nil {
		return nil, err
	}

	// TODO https://github.com/trustbloc/wallet-sdk/issues/457 Add support for determining
	// grant types when no grants are specified.
	// See https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html#section-4.1.1 for more info.
	preAuthorizedCodeGrantParams, authorizationCodeGrantParams, err := determineIssuerGrantCapabilities(credentialOffer)
	if err != nil {
		return nil, err
	}

	credentialTypes, credentialFormats, credentialContexts, err := determineCredentialParameters(credentialOffer,
		issuerInteraction.issuerMetadata)
	if err != nil {
		return nil, err
	}

	return &IssuerInitiatedInteraction{
			interaction:                  issuerInteraction,
			preAuthorizedCodeGrantParams: preAuthorizedCodeGrantParams,
			authorizationCodeGrantParams: authorizationCodeGrantParams,
			credentialTypes:              credentialTypes,
			credentialFormats:            credentialFormats,
			credentialContexts:           credentialContexts,
			credentialConfigIDs:          credentialOffer.CredentialConfigurationIDs,
		},
		config.MetricsLogger.Log(
			&api.MetricsEvent{
				Event:    newInteractionEventText,
				Duration: time.Since(timeStartNewInteraction),
			},
		)
}

// CreateAuthorizationURL creates an authorization URL that can be opened in a browser to proceed to the login page.
// It is the first step in the authorization code flow.
// It creates the authorization URL that can be opened in a browser to proceed to the login page.
// This method can only be used if the issuer supports authorization code grants.
// Check the issuer's capabilities first using the methods available on this IssuerInitiatedInteraction object.
// If scopes are needed, pass them in using the WithScopes option.
func (i *IssuerInitiatedInteraction) CreateAuthorizationURL(clientID, redirectURI string,
	opts ...CreateAuthorizationURLOpt,
) (string, error) {
	if !i.AuthorizationCodeGrantTypeSupported() {
		return "", walleterror.NewInvalidSDKUsageError(ErrorModule,
			errors.New("issuer does not support the authorization code grant type"))
	}

	processedOpts := processCreateAuthorizationURLOpts(opts)

	issuerState, err := i.determineIssuerStateToUse(processedOpts.issuerState)
	if err != nil {
		return "", err
	}

	return i.interaction.createAuthorizationURL(clientID, redirectURI, i.credentialFormats[0], i.credentialTypes[0],
		i.credentialContexts[0], issuerState, processedOpts.scopes, processedOpts.useOAuthDiscoverableClientIDScheme)
}

// RequestCredentialWithPreAuth requests credential(s) from the issuer. This method can only be used for the
// pre-authorized code flow, where it acts as the final step in the interaction with the issuer.
// For the equivalent method for the authorization code flow, see RequestCredentialWithAuth instead.
// If a PIN is required (which can be checked via the PreAuthorizedCodeGrantParams method), then it must be passed
// into this method via the WithPIN option.
func (i *IssuerInitiatedInteraction) RequestCredentialWithPreAuth(jwtSigner api.JWTSigner,
	opts ...RequestCredentialWithPreAuthOpt,
) ([]*verifiable.Credential, error) {
	processedOpts := processRequestCredentialWithPreAuthOpts(opts)

	if i.PreAuthorizedCodeGrantTypeSupported() {
		if i.preAuthorizedCodeGrantParams.PINRequired() && processedOpts.pin == "" {
			return nil, walleterror.NewInvalidSDKUsageError(ErrorModule,
				errors.New("the credential offer requires a user PIN, but none was provided"))
		}
	} else {
		return nil, walleterror.NewInvalidSDKUsageError(ErrorModule,
			errors.New("issuer does not support the pre-authorized code grant"))
	}

	return i.requestCredentialWithPreAuth(jwtSigner, processedOpts)
}

// RequestCredentialWithAuth requests credential(s) from the issuer. This method can only be used for the
// authorization code flow, where it acts as the final step in the interaction with the issuer.
// For the equivalent method for the pre-authorized code flow, see RequestCredentialWithPreAuth instead.
//
// RequestCredentialWithAuth should be called only once all authorization pre-requisite steps have been completed.
// The redirect URI that you pass in here should look like the redirect URI that you passed in to the
// CreateAuthorizationURL, except that now it has some URL query parameters appended to it.
func (i *IssuerInitiatedInteraction) RequestCredentialWithAuth(jwtSigner api.JWTSigner, redirectURIWithParams string,
) ([]*verifiable.Credential, error) {
	if !i.AuthorizationCodeGrantTypeSupported() {
		return nil, walleterror.NewInvalidSDKUsageError(ErrorModule,
			errors.New("issuer does not support the authorization code grant type"))
	}

	err := validateSignerKeyID(jwtSigner)
	if err != nil {
		return nil, err
	}

	err = i.interaction.requestAccessToken(redirectURIWithParams)
	if err != nil {
		return nil, err
	}

	return i.interaction.requestCredentialWithAuth(jwtSigner, i.credentialFormats, i.credentialTypes, i.credentialContexts)
}

// IssuerURI returns the issuer's URI from the initiation request. It's useful to store this somewhere in case
// there's a later need to refresh credential display data using the latest display information from the issuer.
func (i *IssuerInitiatedInteraction) IssuerURI() string {
	return i.interaction.issuerURI
}

// PreAuthorizedCodeGrantTypeSupported indicates whether the issuer supports the pre-authorized code grant type.
func (i *IssuerInitiatedInteraction) PreAuthorizedCodeGrantTypeSupported() bool {
	return i.preAuthorizedCodeGrantParams != nil
}

// PreAuthorizedCodeGrantParams returns an object that can be used to determine the issuer's pre-authorized code grant
// parameters. The caller should call the PreAuthorizedCodeGrantTypeSupported method first and only call this method to
// get the params if PreAuthorizedCodeGrantTypeSupported returns true.
// This method returns an error if (and only if) PreAuthorizedCodeGrantTypeSupported returns false.
func (i *IssuerInitiatedInteraction) PreAuthorizedCodeGrantParams() (*PreAuthorizedCodeGrantParams, error) {
	if i.preAuthorizedCodeGrantParams == nil {
		return nil, walleterror.NewInvalidSDKUsageError(ErrorModule,
			errors.New("issuer does not support the pre-authorized code grant"))
	}

	return i.preAuthorizedCodeGrantParams, nil
}

// CredentialConfigIDs returns the credential config IDs from the credential offer.
func (i *IssuerInitiatedInteraction) CredentialConfigIDs() []string {
	return i.credentialConfigIDs
}

// AuthorizationCodeGrantTypeSupported indicates whether the issuer supports the authorization code grant type.
func (i *IssuerInitiatedInteraction) AuthorizationCodeGrantTypeSupported() bool {
	return i.authorizationCodeGrantParams != nil
}

// AuthorizationCodeGrantParams returns an object that can be used to determine the issuer's authorization code grant
// parameters. The caller should call the AuthorizationCodeGrantTypeSupported method first and only call this method to
// get the params if AuthorizationCodeGrantTypeSupported returns true.
// This method returns an error if (and only if) AuthorizationCodeGrantTypeSupported returns false.
func (i *IssuerInitiatedInteraction) AuthorizationCodeGrantParams() (*AuthorizationCodeGrantParams, error) {
	if i.authorizationCodeGrantParams == nil {
		return nil, walleterror.NewInvalidSDKUsageError(ErrorModule,
			errors.New("issuer does not support the authorization code grant"))
	}

	return i.authorizationCodeGrantParams, nil
}

// DynamicClientRegistrationSupported indicates whether the issuer supports dynamic client registration.
func (i *IssuerInitiatedInteraction) DynamicClientRegistrationSupported() (bool, error) {
	return i.interaction.dynamicClientRegistrationSupported()
}

// DynamicClientRegistrationEndpoint returns the issuer's dynamic client registration endpoint.
// The caller should call the DynamicClientRegistrationSupported method first and only call this method
// if DynamicClientRegistrationSupported returns true.
// This method will return an error if the issuer does not support dynamic client registration.
func (i *IssuerInitiatedInteraction) DynamicClientRegistrationEndpoint() (string, error) {
	return i.interaction.dynamicClientRegistrationEndpoint()
}

// IssuerMetadata returns the issuer's metadata.
func (i *IssuerInitiatedInteraction) IssuerMetadata() (*issuer.Metadata, error) {
	err := i.interaction.populateIssuerMetadata(getIssuerMetadataEventText)
	if err != nil {
		return nil, err
	}

	return i.interaction.issuerMetadata, nil
}

// OfferedCredentialsTypes returns types of offered credentials.
func (i *IssuerInitiatedInteraction) OfferedCredentialsTypes() [][]string {
	return i.credentialTypes
}

// VerifyIssuer verifies the issuer via its issuer metadata. If successful, then the service URL is returned.
// An error means that either the issuer failed the verification check, or something went wrong during the
// process (and so a verification status could not be determined).
func (i *IssuerInitiatedInteraction) VerifyIssuer() (string, error) {
	return i.interaction.verifyIssuer()
}

// IssuerTrustInfo returns issuer trust info like, did, domain, credential type, format.
func (i *IssuerInitiatedInteraction) IssuerTrustInfo() (*IssuerTrustInfo, error) {
	return i.interaction.issuerFullTrustInfo(i.credentialTypes, i.credentialFormats)
}

// RequireAcknowledgment if true indicates that the issuer requires to be acknowledged if
// the user accepts or rejects credentials.
func (i *IssuerInitiatedInteraction) RequireAcknowledgment() (bool, error) {
	return i.interaction.requireAcknowledgment()
}

// Acknowledgment return not nil Acknowledgment if the issuer requires to be acknowledged that
// the user accepts or rejects credentials.
func (i *IssuerInitiatedInteraction) Acknowledgment() (*Acknowledgment, error) {
	authToken := i.interaction.authToken
	if i.authToken != nil {
		authToken = i.authToken
	}

	return i.interaction.requestedAcknowledgmentObj(authToken)
}

//nolint:funlen
func (i *IssuerInitiatedInteraction) requestCredentialWithPreAuth(jwtSigner api.JWTSigner,
	opts *requestCredentialWithPreAuthOpts,
) ([]*verifiable.Credential, error) {
	timeStartRequestCredential := time.Now()

	err := validateSignerKeyID(jwtSigner)
	if err != nil {
		return nil, err
	}

	err = i.interaction.populateIssuerMetadata(requestCredentialEventText)
	if err != nil {
		return nil, err
	}

	var attestationVP string

	if opts.attestationVC != "" {
		var issuerDID string

		jwtKID := i.interaction.issuerMetadata.GetJWTKID()

		if jwtKID != nil {
			issuerDID = strings.Split(*jwtKID, "#")[0]
		}

		attestationVP, err = createAttestationVP(
			opts.attestationVC,
			opts.attestationVPSigner,
			i.interaction.documentLoader,
			issuerDID,
			i.preAuthorizedCodeGrantParams.preAuthorizedCode,
		)
		if err != nil {
			return nil, err
		}
	}

	credentialResponses, err := i.getCredentialResponsesWithPreAuth(opts.pin, jwtSigner, attestationVP)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential response: %w", err)
	}

	vcs, err := i.interaction.getVCsFromCredentialResponses(credentialResponses)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			ErrorModule,
			CredentialParseFailedCode,
			CredentialParseError, err)
	}

	subjectIDs := getSubjectIDs(vcs)

	err = i.interaction.metricsLogger.Log(&api.MetricsEvent{
		Event:    requestCredentialEventText,
		Duration: time.Since(timeStartRequestCredential),
	})
	if err != nil {
		return nil, err
	}

	return vcs, i.interaction.activityLogger.Log(&api.Activity{
		ID:   uuid.New(),
		Type: api.LogTypeCredentialActivity,
		Time: time.Now(),
		Data: api.Data{
			Client:    i.interaction.issuerMetadata.CredentialIssuer,
			Operation: activityLogOperation,
			Status:    api.ActivityLogStatusSuccess,
			Params:    map[string]interface{}{"subjectIDs": subjectIDs},
		},
	})
}

func (i *IssuerInitiatedInteraction) getCredentialResponsesWithPreAuth(
	pin string, signer api.JWTSigner, attestationVP string,
) ([]CredentialResponse, error) {
	tokenEndpoint, err := i.interaction.getTokenEndpoint()
	if err != nil {
		return nil, err
	}

	tokenResponse, err := i.getPreAuthTokenResponse(pin, tokenEndpoint, attestationVP)
	if err != nil {
		return nil, fmt.Errorf("failed to get token response: %w", err)
	}

	i.authToken = &universalAuthToken{
		AccessToken: tokenResponse.AccessToken, TokenType: tokenResponse.TokenType,
		ExpiresAt: tokenResponse.expiry(), RefreshToken: tokenResponse.RefreshToken,
	}

	return i.getCredentialResponse(tokenResponse, tokenResponse.CNonce, signer, true)
}

//nolint:funlen,gocyclo,nonamedreturns
func (i *IssuerInitiatedInteraction) getCredentialResponse(
	tokenResponse *preAuthTokenResponse,
	nonce any,
	signer api.JWTSigner,
	allowRetry bool) (
	credentialResponse []CredentialResponse,
	err error,
) {
	proofJWT, err := i.interaction.createClaimsProof(nonce, signer)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err == nil || !allowRetry {
			return
		}

		proofError := &InvalidProofError{}
		if errors.As(err, &proofError) {
			credentialResponse, err = i.getCredentialResponse(tokenResponse, nonce, signer, false)
		}
	}()

	if len(i.credentialTypes) > 1 && i.interaction.issuerMetadata.BatchCredentialEndpoint != "" {
		return i.getCredentialResponsesBatch(proofJWT, tokenResponse)
	}

	credentialResponses := make([]CredentialResponse, len(i.credentialTypes))

	for index := range i.credentialTypes {
		requestBody, err := i.interaction.createCredentialRequestBody(proofJWT, i.credentialFormats[index],
			i.credentialTypes[index], i.credentialContexts[index])
		if err != nil {
			return nil, err
		}

		headers := http.Header{}
		headers.Add("Authorization", "Bearer "+tokenResponse.AccessToken)

		fetchCredentialResponseEventText := fmt.Sprintf(fetchCredentialViaGETReqEventText, index+1,
			len(i.credentialTypes), i.interaction.issuerMetadata.CredentialEndpoint)

		responseBytes, err := httprequest.New(i.interaction.httpClient, i.interaction.metricsLogger).DoContext(context.TODO(),
			http.MethodPost, i.interaction.issuerMetadata.CredentialEndpoint, "application/json", headers,
			bytes.NewReader(requestBody), fetchCredentialResponseEventText, requestCredentialEventText,
			[]int{http.StatusOK, http.StatusCreated}, processCredentialErrorResponse)
		if err != nil {
			return nil, err
		}

		var credentialResponse CredentialResponse

		err = json.Unmarshal(responseBytes, &credentialResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response from the issuer's credential endpoint: %w", err)
		}

		credentialResponses[index] = credentialResponse

		i.interaction.storeAcknowledgmentID(credentialResponse.AckID)
	}

	return credentialResponses, nil
}

//nolint:funlen
func (i *IssuerInitiatedInteraction) getCredentialResponsesBatch(
	proofJWT string,
	tokenResponse *preAuthTokenResponse,
) ([]CredentialResponse, error) {
	numberOfCredentials := len(i.credentialTypes)
	credentialResponses := make([]CredentialResponse, numberOfCredentials)

	batchCredentialReq := &batchCredentialRequest{
		CredentialRequests: make([]credentialRequest, numberOfCredentials),
	}

	for index := range i.credentialTypes {
		batchCredentialReq.CredentialRequests[index] = credentialRequest{
			CredentialDefinition: &credentialDefinition{
				Type: i.credentialTypes[index],
			},
			Format: i.credentialFormats[index],
			Proof: proof{
				ProofType: "jwt",
				JWT:       proofJWT,
			},
		}
	}

	b, err := json.Marshal(batchCredentialReq)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+tokenResponse.AccessToken)

	fetchCredentialResponseEventText := fmt.Sprintf(fetchCredentialViaGETReqEventText, numberOfCredentials,
		numberOfCredentials, i.interaction.issuerMetadata.BatchCredentialEndpoint)

	b, err = httprequest.New(i.interaction.httpClient, i.interaction.metricsLogger).DoContext(context.TODO(),
		http.MethodPost, i.interaction.issuerMetadata.BatchCredentialEndpoint, "application/json", headers,
		bytes.NewReader(b), fetchCredentialResponseEventText, requestCredentialEventText,
		[]int{http.StatusOK, http.StatusCreated}, processCredentialErrorResponse)
	if err != nil {
		return nil, err
	}

	var response batchCredentialResponse

	err = json.Unmarshal(b, &response)
	if err != nil {
		return nil, err
	}

	for index, credentialResp := range response.CredentialResponses {
		credentialResponses[index] = CredentialResponse{
			Credential:      credentialResp.Credential,
			TransactionID:   credentialResp.TransactionID,
			CNonce:          *response.CNonce,
			CNonceExpiresIn: *response.CNonceExpiresIn,
			AckID:           credentialResp.AckID,
		}

		i.interaction.storeAcknowledgmentID(credentialResp.AckID)
	}

	return credentialResponses, nil
}

func (i *IssuerInitiatedInteraction) getPreAuthTokenResponse(
	pin, tokenEndpoint, attestationVP string,
) (*preAuthTokenResponse, error) {
	params := url.Values{}
	params.Add("grant_type", preAuthorizedGrantType)
	params.Add("pre-authorized_code", i.preAuthorizedCodeGrantParams.preAuthorizedCode)

	if pin != "" {
		params.Add("tx_code", pin)
	}

	if attestationVP != "" {
		params.Add("client_assertion_type", attestJWTClientAuthType)
		params.Add("client_assertion", attestationVP)
	}

	paramsReader := strings.NewReader(params.Encode())

	responseBytes, err := httprequest.New(i.interaction.httpClient, i.interaction.metricsLogger).Do(
		http.MethodPost, tokenEndpoint, "application/x-www-form-urlencoded", paramsReader,
		fmt.Sprintf(fetchTokenViaPOSTReqEventText, tokenEndpoint),
		requestCredentialEventText, tokenErrorResponseHandler)
	if err != nil {
		return nil, fmt.Errorf("issuer's token endpoint %s: %w", attestationVP, err)
	}

	var tokenResp preAuthTokenResponse

	err = json.Unmarshal(responseBytes, &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from the issuer's token endpoint: %w", err)
	}

	return &tokenResp, nil
}

func (i *IssuerInitiatedInteraction) determineIssuerStateToUse(issuerStateFromOptions *string) (*string, error) {
	if i.authorizationCodeGrantParams.IssuerState != nil {
		// The spec says that if an issuer state is provided in the credential offer, then it must be used.
		// However, the spec also leaves open the possibility that the issuer state value could come from somewhere
		// else, so the option to set one is provided to the caller for such a case. To avoid confusion, if someone
		// sets an issuer state, but they shouldn't (because there's already one in the credential offer), then we
		// return an error.
		// While it's unnecessary to do so, if the caller specifies the same issuer state as what's in the credential
		// offer, then there's no conflict and no error is returned.
		if issuerStateFromOptions != nil && *i.authorizationCodeGrantParams.IssuerState != *issuerStateFromOptions {
			return nil, walleterror.NewInvalidSDKUsageError(ErrorModule,
				errors.New("the credential offer already specifies an issuer state, "+
					"and a conflicting issuer state value was provided. An issuer state should only be provided if "+
					"required by the issuer and the credential offer does not specify one already"))
		}

		return i.authorizationCodeGrantParams.IssuerState, nil
	}

	return issuerStateFromOptions, nil
}

func tokenErrorResponseHandler(statusCode int, respBody []byte) error {
	detailedErr := fmt.Errorf(
		"received status code [%d] with body [%s] from issuer's token endpoint", statusCode, respBody)

	var errResponse errorResponse

	err := json.Unmarshal(respBody, &errResponse)
	if err != nil {
		return walleterror.NewExecutionError(ErrorModule,
			OtherTokenResponseErrorCode,
			OtherTokenRequestError,
			detailedErr)
	}

	switch errResponse.Error {
	case "invalid_request":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidTokenRequestErrorCode,
			InvalidTokenRequestError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	case "invalid_grant":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidGrantErrorCode,
			InvalidGrantError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	case "invalid_client":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidClientErrorCode,
			InvalidClientError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	default:
		return walleterror.NewExecutionError(ErrorModule,
			OtherTokenResponseErrorCode,
			OtherTokenRequestError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	}
}

func getCredentialOffer(initiateIssuanceURI string, httpClient *http.Client, metricsLogger api.MetricsLogger,
) (*CredentialOffer, error) {
	requestURIParsed, err := url.Parse(initiateIssuanceURI)
	if err != nil {
		return nil, walleterror.NewValidationError(
			ErrorModule,
			InvalidIssuanceURICode,
			InvalidIssuanceURIError,
			err)
	}

	if requestURIParsed.Scheme != "openid-credential-offer" {
		return nil, walleterror.NewValidationError(
			ErrorModule,
			UnsupportedIssuanceURISchemeCode,
			UnsupportedIssuanceURISchemeError,
			fmt.Errorf("%s is not a supported issuance URL scheme", requestURIParsed.Scheme))
	}

	var credentialOfferJSON []byte

	switch {
	case requestURIParsed.Query().Has("credential_offer"):
		credentialOfferJSON = []byte(requestURIParsed.Query().Get("credential_offer"))
	case requestURIParsed.Query().Has("credential_offer_uri"):
		credentialOfferURI := requestURIParsed.Query().Get("credential_offer_uri")

		credentialOfferJSON, err = getCredentialOfferJSONFromCredentialOfferURI(
			credentialOfferURI, httpClient, metricsLogger)
		if err != nil {
			return nil, err
		}
	default:
		return nil,
			walleterror.NewValidationError(
				ErrorModule,
				InvalidIssuanceURICode,
				InvalidIssuanceURIError,
				errors.New("credential offer query parameter missing from initiate issuance URI"))
	}

	var credentialOffer CredentialOffer

	err = json.Unmarshal(credentialOfferJSON, &credentialOffer)
	if err != nil {
		return nil, walleterror.NewValidationError(
			ErrorModule,
			InvalidCredentialOfferCode,
			InvalidCredentialOfferError,
			fmt.Errorf("failed to unmarshal credential offer JSON into a credential offer object: %w", err))
	}

	return &credentialOffer, nil
}

func getCredentialOfferJSONFromCredentialOfferURI(credentialOfferURI string,
	httpClient *http.Client, metricsLogger api.MetricsLogger,
) ([]byte, error) {
	responseBytes, err := httprequest.New(httpClient, metricsLogger).Do(
		http.MethodGet, credentialOfferURI, "", nil,
		fmt.Sprintf(fetchCredOfferViaGETReqEventText, credentialOfferURI), newInteractionEventText, nil)
	if err != nil {
		return nil, walleterror.NewValidationError(
			ErrorModule,
			InvalidCredentialOfferCode,
			InvalidCredentialOfferError,
			fmt.Errorf("failed to get credential offer from the endpoint specified in the "+
				"credential_offer_uri URL query parameter: %w", err))
	}

	return responseBytes, nil
}

func determineCredentialParameters(
	credentialOffer *CredentialOffer,
	issuerMetadata *issuer.Metadata,
) ([][]string, []string, [][]string, error) {
	types := make([][]string, len(credentialOffer.CredentialConfigurationIDs))
	formats := make([]string, len(credentialOffer.CredentialConfigurationIDs))
	contexts := make([][]string, len(credentialOffer.CredentialConfigurationIDs))

	for i := range len(credentialOffer.CredentialConfigurationIDs) {
		id := credentialOffer.CredentialConfigurationIDs[i]

		configuration, ok := issuerMetadata.CredentialConfigurationsSupported[id]
		if !ok {
			return nil, nil, nil, walleterror.NewValidationError(
				ErrorModule,
				InvalidCredentialConfigurationIDCode,
				InvalidCredentialConfigurationIDError,
				fmt.Errorf("invalid credential configuration ID (%s) in credential offer", id),
			)
		}

		types[i] = configuration.CredentialDefinition.Type

		if configuration.Format != jwtVCJSONCredentialFormat &&
			configuration.Format != jwtVCJSONLDCredentialFormat &&
			configuration.Format != ldpVCCredentialFormat {
			return nil, nil, nil, walleterror.NewValidationError(
				ErrorModule,
				UnsupportedCredentialTypeInOfferCode,
				UnsupportedCredentialTypeInOfferError,
				fmt.Errorf("unsupported credential type (%s) in credential offer at index %d of "+
					"credential_configurations_supported (must be jwt_vc_json or jwt_vc_json-ld)",
					configuration.Format, i),
			)
		}

		formats[i] = configuration.Format

		if configuration.CredentialDefinition != nil && configuration.Format == ldpVCCredentialFormat {
			contexts[i] = configuration.CredentialDefinition.Context
		}
	}

	return types, formats, contexts, nil
}

func validateSignerKeyID(jwtSigner api.JWTSigner) error {
	kidParts := strings.Split(jwtSigner.GetKeyID(), "#")
	if len(kidParts) < 2 { //nolint: mnd
		return walleterror.NewExecutionError(
			ErrorModule,
			KeyIDMissingDIDPartCode,
			KeyIDMissingDIDPartError,
			fmt.Errorf("key ID (%s) is missing the DID part", jwtSigner.GetKeyID()))
	}

	return nil
}

func getSubjectIDs(vcs []*verifiable.Credential) []string {
	var subjectIDs []string

	for i := range vcs {
		subjects := vcs[i].Contents().Subject

		for j := range subjects {
			subjectIDs = append(subjectIDs, subjects[j].ID)
		}
	}

	return subjectIDs
}

func createAttestationVP(
	attestationVCData string,
	attestationVPSigner api.JWTSigner,
	documentLoader ld.DocumentLoader,
	issuerDID,
	nonce string,
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

	if nonce != "" {
		attestationVP.CustomFields = map[string]interface{}{
			"nonce": nonce,
		}
	}

	claims, err := attestationVP.JWTClaims([]string{issuerDID}, false)
	if err != nil {
		return "", err
	}

	return signToken(claims, attestationVPSigner)
}

func signToken(claims interface{}, signer api.JWTSigner) (string, error) {
	headers := jose.Headers{}
	headers["typ"] = "openid4vci-proof+jwt"

	token, err := jwt.NewSigned(claims, jwt.SignParameters{AdditionalHeaders: headers}, signer)
	if err != nil {
		return "", fmt.Errorf("sign token failed: %w", err)
	}

	tokenBytes, err := token.Serialize(false)
	if err != nil {
		return "", fmt.Errorf("serialize token failed: %w", err)
	}

	return tokenBytes, nil
}

func (e *preAuthTokenResponse) expiry() time.Time {
	if v := e.ExpiresIn; v != 0 {
		return time.Now().Add(time.Duration(v) * time.Second)
	}

	return time.Time{}
}
