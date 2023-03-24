/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4ci provides APIs for wallets to receive verifiable credentials via OIDC for Credential Issuance.
package openid4ci

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jwt"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	metadatafetcher "github.com/trustbloc/wallet-sdk/pkg/internal/issuermetadata"
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
)

// Interaction represents a single OpenID4CI interaction between a wallet and an issuer. The methods defined on this
// object are used to help guide the calling code through the OpenID4CI flow.
type Interaction struct {
	issuerURI              string
	credentialTypes        [][]string
	credentialFormats      []string
	preAuthorizedCodeGrant *Grant
	clientID               string
	didResolver            *didResolverWrapper
	activityLogger         api.ActivityLogger
	metricsLogger          api.MetricsLogger
	disableVCProofChecks   bool
	httpClient             httpClient
	documentLoader         ld.DocumentLoader
}

// NewInteraction creates a new OpenID4CI Interaction.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// Calling this function represents taking the first step in the flow.
// This function takes in an Initiate Issuance Request object from an issuer (as defined in
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-5.1), encoded using URL query
// parameters. This object is intended for going through the full flow only once (i.e. one interaction), after which
// it should be discarded. Any new interactions should use a fresh Interaction instance.
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
	preAuthorizedCodeGrant, exists := credentialOffer.Grants["urn:ietf:params:oauth:grant-type:pre-authorized_code"]
	if !exists {
		return nil, walleterror.NewValidationError(
			module,
			PreAuthorizedGrantTypeRequiredCode,
			PreAuthorizedGrantTypeRequiredError,
			errors.New("pre-authorized grant type is required in the credential offer "+
				"(support for other grant types not implemented)"))
	}

	credentialTypes, credentialFormats, err := determineCredentialTypesAndFormats(credentialOffer)
	if err != nil {
		return nil, err
	}

	return &Interaction{
			issuerURI:              credentialOffer.CredentialIssuer,
			credentialTypes:        credentialTypes,
			credentialFormats:      credentialFormats,
			preAuthorizedCodeGrant: &preAuthorizedCodeGrant,
			clientID:               config.ClientID,
			didResolver:            &didResolverWrapper{didResolver: config.DIDResolver},
			activityLogger:         config.ActivityLogger,
			metricsLogger:          config.MetricsLogger,
			disableVCProofChecks:   config.DisableVCProofChecks,
			httpClient:             config.HTTPClient,
			documentLoader:         config.DocumentLoader,
		}, config.MetricsLogger.Log(&api.MetricsEvent{
			Event:    newInteractionEventText,
			Duration: time.Since(timeStartNewInteraction),
		})
}

// Authorize is used by a wallet to authorize an issuer's OIDC Verifiable Credential Issuance Request.
// After initializing the Interaction object with an Issuance Request, this should be the first method you call in
// order to continue with the flow.
// It only supports the pre-authorized flow in its current implementation.
// Once the authorization flow is implemented, the following section of the spec will be relevant:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html#name-authorization-endpoint
func (i *Interaction) Authorize() (*AuthorizeResult, error) {
	if i.preAuthorizedCodeGrant == nil {
		return nil, errors.New("interaction not instantiated")
	}

	authorizeResult := &AuthorizeResult{
		UserPINRequired: i.preAuthorizedCodeGrant.UserPINRequired,
	}

	return authorizeResult, nil
}

// RequestCredential is the final step in the interaction.
// This is called after the wallet is authorized and is ready to receive credential(s).
// Relevant sections of the spec:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0-11.html#name-credential-endpoint
func (i *Interaction) RequestCredential(credentialRequestOpts *CredentialRequestOpts, jwtSigner api.JWTSigner) ([]*verifiable.Credential, error) { //nolint:funlen,gocyclo,lll
	timeStartRequestCredential := time.Now()

	if credentialRequestOpts == nil {
		credentialRequestOpts = &CredentialRequestOpts{}
	}

	if i.preAuthorizedCodeGrant.UserPINRequired && credentialRequestOpts.UserPIN == "" {
		return nil, walleterror.NewValidationError(
			module,
			PINRequiredCode,
			PINRequiredError,
			errors.New("the credential offer requires a user PIN, but it was not provided"))
	}

	config, err := i.fetchIssuerOpenIDConfig()
	if err != nil {
		return nil, walleterror.NewExecutionError(
			module,
			IssuerOpenIDConfigFetchFailedCode,
			IssuerOpenIDConfigFetchFailedError,
			fmt.Errorf("failed to fetch issuer's OpenID configuration: %w", err))
	}

	params := url.Values{}
	params.Add("grant_type", "urn:ietf:params:oauth:grant-type:pre-authorized_code")
	params.Add("pre-authorized_code", i.preAuthorizedCodeGrant.PreAuthorizedCode)
	params.Add("user_pin", credentialRequestOpts.UserPIN)

	tokenResp, err := i.getTokenResponse(config.TokenEndpoint, params)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			module,
			TokenFetchFailedCode,
			TokenFetchFailedError,
			fmt.Errorf("failed to get token response: %w", err))
	}

	claims := map[string]interface{}{
		"iss":   i.clientID,
		"aud":   i.issuerURI,
		"iat":   time.Now().Unix(),
		"nonce": tokenResp.CNonce,
	}

	token, err := signToken(claims, jwtSigner)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			module,
			JWTSigningFailedCode,
			JWTSigningFailedError,
			fmt.Errorf("failed to create JWT: %w", err))
	}

	credentialResponses := make([]CredentialResponse, len(i.credentialTypes))

	kidParts := strings.Split(jwtSigner.GetKeyID(), "#")
	if len(kidParts) < 2 { //nolint: gomnd
		return nil, walleterror.NewExecutionError(
			module,
			KeyIDNotContainDIDPartCode,
			KeyIDNotContainDIDPartError,
			fmt.Errorf("kid not containing did part %s", jwtSigner.GetKeyID()))
	}

	metadata, err := metadatafetcher.Get(i.issuerURI, i.httpClient, i.metricsLogger, requestCredentialEventText)
	if err != nil {
		return nil, walleterror.NewExecutionError(
			module,
			MetadataFetchFailedCode,
			MetadataFetchFailedError,
			fmt.Errorf("failed to get issuer metadata: %w", err))
	}

	for index := range i.credentialTypes {
		credentialResponse, errGetCredResp := i.getCredentialResponse(metadata.CredentialEndpoint,
			tokenResp.AccessToken, token, i.credentialFormats[index], i.credentialTypes[index], index+1, len(i.credentialTypes))
		if errGetCredResp != nil {
			return nil,
				walleterror.NewExecutionError(
					module,
					CredentialFetchFailedCode,
					CredentialFetchFailedError,
					fmt.Errorf("failed to get credential response: %w", errGetCredResp))
		}

		credentialResponses[index] = *credentialResponse
	}

	vcs, err := i.getCredentialsFromResponses(credentialResponses)
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
			Client:    metadata.CredentialIssuer,
			Operation: activityLogOperation,
			Status:    api.ActivityLogStatusSuccess,
			Params:    map[string]interface{}{"subjectIDs": subjectIDs},
		},
	})
}

