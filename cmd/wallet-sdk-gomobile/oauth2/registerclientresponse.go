/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oauth2

import (
	"errors"

	goapioauth2 "github.com/trustbloc/wallet-sdk/pkg/oauth2"
)

// RegisterClientResponse represents a response to a new client registration request.
type RegisterClientResponse struct {
	goAPIRegisterClientResponse *goapioauth2.RegisterClientResponse
}

// ClientID returns the client ID.
func (r *RegisterClientResponse) ClientID() string {
	return r.goAPIRegisterClientResponse.ClientID
}

// ClientSecret returns the client secret.
func (r *RegisterClientResponse) ClientSecret() string {
	return r.goAPIRegisterClientResponse.ClientSecret
}

// ClientIDIssuedAt returns the time at which the client ID was issued.
func (r *RegisterClientResponse) ClientIDIssuedAt() int {
	return r.goAPIRegisterClientResponse.ClientIDIssuedAt
}

// ClientSecretExpiresAt returns the time at which the client secret will expire or 0 if it will not expire.
func (r *RegisterClientResponse) ClientSecretExpiresAt() int {
	return r.goAPIRegisterClientResponse.ClientSecretExpiresAt
}

// HasClientMetadata indicates whether this RegisterClientResponse has client metadata.
func (r *RegisterClientResponse) HasClientMetadata() bool {
	return r.goAPIRegisterClientResponse.ClientMetadata != nil
}

// ClientMetadata returns the ClientMetadata object. The HasClientMetadata method should be called first to
// ensure this RegisterClientResponse object has any client metadata first before calling this method.
// If this RegisterClientResponse has no client metadata, then this method returns an error.
func (r *RegisterClientResponse) ClientMetadata() (*ClientMetadata, error) {
	if !r.HasClientMetadata() {
		return nil, errors.New("the register client response object has no client metadata")
	}

	return &ClientMetadata{goAPIClientMetadata: r.goAPIRegisterClientResponse.ClientMetadata}, nil
}
