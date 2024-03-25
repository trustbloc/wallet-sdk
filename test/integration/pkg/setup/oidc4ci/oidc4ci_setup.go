/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oidc4ci

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/trustbloc/wallet-sdk/test/integration/pkg/httprequest"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/oauth"
)

const (
	offerCredentialURLFormat = "%s" + "/issuer/profiles/%s/v1.0/interactions/initiate-oidc"
	vcsAuthorizeEndpoint     = "%s" + "/oidc/authorize"
	vcsTokenEndpoint         = "%s" + "/oidc/token"
	claimDataURL             = "https://mock-login-consent.example.com:8099/claim-data"
	xAPIKey                  = "rw_token"

	oidcProviderURL = "http://localhost:9229/local_5a9GzRvB"
	organizationID  = "f13d1va9lp403pb9lyj89vk55"

	vcsAPIGateway = "https://api-gateway.trustbloc.local:5566"

	preAuthorizedCodeGrantType = "urn:ietf:params:oauth:grant-type:pre-authorized_code"
	authorizationCodeGrantType = "authorization_code"
)

type initiateIssuanceCredentialConfiguration struct {
	ClaimData             map[string]interface{}    `json:"claim_data,omitempty"`
	ClaimEndpoint         string                    `json:"claim_endpoint,omitempty"`
	CredentialTemplateId  string                    `json:"credential_template_id,omitempty"`
	CredentialName        string                    `json:"credential_name,omitempty"`
	CredentialDescription string                    `json:"credential_description,omitempty"`
	CredentialExpiresAt   *time.Time                `json:"credential_expires_at,omitempty"`
	Compose               *composeOIDC4CICredential `json:"compose,omitempty"`
}

type composeOIDC4CICredential struct {
	Credential         *map[string]interface{} `json:"credential,omitempty"`
	IdTemplate         *string                 `json:"id_template"`
	OverrideIssuer     *bool                   `json:"override_issuer"`
	OverrideSubjectDid *bool                   `json:"override_subject_did"`
}

type initiateOIDC4CIRequest struct {
	ClientInitiateIssuanceUrl string   `json:"client_initiate_issuance_url,omitempty"`
	ClientWellknown           string   `json:"client_wellknown,omitempty"`
	GrantType                 string   `json:"grant_type,omitempty"`
	OpState                   string   `json:"op_state,omitempty"`
	ResponseType              string   `json:"response_type,omitempty"`
	Scope                     []string `json:"scope,omitempty"`
	UserPinRequired           bool     `json:"user_pin_required,omitempty"`

	CredentialConfiguration []initiateIssuanceCredentialConfiguration `json:"credential_configuration,omitempty"`
}

type initiateOIDC4CIResponse struct {
	OfferCredentialURL string `json:"offer_credential_URL"`
}

type Setup struct {
	oauthClient       *oauth2.Config // oauthClient is a public client to vcs oidc provider
	cookie            *cookiejar.Jar
	issuerAccessToken string
	organizationID    string
	apiURL            string

	httpRequest        *httprequest.Request
	debug              bool
	offerCredentialURL string
}

func NewSetup(httpRequest *httprequest.Request) (*Setup, error) {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, fmt.Errorf("init cookie jar: %w", err)
	}

	return &Setup{
		cookie:      jar,
		httpRequest: httpRequest,
	}, nil
}

func (s *Setup) AuthorizeIssuerBypassAuth(orgID, VCSAPIDirect string) error {
	s.organizationID = orgID
	s.apiURL = VCSAPIDirect
	return nil
}

func (s *Setup) AuthorizeIssuer(orgID, oidcProviderURL, vcsAPIGateway string) error {
	accessToken, err := oauth.IssueAccessToken(context.Background(), oidcProviderURL,
		orgID, "test-org-secret", []string{"org_admin"})
	if err != nil {
		return err
	}

	s.issuerAccessToken = accessToken
	s.apiURL = vcsAPIGateway

	return s.registerPublicClient()
}

