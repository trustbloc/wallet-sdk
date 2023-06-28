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

// HasClientIDIssuedAt indicates whether this RegisterClientResponse specifies when the client ID was issued.
func (r *RegisterClientResponse) HasClientIDIssuedAt() bool {
	return r.goAPIRegisterClientResponse.ClientIDIssuedAt != nil
}

// ClientIDIssuedAt returns the time at which the client ID was issued.
// The HasClientIDIssuedAt method should be called first to determine whether this RegisterClientResponse
// specifies when the client ID was issued before calling this method.
// This method returns an error if (and only if) HasClientIDIssuedAt returns false.
func (r *RegisterClientResponse) ClientIDIssuedAt() (int, error) {
	if !r.HasClientIDIssuedAt() {
		return -1, errors.New("the register client response object does not specify when the client ID was issued")
	}

	return *r.goAPIRegisterClientResponse.ClientIDIssuedAt, nil
}

// HasClientSecretExpiresAt indicates whether this RegisterClientResponse specifies when the client secret expires.
func (r *RegisterClientResponse) HasClientSecretExpiresAt() bool {
	return r.goAPIRegisterClientResponse.ClientSecretExpiresAt != nil
}

// ClientSecretExpiresAt returns the time at which the client secret will expire or 0 if it will not expire.
// The HasClientSecretExpiresAt method should be called first to determine whether this RegisterClientResponse
// specifies when the client secret will expire before calling this method.
// This method returns an error if (and only if) HasClientSecretExpiresAt returns false.
func (r *RegisterClientResponse) ClientSecretExpiresAt() (int, error) {
	if !r.HasClientSecretExpiresAt() {
		return -1, errors.New("the register client response object does not " +
			"specify when the client secret expires")
	}

	return *r.goAPIRegisterClientResponse.ClientSecretExpiresAt, nil
}

// HasClientMetadata indicates whether this RegisterClientResponse has client metadata.
func (r *RegisterClientResponse) HasClientMetadata() bool {
	return r.goAPIRegisterClientResponse.ClientMetadata != nil
}

// ClientMetadata returns the ClientMetadata object. The HasClientMetadata method should be called first to
// ensure this RegisterClientResponse object has any client metadata first before calling this method.
// This method returns an error if (and only if) HasClientMetadata returns false.
func (r *RegisterClientResponse) ClientMetadata() (*ClientMetadata, error) {
	if !r.HasClientMetadata() {
		return nil, errors.New("the register client response object has no client metadata")
	}

	return &ClientMetadata{goAPIClientMetadata: r.goAPIRegisterClientResponse.ClientMetadata}, nil
}
