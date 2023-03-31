/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// Args contains the required parameters for an Interaction.
type Args struct {
	authorizationRequest string
	crypto               api.Crypto
	didRes               api.DIDResolver
}

// NewArgs creates a new Args object. All parameters are mandatory.
func NewArgs(authorizationRequest string, crypto api.Crypto, didResolver api.DIDResolver) *Args {
	return &Args{
		authorizationRequest: authorizationRequest,
		crypto:               crypto,
		didRes:               didResolver,
	}
}
