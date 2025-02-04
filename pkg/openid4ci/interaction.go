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
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/vc-go/dataintegrity"
	"github.com/trustbloc/vc-go/dataintegrity/suite/ecdsa2019"
	"github.com/trustbloc/vc-go/proof/defaults"
	"github.com/trustbloc/vc-go/verifiable"
	"golang.org/x/oauth2"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/common"
	diderrors "github.com/trustbloc/wallet-sdk/pkg/did"
	"github.com/trustbloc/wallet-sdk/pkg/did/wellknown"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	metadatafetcher "github.com/trustbloc/wallet-sdk/pkg/internal/issuermetadata"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

const getIssuerMetadataEventText = "Get issuer metadata"

// IssuerTrustInfo represent issuer trust information.
type IssuerTrustInfo struct {
	DID                  string
	Domain               string
	CredentialsSupported []SupportedCredential
}

type basicTrustInfo struct {
	DID         string
	Domain      string
	DomainValid bool
}

// This is a common object shared by both the IssuerInitiatedInteraction and WalletInitiatedInteraction objects.
type interaction struct {
	issuerURI               string
	clientID                string
	didResolver             api.DIDResolver
	activityLogger          api.ActivityLogger
	metricsLogger           api.MetricsLogger
	disableVCProofChecks    bool
	documentLoader          ld.DocumentLoader
	issuerMetadata          *issuer.Metadata
	oAuth2Config            *oauth2.Config
	authTokenResponseNonce  interface{}
	authToken               *universalAuthToken
	httpClient              *http.Client
	authCodeURLState        string
	codeVerifier            string
	requestedAcknowledgment *requestedAcknowledgment
}

type requestedAcknowledgment struct {
	// TODO: after update to the latest OIDC4CI this variable can be changed to string
	// since notification_id should be the same for given session.
	// spec: https://openid.github.io/OpenID4VCI/openid-4-verifiable-credential-issuance-wg-draft.html#section-8.3-14
	ackIDs []string
}

func (i *interaction) createAuthorizationURL(clientID, redirectURI, format string, types, credentialContext []string,
	issuerState *string, scopes []string, useOAuthDiscoverableClientIDScheme bool,
) (string, error) {
	err := i.populateIssuerMetadata("Authorization")
	if err != nil {
		return "", err
	}

	i.instantiateOAuth2Config(clientID, redirectURI, scopes)

	err = i.instantiateCodeVerifier()
	if err != nil {
		return "", err
	}

	authorizationDetails, err := i.generateAuthorizationDetails(format, types, credentialContext)
	if err != nil {
		return "", err
	}

	authCodeOptions := i.generateAuthCodeOptions(authorizationDetails, issuerState, useOAuthDiscoverableClientIDScheme)

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

func (i *interaction) generateAuthorizationDetails(format string, types, credentialContext []string) ([]byte, error) {
	// TODO: Add support for requesting multiple credentials at once (by sending an array).
	// Currently we always use the first credential type specified in the offer.
	authorizationDetailsDTO := authorizationDetails{
		CredentialConfigurationID: "",
		CredentialDefinition: &issuer.CredentialDefinition{
			Context:           credentialContext,
			CredentialSubject: nil,
			Type:              types,
		},
		Format:    format,
		Locations: nil,
		Type:      "openid_credential",
	}

	if i.issuerMetadata.AuthorizationServer != "" {
		authorizationDetailsDTO.Locations = []string{i.issuerMetadata.CredentialIssuer}
	}

	authorizationDetailsBytes, err := json.Marshal([]authorizationDetails{authorizationDetailsDTO})
	if err != nil {
		return nil, err
	}

	return authorizationDetailsBytes, nil
}

func (i *interaction) generateAuthCodeOptions(authorizationDetails []byte,
	issuerState *string, useOAuthDiscoverableClientIDScheme bool,
) []oauth2.AuthCodeOption {
	authCodeOptions := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", i.generateCodeChallenge()),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("authorization_details", string(authorizationDetails)),
	}

	if issuerState != nil {
		authCodeOptions = append(authCodeOptions, oauth2.SetAuthURLParam("issuer_state", *issuerState))
	}

	if useOAuthDiscoverableClientIDScheme {
		authCodeOptions = append(authCodeOptions, oauth2.SetAuthURLParam("client_id_scheme",
			"urn:ietf:params:oauth:client-id-scheme:oauth-discoverable-client"))
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

	tokenEndpoint, err := i.getTokenEndpoint()
	if err != nil {
		return err
	}

	i.oAuth2Config.Endpoint.TokenURL = tokenEndpoint

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, i.httpClient)

	authTokenResponse, err := i.oAuth2Config.Exchange(ctx, parsedURI.Query().Get("code"),
		oauth2.SetAuthURLParam("code_verifier", i.codeVerifier))
	if err != nil {
		return err
	}

	i.authToken = &universalAuthToken{
		AccessToken:  authTokenResponse.AccessToken,
		TokenType:    authTokenResponse.TokenType,
		ExpiresAt:    authTokenResponse.Expiry,
		RefreshToken: authTokenResponse.RefreshToken,
	}

	i.authTokenResponseNonce = authTokenResponse.Extra("c_nonce")

	return nil
}

