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
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/component/kmscrypto/doc/jose"
	"github.com/hyperledger/aries-framework-go/component/models/jwt"
	"github.com/hyperledger/aries-framework-go/component/models/verifiable"
	"github.com/piprate/json-gold/ld"
	"golang.org/x/oauth2"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	metadatafetcher "github.com/trustbloc/wallet-sdk/pkg/internal/issuermetadata"
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
	requestCredentialEventText          = "Request credential(s) from issuer"
	fetchOpenIDConfigViaGETReqEventText = "Fetch issuer's OpenID configuration via an HTTP GET request to %s"
	//nolint:gosec //false positive
	fetchTokenViaPOSTReqEventText = "Fetch token via an HTTP POST request to %s"
	//nolint:gosec //false positive
	fetchCredentialViaGETReqEventText  = "Fetch credential %d of %d via an HTTP POST request to %s"
	parseAndCheckProofCheckVCEventText = "Parsing and checking proof for received credential %d of %d"

	preAuthorizedGrantType     = "urn:ietf:params:oauth:grant-type:pre-authorized_code"
	authorizationCodeGrantType = "authorization_code"
)

// Interaction represents a single OpenID4CI interaction between a wallet and an issuer.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// An Interaction is a stateful object, and is intended for going through the full flow only once
// after which it should be discarded. Any new interactions should use a fresh Interaction instance.
type Interaction struct {
	issuerURI                    string
	credentialTypes              [][]string
	credentialFormats            []string
	clientID                     string
	didResolver                  *didResolverWrapper
	activityLogger               api.ActivityLogger
	metricsLogger                api.MetricsLogger
	disableVCProofChecks         bool
	documentLoader               ld.DocumentLoader
	issuerMetadata               *issuer.Metadata
	preAuthorizedCodeGrantParams *PreAuthorizedCodeGrantParams
	authorizationCodeGrantParams *AuthorizationCodeGrantParams
	openIDConfig                 *OpenIDConfig
	oAuth2Config                 *oauth2.Config
	authTokenResponse            *oauth2.Token
	httpClient                   *http.Client
	authCodeURLState             string
	codeVerifier                 string
}

// NewInteraction creates a new OpenID4CI Interaction.
// If no ActivityLogger is provided (via the ClientConfig object), then no activity logging will take place.
func NewInteraction(initiateIssuanceURI string, config *ClientConfig) (*Interaction, error) {
	timeStartNewInteraction := time.Now()

	err := validateRequiredParameters(config)
	if err != nil {
		return nil, err
	}

	setDefaults(config)

	credentialOffer, err := getCredentialOffer(initiateIssuanceURI, config.HTTPClient, config.MetricsLogger)
	if err != nil {
		return nil, err
	}

	// TODO Add support for determining grant types when no grants are specified.
	// See https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html#section-4.1.1 for more info.
	preAuthorizedCodeGrantParams, authorizationCodeGrantParams, err := determineIssuerGrantCapabilities(credentialOffer)
	if err != nil {
		return nil, err
	}

	credentialTypes, credentialFormats, err := determineCredentialTypesAndFormats(credentialOffer)
	if err != nil {
		return nil, err
	}

	return &Interaction{
			preAuthorizedCodeGrantParams: preAuthorizedCodeGrantParams,
			authorizationCodeGrantParams: authorizationCodeGrantParams,
			issuerURI:                    credentialOffer.CredentialIssuer,
			credentialTypes:              credentialTypes,
			credentialFormats:            credentialFormats,
			didResolver:                  &didResolverWrapper{didResolver: config.DIDResolver},
			activityLogger:               config.ActivityLogger,
			metricsLogger:                config.MetricsLogger,
			disableVCProofChecks:         config.DisableVCProofChecks,
			documentLoader:               config.DocumentLoader,
			httpClient:                   config.HTTPClient,
		}, config.MetricsLogger.Log(&api.MetricsEvent{
			Event:    newInteractionEventText,
			Duration: time.Since(timeStartNewInteraction),
		})
}

