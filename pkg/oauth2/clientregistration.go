/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package oauth2 provides a client API for doing OAuth2 dynamic client registration per RFC7591:
// https://datatracker.ietf.org/doc/html/rfc7591
package oauth2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	httpReq, err := http.NewRequest( //nolint: noctx // Timeout expected to be set in HTTP client already
		http.MethodPost, registrationEndpoint, bytes.NewReader(requestBytes))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	if opts.initialAccessBearerToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+opts.initialAccessBearerToken)
	}

	resp, err := opts.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	defer func() {
		errClose := resp.Body.Close()
		if errClose != nil {
			println(fmt.Sprintf("failed to close response body: %s", errClose.Error()))
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("server returned status code %d with body [%s]", resp.StatusCode,
			string(respBody))
	}

	return respBody, nil
}