func (i *interaction) getTokenEndpoint() (string, error) {
	if i.issuerMetadata.TokenEndpoint != "" {
		return i.issuerMetadata.TokenEndpoint, nil
	}

	return "", walleterror.NewExecutionError(
		ErrorModule,
		NoTokenEndpointAvailableErrorCode,
		NoTokenEndpointAvailableError,
		errors.New("no token endpoint specified in issuer's metadata"))
}

func (i *interaction) dynamicClientRegistrationSupported() (bool, error) {
	err := i.populateIssuerMetadata("Dynamic client registration supported")
	if err != nil {
		return false, err
	}

	return i.issuerMetadata.RegistrationEndpoint != nil, nil
}

func (i *interaction) dynamicClientRegistrationEndpoint() (string, error) {
	err := i.populateIssuerMetadata("Dynamic client registration endpoint")
	if err != nil {
		return "", err
	}

	if i.issuerMetadata.RegistrationEndpoint == nil {
		return "", walleterror.NewInvalidSDKUsageError(ErrorModule,
			errors.New("issuer does not support dynamic client registration"))
	}

	return *i.issuerMetadata.RegistrationEndpoint, nil
}

// If the issuer's metadata has not been fetched before in this interaction's lifespan, then this method fetches the
// issuer's metadata and stores it within this interaction object. If the issuer's metadata has already been fetched
// before, then this method does nothing in order to avoid making an unnecessary GET call.
func (i *interaction) populateIssuerMetadata(parentEvent string) error {
	if i.issuerMetadata == nil {
		jwtVerifier := defaults.NewDefaultProofChecker(common.NewVDRKeyResolver(i.didResolver))

		issuerMetadata, err := metadatafetcher.Get(i.issuerURI, i.httpClient, i.metricsLogger, parentEvent, jwtVerifier)
		if err != nil {
			return walleterror.NewExecutionError(
				ErrorModule,
				MetadataFetchFailedCode,
				MetadataFetchFailedError,
				fmt.Errorf("failed to get issuer metadata: %w", err))
		}

		i.issuerMetadata = issuerMetadata
	}

	return nil
}

func (i *interaction) requestCredentialWithAuth(jwtSigner api.JWTSigner, credentialFormats []string,
	credentialTypes, credentialContexts [][]string,
) ([]*verifiable.Credential, error) {
	timeStartRequestCredential := time.Now()

	credentialResponses, err := i.getCredentialResponsesWithAuth(jwtSigner, credentialFormats, credentialTypes,
		credentialContexts)
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

	subjectIDs := getSubjectIDs(vcs)

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
	credentialTypes, credentialContexts [][]string,
) ([]CredentialResponse, error) {
	return i.getCredentialResponse(signer, i.authTokenResponseNonce,
		credentialFormats, credentialTypes, credentialContexts, true)
}

