/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialschema contains a function that can be used to resolve display values per the OpenID4CI spec.
// This implementation follows the 27 October 2022 revision of
// https://openid.net/specs/openid-4-verifiable-credential-issuance-1_0.html#section-11.2
package credentialschema

// Resolve resolves display information for some issued credentials based on an issuer's metadata.
// The CredentialDisplays in the returned ResolvedDisplayData object correspond to the VCs passed in and are in the
// same order.
// This method requires one VC source and one issuer metadata source. See opts.go for more information.
func Resolve(opts ...ResolveOpt) (*ResolvedDisplayData, error) {
	vcs, metadata, preferredLocale, err := processOpts(opts)
	if err != nil {
		return nil, err
	}

	credentialDisplays, err := buildCredentialDisplays(vcs, metadata.CredentialsSupported, preferredLocale)
	if err != nil {
		return nil, err
	}

	issuerOverview := getIssuerDisplay(metadata.CredentialIssuer, preferredLocale)

	return &ResolvedDisplayData{
		IssuerDisplay:      issuerOverview,
		CredentialDisplays: credentialDisplays,
	}, nil
}
