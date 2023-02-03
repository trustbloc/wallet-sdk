/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
)

// ResolveDisplay resolves display information for issued credentials based on an issuer's metadata, which is fetched
// using the issuer's (base) URI.
// The CredentialDisplays in the returned DisplayData object correspond to the VCs passed in and are in the
// same order.
// This method requires one or more VCs and the issuer's base URI.
// The display values are resolved per the 27 October 2022 revision of the OpenID4CI spec:
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-11.2
func ResolveDisplay(vcs *api.VerifiableCredentialsArray, issuerURI, preferredLocale string) (*DisplayData, error) {
	opts, err := prepareOpts(vcs, issuerURI, preferredLocale)
	if err != nil {
		return nil, err
	}

	resolvedDisplayData, err := goapicredentialschema.Resolve(opts...)
	if err != nil {
		return nil, err
	}

	return &DisplayData{resolvedDisplayData: resolvedDisplayData}, nil
}
