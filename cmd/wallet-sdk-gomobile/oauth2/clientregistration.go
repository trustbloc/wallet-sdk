/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package oauth2 provides a client API for doing OAuth2 dynamic client registration per RFC7591:
// https://datatracker.ietf.org/doc/html/rfc7591
package oauth2

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapioauth2 "github.com/trustbloc/wallet-sdk/pkg/oauth2"
)

// RegisterClient registers a new client at the given registration endpoint.
// If the server requires an initial access token, then use RegisterClientWithInitialAccessToken instead.
func RegisterClient(registrationEndpoint string, clientMetadata *ClientMetadata,
	opts *RegisterClientOpts,
) (*RegisterClientResponse, error) {
	// Technically, all fields are optional, so if the caller passes in nil then we will assume that they want to
	// send an empty object to the server
	if clientMetadata == nil {
		clientMetadata = NewClientMetadata()
	}

	if opts == nil {
		opts = NewRegisterClientOpts()
	}

	httpClient := wrapper.NewHTTPClient(opts.httpTimeout, opts.additionalHeaders, opts.disableHTTPClientTLSVerification)

	registerClientResponse, err := goapioauth2.RegisterClient(registrationEndpoint,
		clientMetadata.goAPIClientMetadata,
		goapioauth2.WithInitialAccessBearerToken(opts.initialAccessBearerToken),
		goapioauth2.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	return &RegisterClientResponse{goAPIRegisterClientResponse: registerClientResponse}, nil
}