// CreateAuthorizationURL creates an authorization URL that can be opened in a browser to proceed to the login page.
// It is the first step in the authorization code flow.
// It creates the authorization URL that can be opened in a browser to proceed to the login page.
// This method can only be used if the issuer supports authorization code grants.
// Check the issuer's capabilities first using the Capabilities method.
// If scopes are needed, pass them in using the WithScopes option.
func (i *Interaction) CreateAuthorizationURL(clientID, redirectURI string,
	opts ...CreateAuthorizationURLOpt,
) (string, error) {
	if !i.AuthorizationCodeGrantTypeSupported() {
		return "", errors.New("issuer does not support the authorization code grant type")
	}

	processedOpts := processCreateAuthorizationURLOpts(opts)

	var err error

	i.issuerMetadata, err = metadatafetcher.Get(i.issuerURI, i.httpClient, i.metricsLogger,
		"Authorization")
	if err != nil {
		return "", walleterror.NewExecutionError(
			module,
			MetadataFetchFailedCode,
			MetadataFetchFailedError,
			fmt.Errorf("failed to get issuer metadata: %w", err))
	}

	i.instantiateOAuth2Config(clientID, redirectURI, processedOpts.scopes)

	err = i.instantiateCodeVerifier()
	if err != nil {
		return "", err
	}

	authorizationDetails, err := i.generateAuthorizationDetails()
	if err != nil {
		return "", err
	}

	authCodeOptions := i.generateAuthCodeOptions(authorizationDetails)

	i.authCodeURLState = uuid.New().String()

	i.clientID = clientID

	return i.oAuth2Config.AuthCodeURL(i.authCodeURLState, authCodeOptions...), nil
}

// RequestCredentialWithPreAuth requests credential(s) from the issuer. This method can only be used for the
// pre-authorized code flow, where it acts as the final step in the interaction with the issuer.
// For the equivalent method for the authorization code flow, see RequestCredentialWithAuth instead.
// If a PIN is required (which can be checked via the Capabilities method), then it must be passed
// into this method via the WithPIN option.
func (i *Interaction) RequestCredentialWithPreAuth(jwtSigner api.JWTSigner, opts ...RequestCredentialWithPreAuthOpt,
) ([]*verifiable.Credential, error) {
	processedOpts := processRequestCredentialWithPreAuthOpts(opts)

	return i.requestCredential(jwtSigner, processedOpts.pin)
}

// RequestCredentialWithAuth requests credential(s) from the issuer. This method can only be used for the
// authorization code flow, where it acts as the final step in the interaction with the issuer.
// For the equivalent method for the pre-authorized code flow, see RequestCredentialWithPreAuth instead.
//
// RequestCredentialWithAuth should be called only once all authorization pre-requisite steps have been completed.
// The redirect URI that you pass in here should look like the redirect URI that you passed in to the
// CreateAuthorizationURL, except that now it has some URL query parameters appended to it.
func (i *Interaction) RequestCredentialWithAuth(jwtSigner api.JWTSigner, redirectURIWithParams string,
) ([]*verifiable.Credential, error) {
	err := i.requestAccessToken(redirectURIWithParams)
	if err != nil {
		return nil, err
	}

	return i.requestCredential(jwtSigner, "")
}

// IssuerURI returns the issuer's URI from the initiation request. It's useful to store this somewhere in case
// there's a later need to refresh credential display data using the latest display information from the issuer.
func (i *Interaction) IssuerURI() string {
	return i.issuerURI
}

// PreAuthorizedCodeGrantTypeSupported indicates whether the issuer supports the pre-authorized code grant type.
func (i *Interaction) PreAuthorizedCodeGrantTypeSupported() bool {
	return i.preAuthorizedCodeGrantParams != nil
}

// PreAuthorizedCodeGrantParams returns an object that can be used to determine the issuer's pre-authorized code grant
// parameters. The caller should call the PreAuthorizedCodeGrantTypeSupported method first and only call this method to
// get the params if PreAuthorizedCodeGrantTypeSupported returns true.
// This method returns an error if (and only if) PreAuthorizedCodeGrantTypeSupported returns false.
func (i *Interaction) PreAuthorizedCodeGrantParams() (*PreAuthorizedCodeGrantParams, error) {
	if i.preAuthorizedCodeGrantParams == nil {
		return nil, errors.New("issuer does not support the pre-authorized code grant")
	}

	return i.preAuthorizedCodeGrantParams, nil
}

