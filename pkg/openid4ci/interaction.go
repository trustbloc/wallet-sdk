/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

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
	"time"

	"github.com/hyperledger/aries-framework-go/component/models/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"

	"github.com/google/uuid"
	"github.com/trustbloc/wallet-sdk/pkg/api"
	metadatafetcher "github.com/trustbloc/wallet-sdk/pkg/internal/issuermetadata"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"

	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
	"golang.org/x/oauth2"
)

// This is a common object shared by both the IssuerInitiatedInteraction and WalletInitiatedInteraction objects.
type interaction struct {
	issuerURI            string
	clientID             string
	didResolver          *didResolverWrapper
	activityLogger       api.ActivityLogger
	metricsLogger        api.MetricsLogger
	disableVCProofChecks bool
	documentLoader       ld.DocumentLoader
	issuerMetadata       *issuer.Metadata
	openIDConfig         *OpenIDConfig
	oAuth2Config         *oauth2.Config
	authTokenResponse    *oauth2.Token
	httpClient           *http.Client
	authCodeURLState     string
	codeVerifier         string
}

// createAuthorizationURL creates an authorization URL that can be opened in a browser to proceed to the login page.
// It is the first step in the authorization code flow.
// It creates the authorization URL that can be opened in a browser to proceed to the login page.
// This method can only be used if the issuer supports authorization code grants.
// Check the issuer's capabilities first using the Capabilities method.
// If scopes are needed, pass them in using the WithScopes option.
func (i *interaction) createAuthorizationURL(clientID, redirectURI, format string, types []string, issuerState *string,
	scopes []string,
) (string, error) {
	err := i.populateIssuerMetadata()
	if err != nil {
		return "", err
	}

	i.instantiateOAuth2Config(clientID, redirectURI, scopes)

	err = i.instantiateCodeVerifier()
	if err != nil {
		return "", err
	}

	authorizationDetails, err := i.generateAuthorizationDetails(format, types)
	if err != nil {
		return "", err
	}

	authCodeOptions := i.generateAuthCodeOptions(authorizationDetails, issuerState)

	i.authCodeURLState = uuid.New().String()

	i.clientID = clientID

	return i.oAuth2Config.AuthCodeURL(i.authCodeURLState, authCodeOptions...), nil
}

func (i *interaction) instantiateOAuth2Config(clientID, redirectURI string, scopes []string) {
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

func (i *interaction) instantiateCodeVerifier() error {
	const randomBytesToGenerate = 32
	randomBytes := make([]byte, randomBytesToGenerate)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return err
	}

	i.codeVerifier = base64.RawURLEncoding.EncodeToString(randomBytes)

	return nil
}

