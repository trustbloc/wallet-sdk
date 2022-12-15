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
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/aries-framework-go/pkg/doc/util/didsignjwt"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/wallet-sdk/pkg/common"

	"github.com/trustbloc/wallet-sdk/pkg/credentialschema"
	metadatafetcher "github.com/trustbloc/wallet-sdk/pkg/internal/issuermetadata"
)

// NewInteraction creates a new OpenID4CI Interaction.
// The methods defined on this object are used to help guide the calling code through the OpenID4CI flow.
// Calling this function represents taking the first step in the flow.
// This function takes in an Initiate Issuance Request object from an issuer (as defined in
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-5.1), encoded using URL query
// parameters. This object is intended for going through the full flow only once (i.e. one interaction), after which
// it should be discarded. Any new interactions should use a fresh Interaction instance.
func NewInteraction(initiateIssuanceURI string, config *ClientConfig) (*Interaction, error) {
	err := validateClientConfig(config)
	if err != nil {
		return nil, err
	}

	requestURIParsed, err := url.Parse(initiateIssuanceURI)
	if err != nil {
		return nil, err
	}

	initiationRequest := &InitiationRequest{}

	initiationRequest.IssuerURI = requestURIParsed.Query().Get("issuer")
	initiationRequest.CredentialTypes = requestURIParsed.Query()["credential_type"]
	initiationRequest.PreAuthorizedCode = requestURIParsed.Query().Get("pre-authorized_code")

	userPINRequiredString := requestURIParsed.Query().Get("user_pin_required")

	if userPINRequiredString != "" {
		userPINRequired, err := strconv.ParseBool(userPINRequiredString)
		if err != nil {
			return nil, err
		}

		initiationRequest.UserPINRequired = userPINRequired
	}

	initiationRequest.OpState = requestURIParsed.Query().Get("op_state")

	return &Interaction{
		initiationRequest: initiationRequest,
		userDID:           config.UserDID,
		clientID:          config.ClientID,
		signerProvider:    config.SignerProvider,
		didResolver:       &didResolverWrapper{didResolver: config.DIDResolver},
	}, nil
}

// Authorize is used by a wallet to authorize an issuer's OIDC Verifiable Credential Issuance Request.
// After initializing the Interaction object with an Issuance Request, this should be the first method you call in
// order to continue with the flow.
// It only supports the pre-authorized flow in its current implementation.
// Once the authorization flow is implemented, the following section of the spec will be relevant:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-6
func (i *Interaction) Authorize() (*AuthorizeResult, error) {
	if i.initiationRequest.PreAuthorizedCode == "" {
		return nil, errors.New("pre-authorized code is required (authorization flow not implemented)")
	}

	authorizeResult := &AuthorizeResult{
		UserPINRequired: i.initiationRequest.UserPINRequired,
	}

	return authorizeResult, nil
}