// AuthorizationCodeGrantTypeSupported indicates whether the issuer supports the authorization code grant type.
func (i *Interaction) AuthorizationCodeGrantTypeSupported() bool {
	return i.authorizationCodeGrantParams != nil
}

// AuthorizationCodeGrantParams returns an object that can be used to determine the issuer's authorization code grant
// parameters. The caller should call the AuthorizationCodeGrantTypeSupported method first and only call this method to
// get the params if AuthorizationCodeGrantTypeSupported returns true.
// This method returns an error if (and only if) AuthorizationCodeGrantTypeSupported returns false.
func (i *Interaction) AuthorizationCodeGrantParams() (*AuthorizationCodeGrantParams, error) {
	if i.authorizationCodeGrantParams == nil {
		return nil, errors.New("issuer does not support the authorization code grant")
	}

	return i.authorizationCodeGrantParams, nil
}

// DynamicClientRegistrationSupported indicates whether the issuer supports dynamic client registration.
func (i *Interaction) DynamicClientRegistrationSupported() (bool, error) {
	var err error

	i.openIDConfig, err = i.getOpenIDConfig()
	if err != nil {
		return false, walleterror.NewExecutionError(
			module,
			IssuerOpenIDConfigFetchFailedCode,
			IssuerOpenIDConfigFetchFailedError,
			fmt.Errorf("failed to fetch issuer's OpenID configuration: %w", err))
	}

	return i.openIDConfig.RegistrationEndpoint != nil, nil
}

// DynamicClientRegistrationEndpoint returns the issuer's dynamic client registration endpoint.
// The caller should call the DynamicClientRegistrationSupported method first and only call this method
// if DynamicClientRegistrationSupported returns true.
// This method will return an error if the issuer does not support dynamic client registration.
func (i *Interaction) DynamicClientRegistrationEndpoint() (string, error) {
	var err error

	i.openIDConfig, err = i.getOpenIDConfig()
	if err != nil {
		return "", walleterror.NewExecutionError(
			module,
			IssuerOpenIDConfigFetchFailedCode,
			IssuerOpenIDConfigFetchFailedError,
			fmt.Errorf("failed to fetch issuer's OpenID configuration: %w", err))
	}

	if i.openIDConfig.RegistrationEndpoint == nil {
		return "", errors.New("issuer does not support dynamic client registration")
	}

	return *i.openIDConfig.RegistrationEndpoint, nil
}

func (i *Interaction) requestAccessToken(redirectURIWithAuthCode string) error {
	if i.oAuth2Config == nil {
		return errors.New("authorization URL must be created first")
	}

	parsedURI, err := url.Parse(redirectURIWithAuthCode)
	if err != nil {
		return err
	}

	exists := parsedURI.Query().Has("code")
	if !exists {
		return errors.New("redirect URI is missing an authorization code")
	}

	exists = parsedURI.Query().Has("state")
	if !exists {
		return errors.New("redirect URI is missing a state value")
	}

	state := parsedURI.Query().Get("state")
	if state != i.authCodeURLState {
		return walleterror.NewExecutionError(
			module,
			StateInRedirectURINotMatchingAuthURLCode,
			StateInRedirectURINotMatchingAuthURLError,
			errors.New("state in redirect URI does not match the state from the authorization URL"))
	}

	i.openIDConfig, err = i.getOpenIDConfig()
	if err != nil {
		return walleterror.NewExecutionError(
			module,
			IssuerOpenIDConfigFetchFailedCode,
			IssuerOpenIDConfigFetchFailedError,
			fmt.Errorf("failed to fetch issuer's OpenID configuration: %w", err))
	}

	i.oAuth2Config.Endpoint.TokenURL = i.openIDConfig.TokenEndpoint

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, i.httpClient)

	i.authTokenResponse, err = i.oAuth2Config.Exchange(ctx, parsedURI.Query().Get("code"),
		oauth2.SetAuthURLParam("code_verifier", i.codeVerifier))

	return err
}

