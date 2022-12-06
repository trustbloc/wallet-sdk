/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oidc4vp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/trustbloc/wallet-sdk/test/integration/pkg/httprequest"
	"github.com/trustbloc/wallet-sdk/test/integration/pkg/oauth"
)

const (
	oidcProviderURL                    = "https://localhost:4444"
	credentialServiceURL               = "https://localhost:4455"
	verifierProfileURL                 = credentialServiceURL + "/verifier/profiles"
	verifierProfileURLFormat           = verifierProfileURL + "/%s"
	initiateOidcInteractionURLFormat   = verifierProfileURLFormat + "/interactions/initiate-oidc"
	retrieveInteractionsClaimURLFormat = credentialServiceURL + "/verifier/interactions/%s/claim"
)

type initiateOIDC4VPResponse struct {
	AuthorizationRequest string `json:"authorizationRequest"`
	TxId                 string `json:"txID"`
}

type Setup struct {
	verifierAccessToken string
	httpRequest         *httprequest.Request
}

func NewSetup(httpRequest *httprequest.Request) *Setup {
	return &Setup{
		httpRequest: httpRequest,
	}
}

func (e *Setup) AuthorizeVerifier(orgID string) error {
	accessToken, err := oauth.IssueAccessToken(context.Background(), oidcProviderURL,
		orgID, "test-org-secret", []string{"org_admin"})
	if err != nil {
		return err
	}

	e.verifierAccessToken = accessToken

	return nil
}

func (e *Setup) InitiateInteraction(profileName string) (string, error) {
	endpointURL := fmt.Sprintf(initiateOidcInteractionURLFormat, profileName)

	result := &initiateOIDC4VPResponse{}

	_, err := e.httpRequest.Send(http.MethodPost, endpointURL, "application/json", e.verifierAccessToken, //nolint: bodyclose
		nil, &result)
	if err != nil {
		return "", err
	}

	return result.AuthorizationRequest, nil
}

func (e *Setup) RetrieveInteractionsClaim(txID, authToken string) error {
	endpointURL := fmt.Sprintf(retrieveInteractionsClaimURLFormat, txID)
	_, err := e.httpRequest.Send(http.MethodGet, endpointURL, "application/json", authToken, //nolint: bodyclose
		nil, nil)
	if err != nil {
		return err
	}

	return nil
}