// RequestCredential is the second last step (or last step, if the ResolveDisplay method isn't needed) in the
// interaction. This is called after the wallet is authorized and is ready to receive credential(s).
// Relevant sections of the spec:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-7
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-8
func (i *Interaction) RequestCredential(credentialRequestOpts *CredentialRequestOpts) ([]CredentialResponse, error) {
	if i.initiationRequest.UserPINRequired && credentialRequestOpts.UserPIN == "" {
		return nil, errors.New("invalid user PIN")
	}

	metadata, err := metadatafetcher.Get(i.initiationRequest.IssuerURI)
	if err != nil {
		return nil, fmt.Errorf("failed to get issuer metadata: %w", err)
	}

	i.issuerMetadata = metadata

	params := url.Values{}
	params.Add("grant_type", "urn:ietf:params:oauth:grant-type:pre-authorized_code")
	params.Add("pre-authorized_code", i.initiationRequest.PreAuthorizedCode)
	params.Add("user_pin", credentialRequestOpts.UserPIN)

	tokenResp, err := i.getTokenResponse(metadata.TokenEndpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get token response: %w", err)
	}

	claims := map[string]interface{}{
		"iss":   i.clientID,
		"aud":   i.issuerMetadata.Issuer,
		"iat":   time.Now().Unix(),
		"nonce": tokenResp.CNonce,
	}

	// didsignjwt.SignJWT will create the headers automatically
	jwt, err := didsignjwt.SignJWT(nil, claims, i.userDID, i.signerProvider, i.didResolver)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT: %w", err)
	}

	credentialResponses := make([]CredentialResponse, len(i.initiationRequest.CredentialTypes))

	for index, credentialType := range i.initiationRequest.CredentialTypes {
		credentialResponse, err := i.getCredentialResponse(credentialType, metadata.CredentialEndpoint,
			tokenResp.AccessToken, jwt)
		if err != nil {
			return nil, fmt.Errorf("failed to get credential response: %w", err)
		}

		credentialResponses[index] = *credentialResponse
	}

	var vcs []string

	for _, credentialResponse := range credentialResponses {
		vcs = append(vcs, credentialResponse.Credential)
	}

	i.vcs = vcs

	return credentialResponses, nil
}

// ResolveDisplay is the optional final step that can be called after RequestCredential. It resolves display
// information for the credentials received in this interaction. The CredentialDisplays in the returned
// credentialschema.ResolvedDisplayData object correspond to the VCs received and are in the same order.
// If preferredLocale is not specified, then the first locale specified by the issuer's metadata will be used during
// resolution.
func (i *Interaction) ResolveDisplay(preferredLocale string) (*credentialschema.ResolvedDisplayData, error) {
	var credentials []*verifiable.Credential

	for _, vc := range i.vcs {
		credential, err := verifiable.ParseCredential([]byte(vc),
			verifiable.WithJSONLDDocumentLoader(ld.NewDefaultDocumentLoader(common.DefaultHTTPClient())),
			verifiable.WithDisabledProofCheck())
		if err != nil {
			return nil, err
		}

		credentials = append(credentials, credential)
	}

	return credentialschema.Resolve(
		credentialschema.WithCredentials(credentials),
		credentialschema.WithIssuerMetadata(i.issuerMetadata),
		credentialschema.WithPreferredLocale(preferredLocale))
}

func (i *Interaction) getTokenResponse(tokenEndpointURL string, params url.Values) (*tokenResponse, error) {
	// TODO: Implement trusted list type of mechanism? The gosec warning (correctly) warns about using a variable URL.
	response, err := http.Post(tokenEndpointURL, //nolint: noctx,gosec // TODO: To be re-evaluated later
		"application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code [%d] with body [%s] from issuer's token endpoint",
			response.StatusCode, string(responseBytes))
	}

	defer func() {
		errClose := response.Body.Close()
		if errClose != nil {
			println(fmt.Sprintf("failed to close response body: %s", errClose.Error()))
		}
	}()

	var tokenResp tokenResponse

	err = json.Unmarshal(responseBytes, &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from the issuer's token endpoint: %w", err)
	}

	return &tokenResp, nil
}

func (i *Interaction) getCredentialResponse(credentialType, credentialEndpoint,
	accessToken string, jwt string,
) (*CredentialResponse, error) {
	credentialReq := &credentialRequest{
		Type:   credentialType,
		Format: "jwt_vc",
		DID:    i.userDID,
		Proof: proof{
			ProofType: "jwt", // TODO: support other proof types
			JWT:       jwt,
		},
	}

	credentialReqBytes, err := json.Marshal(credentialReq)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, //nolint: noctx // TODO: To be re-evaluated later
		credentialEndpoint, bytes.NewReader(credentialReqBytes))
	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "BEARER "+accessToken)

	response, err := common.DefaultHTTPClient().Do(request)
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

	var credentialResponse CredentialResponse

	err = json.Unmarshal(responseBytes, &credentialResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from the issuer's credential endpoint: %w", err)
	}

	return &credentialResponse, nil
}