func (s *Setup) registerPublicClient() error {
	// OAuth's clients are imported into vcs from the file (./fixtures/oauth-clients/clients.json)
	s.oauthClient = &oauth2.Config{
		ClientID:    "oidc4vc_client",
		RedirectURL: "https://client.example.com/oauth/redirect",
		Scopes:      []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:   vcsAuthorizeEndpoint,
			TokenURL:  vcsTokenEndpoint,
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}

	return nil
}

type CredentialConfiguration struct {
	ClaimData            map[string]interface{}
	ClaimEndpoint        string `json:"claim_endpoint,omitempty"`
	CredentialTemplateId string
}

func (s *Setup) InitiatePreAuthorizedIssuance(
	issuerProfileID string,
	credentialConfigurations []CredentialConfiguration,
) (string, error) {
	issuanceURL := fmt.Sprintf(offerCredentialURLFormat, s.apiURL, issuerProfileID)

	configuration := make([]initiateIssuanceCredentialConfiguration, 0)

	for _, c := range credentialConfigurations {
		configuration = append(configuration,
			initiateIssuanceCredentialConfiguration{
				ClaimData:            c.ClaimData,
				CredentialTemplateId: c.CredentialTemplateId,
			},
		)
	}

	reqBody, err := json.Marshal(&initiateOIDC4CIRequest{
		CredentialConfiguration: configuration,
		GrantType:               preAuthorizedCodeGrantType,
		Scope:                   []string{"openid", "profile"},
	})
	if err != nil {
		return "", err
	}

	var oidcInitiateResponse initiateOIDC4CIResponse
	_, err = s.httpRequest.Send(http.MethodPost, issuanceURL, "application/json",
		s.getAuthHeaders(), bytes.NewReader(reqBody), &oidcInitiateResponse)
	if err != nil {
		return "", err
	}

	s.offerCredentialURL = oidcInitiateResponse.OfferCredentialURL

	if s.offerCredentialURL == "" {
		return "", fmt.Errorf("offer credential URL is empty")
	}

	return s.offerCredentialURL, nil
}

func (s *Setup) getAuthHeaders() map[string]string {
	headers := map[string]string{}

	if s.issuerAccessToken != "" {
		headers["Authorization"] = "Bearer " + s.issuerAccessToken
	}
	if s.organizationID != "" {
		headers["X-Tenant-ID"] = s.organizationID
		headers["X-API-Key"] = xAPIKey
	}
	return headers
}

func InitiateAuthCodeIssuance() (string, error) {
	accessToken, err := issueAccessToken(context.Background(), oidcProviderURL,
		organizationID, "ejqxi9jb1vew2jbdnogpjcgrz", []string{"org_admin"})
	if err != nil {
		return "", err
	}

	println(accessToken)

	endpointURL := fmt.Sprintf("%s/issuer/profiles/%s/v1.0/interactions/initiate-oidc", vcsAPIGateway,
		"bank_issuer")

	reqBody, err := json.Marshal(&initiateOIDC4CIRequest{
		CredentialConfiguration: []initiateIssuanceCredentialConfiguration{
			{
				CredentialTemplateId: "templateID",
				ClaimEndpoint:        claimDataURL,
			},
		},
		GrantType:    authorizationCodeGrantType,
		OpState:      uuid.New().String(),
		ResponseType: "code",
		Scope:        []string{"openid", "profile"},
	})

	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, endpointURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	}}}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("expected resp status OK but got %d : resp=%s", resp.StatusCode, string(b))
	}

	println("response %s", string(b))
	var r initiateOIDC4CIResponse

	err = json.Unmarshal(b, &r)
	if err != nil {
		return "", err
	}

	return r.OfferCredentialURL, nil
}

func issueAccessToken(ctx context.Context, oidcProviderURL, clientID, secret string, scopes []string) (string, error) {
	conf := clientcredentials.Config{
		TokenURL:     oidcProviderURL + "/oauth2/token",
		ClientID:     clientID,
		ClientSecret: secret,
		Scopes:       scopes,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})

	token, err := conf.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	return token.AccessToken, nil
}
