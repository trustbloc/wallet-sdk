/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// InteractionArgs contains the required parameters for an Interaction.
type InteractionArgs struct {
	initiateIssuanceURI string
	crypto              api.Crypto
	didResolver         api.DIDResolver
}

// NewInteractionArgs creates a new InteractionArgs object. All parameters are mandatory.
func NewInteractionArgs(initiateIssuanceURI string, crypto api.Crypto, didResolver api.DIDResolver) *InteractionArgs {
	return &InteractionArgs{
		initiateIssuanceURI: initiateIssuanceURI,
		crypto:              crypto,
		didResolver:         didResolver,
	}
}