//nolint:nonamedreturns
func (i *interaction) getCredentialResponse(signer api.JWTSigner, nonce any,
	credentialFormats []string, credentialTypes, credentialContexts [][]string, allowRetry bool,
) (credentialResponse []CredentialResponse, err error) {
	defer func() {
		if err == nil || !allowRetry {
			return
		}

		proofError := &InvalidProofError{}
		if errors.As(err, &proofError) {
			credentialResponse, err = i.getCredentialResponse(signer, proofError.CNonce,
				credentialFormats, credentialTypes, credentialContexts, false)
		}
	}()

	proofJWT, err := i.createClaimsProof(nonce, signer)
	if err != nil {
		return nil, err
	}

	credentialResponses := make([]CredentialResponse, len(credentialTypes))

	oAuthHTTPClient := createOAuthHTTPClient(i.oAuth2Config, i.authToken, i.httpClient)

	for index := range credentialTypes {
		requestBody, err := i.createCredentialRequestBody(proofJWT, credentialFormats[index],
			credentialTypes[index], credentialContexts[index])
		if err != nil {
			return nil, err
		}

		fetchCredentialResponseEventText := fmt.Sprintf(fetchCredentialViaGETReqEventText, index+1,
			len(credentialTypes), i.issuerMetadata.CredentialEndpoint)

		// The access token header will be injected automatically by the OAuth HTTP client, so there's no need to
		// explicitly set it on the request object generated by the method call above.
		responseBytes, err := httprequest.New(oAuthHTTPClient, i.metricsLogger).DoContext(context.TODO(),
			http.MethodPost, i.issuerMetadata.CredentialEndpoint, "application/json", nil,
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

		i.storeAcknowledgmentID(credentialResponse.AckID)
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
func createOAuthHTTPClient(
	oAuth2Config *oauth2.Config, token *universalAuthToken, httpClient *http.Client,
) *http.Client {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, httpClient)

	// The HTTP client below only retains the Transport, so we have to set the timeout again.
	// The docs say that the returned client shouldn't be modified, but there doesn't seem to be a clear reason
	// why not - at least for the timeout setting. See https://github.com/golang/oauth2/issues/368 for more info.
	// Also, the docs on NewClient in the OAuth2 library state that the transport from the client in the context above
	// won't be used for non-token-fetching requests, but this seems to be inaccurate (as of writing).
	// Any additional headers injected in by the headerInjectionRoundTripper should be set as expected.
	// See https://github.com/golang/oauth2/issues/324 for more info.
	oAuthHTTPClient := oAuth2Config.Client(ctx, &oauth2.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.ExpiresAt,
	})
	oAuthHTTPClient.Timeout = httpClient.Timeout

	return oAuthHTTPClient
}

func (i *interaction) createCredentialRequestBody(proofJWT, credentialFormat string,
	credentialTypes, credentialContext []string,
) ([]byte, error) {
	var credentialContextToSend *[]string

	if len(credentialContext) > 0 {
		credentialContextToSend = &credentialContext
	}

	credentialReq := &credentialRequest{
		CredentialDefinition: &credentialDefinition{
			Type:    credentialTypes,
			Context: credentialContextToSend,
		},
		Format: credentialFormat,
		Proof: proof{
			ProofType: "jwt",
			JWT:       proofJWT,
		},
	}

	return json.Marshal(credentialReq)
}

func (i *interaction) getVCsFromCredentialResponses(
	credentialResponses []CredentialResponse,
) ([]*verifiable.Credential, error) {
	var vcs []*verifiable.Credential

	credentialOpts := []verifiable.CredentialOpt{
		verifiable.WithJSONLDDocumentLoader(i.documentLoader),
		verifiable.WithProofChecker(defaults.NewDefaultProofChecker(common.NewVDRKeyResolver(i.didResolver))),
	}

	opts := dataintegrity.Options{DIDResolver: &didResolverWrapper{didResolver: i.didResolver}}

	dataIntegrityVerifier, err := dataintegrity.NewVerifier(&opts,
		ecdsa2019.NewVerifierInitializer(&ecdsa2019.VerifierInitializerOptions{
			LDDocumentLoader: i.documentLoader,
		}))
	if err != nil {
		return nil, err
	}

	credentialOpts = append(credentialOpts, verifiable.WithDataIntegrityVerifier(dataIntegrityVerifier))

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

//nolint:funlen
func processCredentialErrorResponse(statusCode int, respBytes []byte) error {
	detailedErr := fmt.Errorf("received status code [%d] with body [%s] from issuer's credential endpoint",
		statusCode, string(respBytes))

	var errResponse errorResponse

	err := json.Unmarshal(respBytes, &errResponse)
	if err != nil {
		return walleterror.NewExecutionError(ErrorModule,
			OtherCredentialRequestErrorCode,
			OtherCredentialRequestError,
			detailedErr)
	}

	switch errResponse.Error {
	case "invalid_request":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidCredentialRequestErrorCode,
			InvalidCredentialRequestError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	case "invalid_token":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidTokenErrorCode,
			InvalidTokenError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	case "unsupported_credential_format":
		return walleterror.NewExecutionError(ErrorModule,
			UnsupportedCredentialFormatErrorCode,
			UnsupportedCredentialFormatError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	case "unsupported_credential_type":
		return walleterror.NewExecutionError(ErrorModule,
			UnsupportedCredentialTypeErrorCode,
			UnsupportedCredentialTypeError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	case "invalid_or_missing_proof":
		return walleterror.NewExecutionError(ErrorModule,
			InvalidOrMissingProofErrorCode,
			InvalidOrMissingProofError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	case "expired_ack_id":
		return walleterror.NewExecutionError(ErrorModule,
			AcknowledgmentExpiredErrorCode,
			AcknowledgmentExpiredError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	case "invalid_proof":
		return NewInvalidProofError(
			walleterror.NewExecutionError(ErrorModule,
				AcknowledgmentExpiredErrorCode,
				AcknowledgmentExpiredError,
				detailedErr,
				walleterror.WithServerErrorCode(errResponse.Error),
				walleterror.WithServerErrorMessage(errResponse.ErrorDescription),
			),
			errResponse.CNonce, errResponse.CNonceExpiresIn)
	default:
		return walleterror.NewExecutionError(ErrorModule,
			OtherCredentialRequestErrorCode,
			OtherCredentialRequestError,
			detailedErr,
			walleterror.WithServerErrorCode(errResponse.Error),
			walleterror.WithServerErrorMessage(errResponse.ErrorDescription))
	}
}

func (i *interaction) issuerFullTrustInfo(
	credentialTypes [][]string, credentialFormats []string,
) (*IssuerTrustInfo, error) {
	trustInfo, err := i.issuerBasicTrustInfo()
	if err != nil {
		return nil, err
	}

	supportedCredentials := make([]SupportedCredential, len(credentialFormats))

	for j := range credentialFormats {
		supportedCredentials[j] = SupportedCredential{
			Format: credentialFormats[j],
			Types:  credentialTypes[j],
		}
	}

	return &IssuerTrustInfo{
		DID:                  trustInfo.DID,
		Domain:               trustInfo.Domain,
		CredentialsSupported: supportedCredentials,
	}, nil
}

func (i *interaction) issuerBasicTrustInfo() (*basicTrustInfo, error) {
	err := i.populateIssuerMetadata("Verify issuer")
	if err != nil {
		return nil, err
	}

	jwtKID := i.issuerMetadata.GetJWTKID()

	if jwtKID == nil {
		var issuerURI *url.URL

		issuerURI, err = url.Parse(i.issuerURI)
		if err != nil {
			return nil, fmt.Errorf("parse issuer uri: %w", err)
		}

		return &basicTrustInfo{
			Domain: issuerURI.Host,
		}, nil
	}

	jwtKIDSplit := strings.Split(*jwtKID, "#")

	did := jwtKIDSplit[0]

	valid, linkedDomain, err := wellknown.ValidateLinkedDomains(did, i.didResolver, i.httpClient)
	if err != nil {
		return nil, err
	}

	return &basicTrustInfo{
		DID:         did,
		Domain:      linkedDomain,
		DomainValid: valid,
	}, nil
}

func (i *interaction) verifyIssuer() (string, error) {
	trustInfo, err := i.issuerBasicTrustInfo()
	if err != nil {
		return "", err
	}

	if !trustInfo.DomainValid {
		return "", walleterror.NewExecutionError(
			diderrors.Module,
			diderrors.DomainAndDidVerificationCode,
			diderrors.DomainAndDidVerificationFailed,
			fmt.Errorf("DID service validation failed: %w", err))
	}

	return trustInfo.Domain, nil
}

func (i *interaction) requestedAcknowledgmentObj(authToken *universalAuthToken) (*Acknowledgment, error) {
	require, err := i.requireAcknowledgment()
	if err != nil {
		return nil, err
	}

	if !require {
		return nil, fmt.Errorf("issuer not support credential acknowledgement")
	}

	return &Acknowledgment{
		AckIDs:                i.requestedAcknowledgment.ackIDs,
		CredentialAckEndpoint: i.issuerMetadata.NotificationEndpoint,
		IssuerURI:             i.issuerURI,
		AuthToken:             authToken,
	}, nil
}

func (i *interaction) requireAcknowledgment() (bool, error) {
	if i.requestedAcknowledgment == nil {
		return false, fmt.Errorf("no acknowledgment data: request credentials first")
	}

	return i.requestedAcknowledgment != nil && i.issuerMetadata.NotificationEndpoint != "", nil
}

func (i *interaction) storeAcknowledgmentID(ackID string) {
	if i.requestedAcknowledgment == nil {
		i.requestedAcknowledgment = &requestedAcknowledgment{}
	}

	i.requestedAcknowledgment.ackIDs = append(i.requestedAcknowledgment.ackIDs, ackID)
}