// IssuerURI returns the issuer's URI from the initiation request. It's useful to store this somewhere in case
// there's a later need to refresh credential display data using the latest display information from the issuer.
func (i *Interaction) IssuerURI() string {
	return i.issuerURI
}

func (i *Interaction) fetchIssuerOpenIDConfig() (*OpenIDConfig, error) {
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

func (i *Interaction) getTokenResponse(tokenEndpointURL string, params url.Values) (*tokenResponse, error) {
	paramsReader := strings.NewReader(params.Encode())

	responseBytes, err := httprequest.New(i.httpClient, i.metricsLogger).Do(
		http.MethodPost, tokenEndpointURL, "application/x-www-form-urlencoded", paramsReader,
		fmt.Sprintf(fetchTokenViaPOSTReqEventText, tokenEndpointURL), requestCredentialEventText)
	if err != nil {
		return nil, fmt.Errorf("issuer's token endpoint: %w", err)
	}

	var tokenResp tokenResponse

	err = json.Unmarshal(responseBytes, &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from the issuer's token endpoint: %w", err)
	}

	return &tokenResp, nil
}

func (i *Interaction) getCredentialResponse(credentialEndpoint, accessToken, tkn, credentialFormat string,
	credentialTypes []string, credNum, maxCredNum int,
) (*CredentialResponse, error) {
	responseBytes, err := i.getRawCredentialResponse(credentialEndpoint, accessToken, tkn, credentialFormat,
		credentialTypes, credNum, maxCredNum)
	if err != nil {
		return nil, err
	}

	var credentialResponse CredentialResponse

	err = json.Unmarshal(responseBytes, &credentialResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from the issuer's credential endpoint: %w", err)
	}

	return &credentialResponse, nil
}

func (i *Interaction) getRawCredentialResponse(credentialEndpoint, accessToken, tkn, credentialFormat string,
	credentialTypes []string, credNum, maxCredNum int,
) ([]byte, error) {
	request, err := i.createCredentialRequest(credentialEndpoint, accessToken, tkn, credentialFormat, credentialTypes)
	if err != nil {
		return nil, err
	}

	timeStartHTTPRequest := time.Now()

	response, err := i.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	err = i.metricsLogger.Log(&api.MetricsEvent{
		Event:       fmt.Sprintf(fetchCredentialViaGETReqEventText, credNum, maxCredNum, credentialEndpoint),
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

func (i *Interaction) createCredentialRequest(credentialEndpoint, accessToken, tkn, credentialFormat string,
	credentialTypes []string,
) (*http.Request, error) {
	credentialReq := &credentialRequest{
		Types:  credentialTypes,
		Format: credentialFormat,
		Proof: proof{
			ProofType: "jwt", // TODO: https://github.com/trustbloc/wallet-sdk/issues/159 support other proof types
			JWT:       tkn,
		},
	}

	credentialReqBytes, err := json.Marshal(credentialReq)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, //nolint: noctx
		credentialEndpoint, bytes.NewReader(credentialReqBytes))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "BEARER "+accessToken)

	return request, nil
}

func (i *Interaction) getCredentialsFromResponses(
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

func getCredentialOffer(initiateIssuanceURI string, httpClient httpClient, metricsLogger api.MetricsLogger,
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
	httpClient httpClient, metricsLogger api.MetricsLogger,
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
	token, err := jwt.NewSigned(claims, nil, signer)
	if err != nil {
		return "", fmt.Errorf("sign token failed: %w", err)
	}

	tokenBytes, err := token.Serialize(false)
	if err != nil {
		return "", fmt.Errorf("serialize token failed: %w", err)
	}

	return tokenBytes, nil
}
