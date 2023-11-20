/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oidc4vp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/trustbloc/wallet-sdk/test/integration/pkg/httprequest"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/oauth"
)

const (
	verifierProfileURL                 = "%s" + "/verifier/profiles"
	verifierProfileURLFormat           = verifierProfileURL + "/%s"
	initiateOidcInteractionURLFormat   = verifierProfileURLFormat + "/v1.0/interactions/initiate-oidc"
	retrieveInteractionsClaimURLFormat = "%s" + "/verifier/interactions/%s/claim"
	xAPIKey                            = "rw_token"
)

type initiateOIDC4VPResponse struct {
	AuthorizationRequest string `json:"authorizationRequest"`
	TxId                 string `json:"txID"`
}

type Setup struct {
	verifierAccessToken string
	organizationID      string
	apiURL              string
	httpRequest         *httprequest.Request
}

type initiateOIDC4VPData struct {
	Purpose *string   `json:"purpose,omitempty"`
	Scopes  *[]string `json:"scopes"`
}

func NewSetup(httpRequest *httprequest.Request) *Setup {
	return &Setup{
		httpRequest: httpRequest,
	}
}

func (s *Setup) AuthorizeVerifierBypassAuth(orgID, vcsAPIDirect string) error {
	s.organizationID = orgID
	s.apiURL = vcsAPIDirect
	return nil
}

func (e *Setup) AuthorizeVerifier(orgID, oidcProviderURL, vcsAPIGateway string) error {
	accessToken, err := oauth.IssueAccessToken(context.Background(), oidcProviderURL,
		orgID, "test-org-secret", []string{"org_admin"})
	if err != nil {
		return err
	}

	e.verifierAccessToken = accessToken
	e.apiURL = vcsAPIGateway

	return nil
}

func (e *Setup) InitiateInteraction(profileName, purpose string, customScopes []string) (string, error) {
	endpointURL := fmt.Sprintf(initiateOidcInteractionURLFormat, e.apiURL, profileName)

	if customScopes == nil {
		customScopes = []string{}
	}

	reqBody, err := json.Marshal(&initiateOIDC4VPData{
		Purpose: &purpose,
		Scopes:  &customScopes,
	})
	if err != nil {
		return "", fmt.Errorf("marshal initiate oidc4vp req: %w", err)
	}

	result := &initiateOIDC4VPResponse{}

	_, err = e.httpRequest.Send(http.MethodPost, endpointURL, "application/json", e.getAuthHeaders(), //nolint: bodyclose
		bytes.NewReader(reqBody), &result)
	if err != nil {
		return "", err
	}

	return result.AuthorizationRequest, nil
}

func (e *Setup) RetrieveInteractionsClaim(txID string) error {
	endpointURL := fmt.Sprintf(retrieveInteractionsClaimURLFormat, e.apiURL, txID)
	_, err := e.httpRequest.Send(http.MethodGet, endpointURL, "application/json", e.getAuthHeaders(), //nolint: bodyclose
		nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (e *Setup) getAuthHeaders() map[string]string {
	headers := map[string]string{}

	if e.verifierAccessToken != "" {
		headers["Authorization"] = "Bearer " + e.verifierAccessToken
	}
	if e.organizationID != "" {
		headers["X-Tenant-ID"] = e.organizationID
		headers["X-API-Key"] = xAPIKey
	}
	return headers
}