func (i *interaction) generateAuthorizationDetails(format string, types []string) ([]byte, error) {
	// TODO: Add support for requesting multiple credentials at once (by sending an array).
	// Currently we always use the first credential type specified in the offer.
	authorizationDetails := &authorizationDetails{
		Type:   "openid_credential",
		Types:  types,
		Format: format,
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

func (i *interaction) generateAuthCodeOptions(authorizationDetails []byte,
	issuerState *string,
) []oauth2.AuthCodeOption {
	authCodeOptions := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", i.generateCodeChallenge()),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("authorization_details", string(authorizationDetails)),
	}

	if issuerState != nil {
		authCodeOptions = append(authCodeOptions, oauth2.SetAuthURLParam("issuer_state", *issuerState))
	}

	return authCodeOptions
}

func (i *interaction) generateCodeChallenge() string {
	codeVerifierHash := sha256.Sum256([]byte(i.codeVerifier))

	codeChallenge := base64.RawURLEncoding.EncodeToString(codeVerifierHash[:])

	return codeChallenge
}

func (i *interaction) requestAccessToken(redirectURIWithAuthCode string) error {
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
			ErrorModule,
			StateInRedirectURINotMatchingAuthURLCode,
			StateInRedirectURINotMatchingAuthURLError,
			errors.New("state in redirect URI does not match the state from the authorization URL"))
	}

	i.openIDConfig, err = i.getOpenIDConfig()
	if err != nil {
		return walleterror.NewExecutionError(
			ErrorModule,
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

func (i *interaction) dynamicClientRegistrationSupported() (bool, error) {
	var err error

	i.openIDConfig, err = i.getOpenIDConfig()
	if err != nil {
		return false, walleterror.NewExecutionError(
			ErrorModule,
			IssuerOpenIDConfigFetchFailedCode,
			IssuerOpenIDConfigFetchFailedError,
			fmt.Errorf("failed to fetch issuer's OpenID configuration: %w", err))
	}

	return i.openIDConfig.RegistrationEndpoint != nil, nil
}

func (i *interaction) dynamicClientRegistrationEndpoint() (string, error) {
	var err error

	i.openIDConfig, err = i.getOpenIDConfig()
	if err != nil {
		return "", walleterror.NewExecutionError(
			ErrorModule,
			IssuerOpenIDConfigFetchFailedCode,
			IssuerOpenIDConfigFetchFailedError,
			fmt.Errorf("failed to fetch issuer's OpenID configuration: %w", err))
	}

	if i.openIDConfig.RegistrationEndpoint == nil {
		return "", walleterror.NewInvalidSDKUsageError(ErrorModule,
			errors.New("issuer does not support dynamic client registration"))
	}

	return *i.openIDConfig.RegistrationEndpoint, nil
}

// getOpenIDConfig fetches the OpenID configuration from the issuer. If the OpenID configuration has already been
// fetched before, then it's returned without making an additional call.
func (i *interaction) getOpenIDConfig() (*OpenIDConfig, error) {
	if i.openIDConfig != nil {
		return i.openIDConfig, nil
	}

	openIDConfigEndpoint := i.issuerURI + "/.well-known/openid-configuration"

	responseBytes, err := httprequest.New(i.httpClient, i.metricsLogger).Do(
		http.MethodGet, openIDConfigEndpoint, "", nil,
		fmt.Sprintf(fetchOpenIDConfigViaGETReqEventText, openIDConfigEndpoint), requestCredentialEventText, nil)
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

// getIssuerMetadata returns the issuer's metadata. If the issuer's metadata has already been fetched before,
// then it's returned without making an additional call.
func (i *interaction) getIssuerMetadata() (*issuer.Metadata, error) {
	if i.issuerMetadata == nil {
		err := i.populateIssuerMetadata()
		if err != nil {
			return nil, err
		}
	}

	return i.issuerMetadata, nil
}

// populateIssuerMetadata fetches the issuer's metadata and stores it within this interaction object.
func (i *interaction) populateIssuerMetadata() error {
	issuerMetadata, err := metadatafetcher.Get(i.issuerURI, i.httpClient, i.metricsLogger,
		"Authorization")
	if err != nil {
		return walleterror.NewExecutionError(
			ErrorModule,
			MetadataFetchFailedCode,
			MetadataFetchFailedError,
			fmt.Errorf("failed to get issuer metadata: %w", err))
	}

	i.issuerMetadata = issuerMetadata

	return nil
}

func (i *interaction) requestCredentialWithAuth(jwtSigner api.JWTSigner, credentialFormats []string,
	credentialTypes [][]string,
) ([]*verifiable.Credential, error) {
	timeStartRequestCredential := time.Now()

	credentialResponses, err := i.getCredentialResponsesWithAuth(jwtSigner, credentialFormats, credentialTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential response: %w", err)
	}

	vcs, err := i.getVCsFromCredentialResponses(credentialResponses)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			ErrorModule,
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

// credentialsFormats and credentialTypes need to have the same length.
func (i *interaction) getCredentialResponsesWithAuth(signer api.JWTSigner, credentialFormats []string,
	credentialTypes [][]string,
) ([]CredentialResponse, error) {
	proofJWT, err := i.createClaimsProof(i.authTokenResponse.Extra("c_nonce"), signer)
	if err != nil {
		return nil, err
	}

	credentialResponses := make([]CredentialResponse, len(credentialTypes))

	oAuthHTTPClient := i.createOAuthHTTPClient()

	for index := range credentialTypes {
		request, err := i.createCredentialRequestWithoutAccessToken(proofJWT, credentialFormats[index],
			credentialTypes[index])
		if err != nil {
			return nil, err
		}

		// The access token header will be injected automatically by the OAuth HTTP client, so there's no need to
		// explicitly set it on the request object generated by the method call above.

		fetchCredentialResponseEventText := fmt.Sprintf(fetchCredentialViaGETReqEventText, index+1,
			len(credentialTypes), i.issuerMetadata.CredentialEndpoint)

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

func (i *interaction) createClaimsProof(nonce interface{}, signer api.JWTSigner) (string, error) {
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
			ErrorModule,
			JWTSigningFailedCode,
			JWTSigningFailedError,
			fmt.Errorf("failed to create JWT: %w", err))
	}

	return proofJWT, nil
}

// createOAuthHTTPClient creates the OAuth2 client wrapper using the OAuth2 library.
// Due to some peculiarities with the OAuth2 library, we need to do some things here to ensure our custom HTTP client
// settings get preserved. Check the comments in the method below for more details.
func (i *interaction) createOAuthHTTPClient() *http.Client {
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

// The returned *http.Request will not have the access token set on it. The caller must ensure that it's set
// before sending the request to the server.
func (i *interaction) createCredentialRequestWithoutAccessToken(proofJWT, credentialFormat string,
	credentialTypes []string,
) (*http.Request, error) {
	credentialReq := &credentialRequest{
		Types:  credentialTypes,
		Format: credentialFormat,
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

func (i *interaction) getRawCredentialResponse(credentialReq *http.Request, eventText string, httpClient *http.Client,
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
		return nil, processCredentialErrorResponse(response.StatusCode, responseBytes)
	}

	defer func() {
		errClose := response.Body.Close()
		if errClose != nil {
			println(fmt.Sprintf("failed to close response body: %s", errClose.Error()))
		}
	}()

	return responseBytes, nil
}

func (i *interaction) getVCsFromCredentialResponses(
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

func processCredentialErrorResponse(statusCode int, respBytes []byte) error {
	detailedErr := fmt.Errorf("received status code [%d] with body [%s] from issuer's credential endpoint",
		statusCode, string(respBytes))

	var errorResponse errorResponse

	err := json.Unmarshal(respBytes, &errorResponse)
	if err != nil {
		return walleterror.NewExecutionError(ErrorModule,
			OtherCredentialRequestErrorCode,
			OtherCredentialRequestError,
			detailedErr)
	}

	switch errorResponse.Error {
	case "invalid_request":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidCredentialRequestErrorCode,
			InvalidCredentialRequestError,
			detailedErr)
	case "invalid_token":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidTokenErrorCode,
			InvalidTokenError,
			detailedErr)
	case "unsupported_credential_format":
		return walleterror.NewExecutionError(ErrorModule,
			UnsupportedCredentialFormatErrorCode,
			UnsupportedCredentialFormatError,
			detailedErr)
	case "unsupported_credential_type":
		return walleterror.NewExecutionError(ErrorModule,
			UnsupportedCredentialTypeErrorCode,
			UnsupportedCredentialTypeError,
			detailedErr)
	case "invalid_or_missing_proof":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidOrMissingProofErrorCode,
			InvalidOrMissingProofError,
			detailedErr)
	default:
		return walleterror.NewExecutionError(ErrorModule,
			OtherCredentialRequestErrorCode,
			OtherCredentialRequestError,
			detailedErr)
	}
}