func (i *Interaction) instantiateOAuth2Config(clientID, redirectURI string, scopes []string) {
	i.oAuth2Config = &oauth2.Config{
		ClientID: clientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:   i.issuerMetadata.AuthorizationServer,
			AuthStyle: oauth2.AuthStyleInHeader,
		},
		RedirectURL: redirectURI,
	}

	if len(scopes) != 0 {
		i.oAuth2Config.Scopes = scopes
	}
}

func (i *Interaction) instantiateCodeVerifier() error {
	const randomBytesToGenerate = 32
	randomBytes := make([]byte, randomBytesToGenerate)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return err
	}

	i.codeVerifier = base64.RawURLEncoding.EncodeToString(randomBytes)

	return nil
}

func (i *Interaction) generateCodeChallenge() string {
	codeVerifierHash := sha256.Sum256([]byte(i.codeVerifier))

	codeChallenge := base64.RawURLEncoding.EncodeToString(codeVerifierHash[:])

	return codeChallenge
}

func (i *Interaction) generateAuthorizationDetails() ([]byte, error) {
	// TODO: Add support for requesting multiple credentials at once (by sending an array).
	// Currently we always use the first credential type specified in the offer.
	authorizationDetails := &authorizationDetails{
		Type:   "openid_credential",
		Types:  i.credentialTypes[0],
		Format: i.credentialFormats[0],
	}

	if i.issuerMetadata.AuthorizationServer != "" {
		authorizationDetails.Locations = []string{i.issuerMetadata.CredentialIssuer}
	}

	authorizationDetailsBytes, err := json.Marshal(authorizationDetails)
	if err != nil {
		return nil, err
	}

	return authorizationDetailsBytes, nil
}

func (i *Interaction) generateAuthCodeOptions(authorizationDetails []byte) []oauth2.AuthCodeOption {
	authCodeOptions := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", i.generateCodeChallenge()),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("authorization_details", string(authorizationDetails)),
	}

	if i.authorizationCodeGrantParams.IssuerState != nil {
		authCodeOptions = append(authCodeOptions, oauth2.SetAuthURLParam("issuer_state",
			*i.authorizationCodeGrantParams.IssuerState))
	}

	return authCodeOptions
}

func (i *Interaction) requestCredential(jwtSigner api.JWTSigner, //nolint:funlen
	pin string,
) ([]*verifiable.Credential, error) {
	timeStartRequestCredential := time.Now()

	err := validateSignerKeyID(jwtSigner)
	if err != nil {
		return nil, err
	}

	grantType, err := i.determineGrantTypeToUse(pin)
	if err != nil {
		return nil, err
	}

	var credentialResponses []CredentialResponse

	if grantType == preAuthorizedGrantType {
		credentialResponses, err = i.getCredentialResponsesUsingPreAuth(pin, jwtSigner)
	} else {
		credentialResponses, err = i.getCredentialResponsesUsingAuth(jwtSigner)
	}

	if err != nil {
		return nil,
			walleterror.NewExecutionError(
				module,
				CredentialFetchFailedCode,
				CredentialFetchFailedError,
				fmt.Errorf("failed to get credential response: %w", err))
	}

	vcs, err := i.getVCsFromCredentialResponses(credentialResponses)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			module,
			CredentialParseFailedCode,
			CredentialParseError, err)
	}

	subjectIDs, err := getSubjectIDs(vcs)
	if err != nil {
		return nil, err
	}

	err = i.metricsLogger.Log(&api.MetricsEvent{
		Event:    requestCredentialEventText,
		Duration: time.Since(timeStartRequestCredential),
	})
	if err != nil {
		return nil, err
	}

	return vcs, i.activityLogger.Log(&api.Activity{
		ID:   uuid.New(),
		Type: api.LogTypeCredentialActivity,
		Time: time.Now(),
		Data: api.Data{
			Client:    i.issuerMetadata.CredentialIssuer,
			Operation: activityLogOperation,
			Status:    api.ActivityLogStatusSuccess,
			Params:    map[string]interface{}{"subjectIDs": subjectIDs},
		},
	})
}

