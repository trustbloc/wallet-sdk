/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// Args contains the required parameters for an Interaction.
type Args struct {
	initiateIssuanceURI string
	crypto              api.Crypto
	didResolver         api.DIDResolver
}

// NewArgs creates a new Args object. All parameters are mandatory.
func NewArgs(initiateIssuanceURI string, crypto api.Crypto, didResolver api.DIDResolver) *Args {
	return &Args{
		initiateIssuanceURI: initiateIssuanceURI,
		crypto:              crypto,
		didResolver:         didResolver,
	}
}
