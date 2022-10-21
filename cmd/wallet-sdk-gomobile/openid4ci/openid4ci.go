/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4ci contains functionality for doing OpenID4CI operations.
package openid4ci

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// Instance helps with OpenID4CI operations.
type Instance struct {
	initiateIssuanceRequest []byte
	format                  string
	clientCredentialReader  api.CredentialReader
	keyHandleReader         api.KeyHandleReader
	didResolver             api.DIDResolver
}

// NewInstance returns a new OpenID4CI Instance.
func NewInstance(initiateIssuanceRequest []byte, format string, clientCredentialReader api.CredentialReader,
	keyHandleReader api.KeyHandleReader, didResolver api.DIDResolver,
) *Instance {
	return &Instance{
		initiateIssuanceRequest: initiateIssuanceRequest,
		format:                  format,
		clientCredentialReader:  clientCredentialReader,
		keyHandleReader:         keyHandleReader,
		didResolver:             didResolver,
	}
}

// Authorize does something (TODO: Implement).
func (o *Instance) Authorize(preAuthorizedCode, authorizationRedirectEndpoint string) error {
	return nil
}

// RequestCredential does something (TODO: Implement).
func (o *Instance) RequestCredential(authCode, kid string) ([]byte, error) {
	return []byte("Credential"), nil
}