// Based on the current state of this Interaction so far, as well as the PIN passed in by the caller,
// this method determines which grant type flow we're "in" and should use for the RequestCredential method.
// This method returns an error if it detects that the Interaction object is in an invalid state for
// requesting a credential.
func (i *Interaction) determineGrantTypeToUse( //nolint: gocyclo // Difficult to decompose nicely
	pin string,
) (string, error) {
	if i.preAuthorizedCodeGrantParams == nil && i.authorizationCodeGrantParams == nil {
		return "", errors.New("interaction not instantiated")
	}

	if i.onlyPreAuthorizedCodeGrantSupported() {
		if i.preAuthorizedCodeGrantParams.PINRequired() && pin == "" {
			return "", walleterror.NewValidationError(
				module,
				PINRequiredCode,
				PINRequiredError,
				errors.New("the credential offer requires a user PIN, but none was provided"))
		}

		return preAuthorizedGrantType, nil
	}

	if i.onlyAuthorizationCodeGrantSupported() {
		if i.authTokenResponse == nil {
			return "", errors.New("issuer requires authorization before credential issuance. " +
				"Complete authorization steps first")
		}

		return authorizationCodeGrantType, nil
	}

	// Reaching this point means that the issuer supports both authorization grant types.
	// In this case, the caller needs to have met at least one of the two grant type's parameter requirements.
	// Whichever one is met determines the grant type we'll use.
	// If, for some reason, the caller has met both grant type requirements simultaneously, then we just use the
	// pre-authorized code grant type.

	preAuthCodeRequirementsMet := !i.preAuthorizedCodeGrantParams.PINRequired() || pin != ""
	if preAuthCodeRequirementsMet {
		return preAuthorizedGrantType, nil
	}

	authCodeRequirementsMet := i.authTokenResponse != nil
	if !authCodeRequirementsMet {
		return "", errors.New("authorization requirements not met. " +
			"Either a PIN must be provided or the authorization code grant steps must be completed first")
	}

	return authorizationCodeGrantType, nil
}

func (i *Interaction) onlyPreAuthorizedCodeGrantSupported() bool {
	return i.PreAuthorizedCodeGrantTypeSupported() && !i.AuthorizationCodeGrantTypeSupported()
}

func (i *Interaction) onlyAuthorizationCodeGrantSupported() bool {
	return i.AuthorizationCodeGrantTypeSupported() && !i.PreAuthorizedCodeGrantTypeSupported()
}

func (i *Interaction) getCredentialResponsesUsingPreAuth(pin string, //nolint:funlen // Difficult to decompose
	signer api.JWTSigner,
) ([]CredentialResponse, error) {
	var err error

	i.openIDConfig, err = i.getOpenIDConfig()
	if err != nil {
		return nil, walleterror.NewExecutionError(
			module,
			IssuerOpenIDConfigFetchFailedCode,
			IssuerOpenIDConfigFetchFailedError,
			fmt.Errorf("failed to fetch issuer's OpenID configuration: %w", err))
	}

	tokenResponse, err := i.getPreAuthTokenResponse(pin)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			module,
			TokenFetchFailedCode,
			TokenFetchFailedError,
			fmt.Errorf("failed to get token response: %w", err))
	}

	proofJWT, err := i.createClaimsProof(tokenResponse.CNonce, signer)
	if err != nil {
		return nil, err
	}

	i.issuerMetadata, err = metadatafetcher.Get(i.issuerURI, i.httpClient, i.metricsLogger,
		requestCredentialEventText)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			module,
			MetadataFetchFailedCode,
			MetadataFetchFailedError,
			fmt.Errorf("failed to get issuer metadata: %w", err))
	}

	credentialResponses := make([]CredentialResponse, len(i.credentialTypes))

	for index := range i.credentialTypes {
		request, err := i.createCredentialRequestWithoutAccessToken(proofJWT, index)
		if err != nil {
			return nil, err
		}

		request.Header.Add("Authorization", "Bearer "+tokenResponse.AccessToken)

		fetchCredentialResponseEventText := fmt.Sprintf(fetchCredentialViaGETReqEventText, index+1,
			len(i.credentialTypes), i.issuerMetadata.CredentialEndpoint)

		responseBytes, err := i.getRawCredentialResponse(request, fetchCredentialResponseEventText, i.httpClient)
		if err != nil {
			return nil, err
		}

		var credentialResponse CredentialResponse

		err = json.Unmarshal(responseBytes, &credentialResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response from the issuer's credential endpoint: %w", err)
		}

		credentialResponses[index] = credentialResponse
	}

	return credentialResponses, nil
}

