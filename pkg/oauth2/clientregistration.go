/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package oauth2 provides a client API for doing OAuth2 dynamic client registration per RFC7591:
// https://datatracker.ietf.org/doc/html/rfc7591
package oauth2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
)

const (
	newRegisterClientEventText  = "Register client"
	fetchRequestObjectEventText = "Fetch request object via an HTTP GET request to %s"
)

// RegisterClient registers a new client at the given registration endpoint.
// If the server requires an initial access token, then use the WithInitialAccessBearerToken option.
func RegisterClient(registrationEndpoint string, clientMetadata *ClientMetadata,
	opts ...Opt,
) (*RegisterClientResponse, error) {
	if registrationEndpoint == "" {
		return nil, errors.New("registration endpoint cannot be blank")
	}

	// Technically, all fields are optional, so if the caller passes in nil then we will assume that they want to
	// send an empty object to the server
	if clientMetadata == nil {
		clientMetadata = &ClientMetadata{}
	}

	processedOpts := processOpts(opts)

	clientMetadataBytes, err := json.Marshal(clientMetadata)
	if err != nil {
		return nil, err
	}

	respBody, err := getRawResponse(clientMetadataBytes, registrationEndpoint, processedOpts)
	if err != nil {
		return nil, err
	}

	var response *RegisterClientResponse

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body into a RegisterClientResponse: %w", err)
	}

	return response, nil
}

func getRawResponse(requestBytes []byte, registrationEndpoint string, opts *opts) ([]byte, error) {
	headers := http.Header{}
	if opts.initialAccessBearerToken != "" {
		headers.Set("Authorization", "Bearer "+opts.initialAccessBearerToken)
	}

	metricsEvent := fmt.Sprintf(fetchRequestObjectEventText, registrationEndpoint)

	return httprequest.New(opts.httpClient, opts.metricsLogger).DoContext(context.TODO(),
		http.MethodPost, registrationEndpoint, "application/json", headers,
		bytes.NewReader(requestBytes), metricsEvent, newRegisterClientEventText,
		[]int{http.StatusCreated}, nil)
}
