/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// IssuerInitiatedInteractionArgs contains the required parameters for an IssuerInitiatedInteraction.
type IssuerInitiatedInteractionArgs struct {
	initiateIssuanceURI string
	crypto              api.Crypto
	didResolver         api.DIDResolver
}

// NewIssuerInitiatedInteractionArgs creates a new IssuerInitiatedInteractionArgs object. All parameters are mandatory.
func NewIssuerInitiatedInteractionArgs(initiateIssuanceURI string, crypto api.Crypto,
	didResolver api.DIDResolver,
) *IssuerInitiatedInteractionArgs {
	return &IssuerInitiatedInteractionArgs{
		initiateIssuanceURI: initiateIssuanceURI,
		crypto:              crypto,
		didResolver:         didResolver,
	}
}