func (i *Interaction) createClaimsProof(nonce interface{}, signer api.JWTSigner) (string, error) {
	claims := map[string]interface{}{
		"aud":   i.issuerURI,
		"iat":   time.Now().Unix(),
		"nonce": nonce,
	}

	if i.clientID != "" {
		claims["iss"] = i.clientID // Only used in the authorization code flow.
	}

	proofJWT, err := signToken(claims, signer)
	if err != nil {
		return "", walleterror.NewExecutionError(
			module,
			JWTSigningFailedCode,
			JWTSigningFailedError,
			fmt.Errorf("failed to create JWT: %w", err))
	}

	return proofJWT, nil
}

func (i *Interaction) getCredentialResponsesUsingAuth(signer api.JWTSigner) ([]CredentialResponse, error) {
	proofJWT, err := i.createClaimsProof(i.authTokenResponse.Extra("c_nonce"), signer)
	if err != nil {
		return nil, err
	}

	credentialResponses := make([]CredentialResponse, len(i.credentialTypes))

	oAuthHTTPClient := i.createOAuthHTTPClient()

	for index := range i.credentialTypes {
		request, err := i.createCredentialRequestWithoutAccessToken(proofJWT, index)
		if err != nil {
			return nil, err
		}

		// The access token header will be injected automatically by the OAuth HTTP client, so there's no need to
		// explicitly set it on the request object generated by the method call above.

		fetchCredentialResponseEventText := fmt.Sprintf(fetchCredentialViaGETReqEventText, index+1,
			len(i.credentialTypes), i.issuerMetadata.CredentialEndpoint)

		responseBytes, err := i.getRawCredentialResponse(request, fetchCredentialResponseEventText, oAuthHTTPClient)
		if err != nil {
			return nil, err
		}

		var credentialResponse CredentialResponse

		err = json.Unmarshal(responseBytes, &credentialResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal response from the issuer's credential endpoint: %w", err)
		}

		credentialResponses[index] = credentialResponse
	}

	return credentialResponses, nil
}

// getOpenIDConfig fetches the OpenID configuration from the issuer. If the OpenID configuration has already been
// fetched before, then it's returned without making an additional call.
func (i *Interaction) getOpenIDConfig() (*OpenIDConfig, error) {
	if i.openIDConfig != nil {
		return i.openIDConfig, nil
	}

	openIDConfigEndpoint := i.issuerURI + "/.well-known/openid-configuration"

	responseBytes, err := httprequest.New(i.httpClient, i.metricsLogger).Do(
		http.MethodGet, openIDConfigEndpoint, "", nil,
		fmt.Sprintf(fetchOpenIDConfigViaGETReqEventText, openIDConfigEndpoint), requestCredentialEventText)
	if err != nil {
		return nil, fmt.Errorf("openid configuration endpoint: %w", err)
	}

	var config OpenIDConfig

	err = json.Unmarshal(responseBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from the issuer's "+
			"OpenID configuration endpoint: %w", err)
	}

	return &config, nil
}

func (i *Interaction) getRawCredentialResponse(credentialReq *http.Request, eventText string, httpClient *http.Client,
) ([]byte, error) {
	timeStartHTTPRequest := time.Now()

	response, err := httpClient.Do(credentialReq)
	if err != nil {
		return nil, err
	}

	err = i.metricsLogger.Log(&api.MetricsEvent{
		Event:       eventText,
		ParentEvent: requestCredentialEventText,
		Duration:    time.Since(timeStartHTTPRequest),
	})
	if err != nil {
		return nil, err
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code [%d] with body [%s] from issuer's credential endpoint",
			response.StatusCode, string(responseBytes))
	}

	defer func() {
		errClose := response.Body.Close()
		if errClose != nil {
			println(fmt.Sprintf("failed to close response body: %s", errClose.Error()))
		}
	}()

	return responseBytes, nil
}

