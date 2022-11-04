/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package openid4vp contains functionality for doing OpenID4VP operations.
package openid4vp

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/linkeddomains"
)

// Instance is used to help with OpenID4VP operations (TODO: Implement).
type Instance struct {
	authorizationRequest []byte
	keyHandleReader      api.KeyReader
	didResolver          api.DIDResolver
	linkedDomains        *linkeddomains.LD
}

// NewInstance returns a new OpenID4VP Instance.
func NewInstance(authorizationRequest []byte, keyHandleReader api.KeyReader,
	didResolver api.DIDResolver, linkedDomains *linkeddomains.LD,
) (*Instance, error) {
	return &Instance{
		authorizationRequest: authorizationRequest,
		keyHandleReader:      keyHandleReader,
		didResolver:          didResolver,
		linkedDomains:        linkedDomains,
	}, nil
}

// GetQuery does something (TODO: Implement).
func (o *Instance) GetQuery() ([]byte, error) {
	return []byte("Example"), nil
}

// PresentCredential does something (TODO: Implement).
func (o *Instance) PresentCredential(presentation []byte, kid string) error {
	return nil
}
