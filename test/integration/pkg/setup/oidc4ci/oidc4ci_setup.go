/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oidc4ci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/trustbloc/wallet-sdk/test/integration/pkg/httprequest"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/oauth"
)

const (
	vcsAPIGateway                       = "https://localhost:4455"
	VCSAPIDirect                        = "http://localhost:8075"
	initiateCredentialIssuanceURLFormat = "%s" + "/issuer/profiles/%s/interactions/initiate-oidc"
	vcsAuthorizeEndpoint                = "%s" + "/oidc/authorize"
	vcsTokenEndpoint                    = "%s" + "/oidc/token"
	oidcProviderURL                     = "https://localhost:4444"
	claimDataURL                        = "https://mock-login-consent.example.com:8099/claim-data"
	xAPIKey                             = "rw_token"
)

type initiateOIDC4CIRequest struct {
	ClaimData                 *map[string]interface{} `json:"claim_data,omitempty"`
	ClaimEndpoint             string                  `json:"claim_endpoint,omitempty"`
	ClientInitiateIssuanceUrl string                  `json:"client_initiate_issuance_url,omitempty"`
	ClientWellknown           string                  `json:"client_wellknown,omitempty"`
	CredentialTemplateId      string                  `json:"credential_template_id,omitempty"`
	GrantType                 string                  `json:"grant_type,omitempty"`
	OpState                   string                  `json:"op_state,omitempty"`
	ResponseType              string                  `json:"response_type,omitempty"`
	Scope                     []string                `json:"scope,omitempty"`
	UserPinRequired           *bool                   `json:"user_pin_required,omitempty"`
}

type initiateOIDC4CIResponse struct {
	InitiateIssuanceUrl string `json:"initiate_issuance_url"`
	TxId                string `json:"tx_id"`
}

type Setup struct {
	oauthClient       *oauth2.Config // oauthClient is a public client to vcs oidc provider
	cookie            *cookiejar.Jar
	issuerAccessToken string
	organizationID    string
	apiURL            string

	httpRequest         *httprequest.Request
	debug               bool
	initiateIssuanceURL string
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

func (s *Setup) AuthorizeIssuerBypassAuth(orgID string) error {
	s.organizationID = orgID
	s.apiURL = VCSAPIDirect
	return nil
}

func (s *Setup) AuthorizeIssuer(orgID string) error {
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

func (s *Setup) InitiateCredentialIssuance(issuerProfileID string) (string, error) {
	endpointURL := fmt.Sprintf(initiateCredentialIssuanceURLFormat, s.apiURL, issuerProfileID)

	reqBody, err := json.Marshal(&initiateOIDC4CIRequest{
		CredentialTemplateId: "templateID",
		GrantType:            "authorization_code",
		OpState:              uuid.New().String(),
		ResponseType:         "code",
		Scope:                []string{"openid", "profile"},
		ClaimEndpoint:        claimDataURL,
	})
	if err != nil {
		return "", fmt.Errorf("marshal initiate oidc4ci req: %w", err)
	}

	var oidc44CIResponse initiateOIDC4CIResponse

	_, err = s.httpRequest.Send(http.MethodPost,
		endpointURL, "application/json", s.getAuthHeaders(), bytes.NewReader(reqBody), &oidc44CIResponse)
	if err != nil {
		return "", fmt.Errorf("https do: %w", err)
	}

	s.initiateIssuanceURL = oidc44CIResponse.InitiateIssuanceUrl

	if s.initiateIssuanceURL == "" {
		return "", fmt.Errorf("initiate issuance URL is empty")
	}

	return s.initiateIssuanceURL, nil
}

func (s *Setup) InitiatePreAuthorizedIssuance(issuerProfileID string) (string, error) {
	issuanceURL := fmt.Sprintf(initiateCredentialIssuanceURLFormat, s.apiURL, issuerProfileID)

	var claimData map[string]interface{}

	if issuerProfileID == "drivers_license_issuer" {
		claimData = map[string]interface{}{
			"birthdate":            "1990-01-01",
			"document_number":      "123-456-789",
			"driving_privileges":   "G2",
			"expiry_date":          "2025-05-26",
			"family_name":          "Smith",
			"given_name":           "John",
			"issue_date":           "2020-05-27",
			"issuing_authority":    "Ministry of Transport Ontario",
			"issuing_country":      "Canada",
			"resident_address":     "4726 Pine Street",
			"resident_city":        "Toronto",
			"resident_postal_code": "A1B 2C3",
			"resident_province":    "Ontario",
		}
	} else {
		claimData = map[string]interface{}{
			"displayName":       "John Doe",
			"givenName":         "John",
			"jobTitle":          "Software Developer",
			"surname":           "Doe",
			"preferredLanguage": "English",
			"mail":              "john.doe@foo.bar",
			"photo":             "data-URL-encoded image",
		}
	}

	reqBody, err := json.Marshal(&initiateOIDC4CIRequest{
		ClaimData:            &claimData,
		CredentialTemplateId: "templateID",
		GrantType:            "authorization_code",
		Scope:                []string{"openid", "profile"},
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

	s.initiateIssuanceURL = oidcInitiateResponse.InitiateIssuanceUrl

	if s.initiateIssuanceURL == "" {
		return "", fmt.Errorf("initiate issuance URL is empty")
	}

	return s.initiateIssuanceURL, nil
}

func (s *Setup) getAuthHeaders() map[string]string {
	headers := map[string]string{}

	if s.issuerAccessToken != "" {
		headers["Authorization"] = "Bearer " + s.issuerAccessToken
	}
	if s.organizationID != "" {
		headers["X-User"] = s.organizationID
		headers["X-API-Key"] = xAPIKey
	}
	return headers
}