// The returned *http.Request will not have the access token set on it. The caller must ensure that it's set
// before sending the request to the server.
func (i *Interaction) createCredentialRequestWithoutAccessToken(proofJWT string, credentialFormatAndTypesIndex int,
) (*http.Request, error) {
	credentialReq := &credentialRequest{
		Types:  i.credentialTypes[credentialFormatAndTypesIndex],
		Format: i.credentialFormats[credentialFormatAndTypesIndex],
		Proof: proof{
			ProofType: "jwt", // TODO: https://github.com/trustbloc/wallet-sdk/issues/159 support other proof types
			JWT:       proofJWT,
		},
	}

	credentialReqBytes, err := json.Marshal(credentialReq)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, //nolint: noctx
		i.issuerMetadata.CredentialEndpoint, bytes.NewReader(credentialReqBytes))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")

	return request, nil
}

func (i *Interaction) getVCsFromCredentialResponses(
	credentialResponses []CredentialResponse,
) ([]*verifiable.Credential, error) {
	var vcs []*verifiable.Credential

	vdrKeyResolver := verifiable.NewVDRKeyResolver(i.didResolver)

	credentialOpts := []verifiable.CredentialOpt{
		verifiable.WithJSONLDDocumentLoader(i.documentLoader),
		verifiable.WithPublicKeyFetcher(vdrKeyResolver.PublicKeyFetcher()),
	}

	if i.disableVCProofChecks {
		credentialOpts = append(credentialOpts, verifiable.WithDisabledProofCheck())
	}

	for j := range credentialResponses {
		timeStartParseCredential := time.Now()

		credentialResponseBytes, err := credentialResponses[j].SerializeToCredentialsBytes()
		if err != nil {
			return nil, fmt.Errorf("failed to parse credential from credential response at index %d: %w", j, err)
		}

		vc, err := verifiable.ParseCredential(credentialResponseBytes, credentialOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse credential from credential response at index %d: %w", j, err)
		}

		err = i.metricsLogger.Log(&api.MetricsEvent{
			Event:       fmt.Sprintf(parseAndCheckProofCheckVCEventText, j+1, len(credentialResponses)),
			ParentEvent: requestCredentialEventText,
			Duration:    time.Since(timeStartParseCredential),
		})
		if err != nil {
			return nil, err
		}

		vcs = append(vcs, vc)
	}

	return vcs, nil
}

func (i *Interaction) getPreAuthTokenResponse(pin string) (*preAuthTokenResponse, error) {
	params := url.Values{}
	params.Add("grant_type", preAuthorizedGrantType)
	params.Add("pre-authorized_code", i.preAuthorizedCodeGrantParams.preAuthorizedCode)

	if pin != "" {
		params.Add("user_pin", pin)
	}

	paramsReader := strings.NewReader(params.Encode())

	responseBytes, err := httprequest.New(i.httpClient, i.metricsLogger).Do(
		http.MethodPost, i.openIDConfig.TokenEndpoint, "application/x-www-form-urlencoded", paramsReader,
		fmt.Sprintf(fetchTokenViaPOSTReqEventText, i.openIDConfig.TokenEndpoint), requestCredentialEventText)
	if err != nil {
		return nil, fmt.Errorf("issuer's token endpoint: %w", err)
	}

	var tokenResp preAuthTokenResponse

	err = json.Unmarshal(responseBytes, &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from the issuer's token endpoint: %w", err)
	}

	return &tokenResp, nil
}

// createOAuthHTTPClient creates the OAuth2 client wrapper using the OAuth2 library.
// Due to some peculiarities with the OAuth2 library, we need to do some things here to ensure our custom HTTP client
// settings get preserved. Check the comments in the method below for more details.
func (i *Interaction) createOAuthHTTPClient() *http.Client {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, i.httpClient)

	// The HTTP client below only retains the Transport, so we have to set the timeout again.
	// The docs say that the returned client shouldn't be modified, but there doesn't seem to be a clear reason
	// why not - at least for the timeout setting. See https://github.com/golang/oauth2/issues/368 for more info.
	// Also, the docs on NewClient in the OAuth2 library state that the transport from the client in the context above
	// won't be used for non-token-fetching requests, but this seems to be inaccurate (as of writing).
	// Any additional headers injected in by the headerInjectionRoundTripper should be set as expected.
	// See https://github.com/golang/oauth2/issues/324 for more info.
	oAuthHTTPClient := i.oAuth2Config.Client(ctx, i.authTokenResponse)
	oAuthHTTPClient.Timeout = i.httpClient.Timeout

	return oAuthHTTPClient
}

func getCredentialOffer(initiateIssuanceURI string, httpClient *http.Client, metricsLogger api.MetricsLogger,
) (*CredentialOffer, error) {
	requestURIParsed, err := url.Parse(initiateIssuanceURI)
	if err != nil {
		return nil, walleterror.NewValidationError(
			module,
			InvalidIssuanceURICode,
			InvalidIssuanceURIError,
			err)
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
				module,
				InvalidIssuanceURICode,
				InvalidIssuanceURIError,
				errors.New("credential offer query parameter missing from initiate issuance URI"))
	}

	var credentialOffer CredentialOffer

	err = json.Unmarshal(credentialOfferJSON, &credentialOffer)
	if err != nil {
		return nil, walleterror.NewValidationError(
			module,
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
		fmt.Sprintf(fetchCredOfferViaGETReqEventText, credentialOfferURI), newInteractionEventText)
	if err != nil {
		return nil, walleterror.NewValidationError(
			module,
			InvalidCredentialOfferCode,
			InvalidCredentialOfferError,
			fmt.Errorf("failed to get credential offer from the endpoint specified in the "+
				"credential_offer_uri URL query parameter: %w", err))
	}

	return responseBytes, nil
}

func determineCredentialTypesAndFormats(credentialOffer *CredentialOffer) ([][]string, []string, error) {
	// TODO Add support for credential offer objects that contain a credentials field with JSON strings instead.
	// See https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html#section-4.1.1 for more info.
	credentialTypes := make([][]string, len(credentialOffer.Credentials))
	credentialFormats := make([]string, len(credentialOffer.Credentials))

	for i := 0; i < len(credentialOffer.Credentials); i++ {
		if credentialOffer.Credentials[i].Format != jwtVCJSONCredentialFormat &&
			credentialOffer.Credentials[i].Format != jwtVCJSONLDCredentialFormat &&
			credentialOffer.Credentials[i].Format != ldpVCCredentialFormat {
			return nil, nil, walleterror.NewValidationError(
				module,
				UnsupportedCredentialTypeInOfferCode,
				UnsupportedCredentialTypeInOfferError,
				fmt.Errorf("unsupported credential type (%s) in credential offer at index %d of "+
					"credentials object (must be jwt_vc_json or jwt_vc_json-ld)",
					credentialOffer.Credentials[i].Format, i))
		}

		credentialTypes[i] = credentialOffer.Credentials[i].Types
		credentialFormats[i] = credentialOffer.Credentials[i].Format
	}

	return credentialTypes, credentialFormats, nil
}

func validateSignerKeyID(jwtSigner api.JWTSigner) error {
	kidParts := strings.Split(jwtSigner.GetKeyID(), "#")
	if len(kidParts) < 2 { //nolint: gomnd
		return walleterror.NewExecutionError(
			module,
			KeyIDNotContainDIDPartCode,
			KeyIDNotContainDIDPartError,
			fmt.Errorf("key ID (%s) is missing the DID part", jwtSigner.GetKeyID()))
	}

	return nil
}

func getSubjectIDs(vcs []*verifiable.Credential) ([]string, error) {
	var subjectIDs []string

	for i := 0; i < len(vcs); i++ {
		subjects, ok := vcs[i].Subject.([]verifiable.Subject)
		if !ok {
			return nil, fmt.Errorf("unexpected VC subject type for credential at index %d", i)
		}

		for j := 0; j < len(subjects); j++ {
			subjectIDs = append(subjectIDs, subjects[j].ID)
		}
	}

	return subjectIDs, nil
}

func signToken(claims interface{}, signer api.JWTSigner) (string, error) {
	headers := jose.Headers{}
	// TODO: Send "typ" header.
	// headers["typ"] = "openid4vci-proof+jwt"

	token, err := jwt.NewSigned(claims, headers, signer)
	if err != nil {
		return "", fmt.Errorf("sign token failed: %w", err)
	}

	tokenBytes, err := token.Serialize(false)
	if err != nil {
		return "", fmt.Errorf("serialize token failed: %w", err)
	}

	return tokenBytes, nil
}
