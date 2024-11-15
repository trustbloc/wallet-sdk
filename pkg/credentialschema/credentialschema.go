/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialschema contains a function that can be used to resolve display values per the OpenID4CI spec.
package credentialschema

import (
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

// Resolve resolves display information for some issued credentials based on an issuer's metadata.
// The CredentialDisplays in the returned ResolvedDisplayData object correspond to the VCs passed in and are in the
// same order.
// This method requires one VC source and one issuer metadata source. See opts.go for more information.
func Resolve(opts ...ResolveOpt) (*ResolvedDisplayData, error) {
	credentialConfigMappings, issuerMetadata, preferredLocale, maskingString, err := processOpts(opts)
	if err != nil {
		return nil, err
	}

	if maskingString == nil {
		defaultMaskingString := "•"
		maskingString = &defaultMaskingString
	}

	credentialDisplays, err := buildCredentialDisplays(credentialConfigMappings,
		preferredLocale, *maskingString)
	if err != nil {
		return nil, err
	}

	issuerOverview := getIssuerDisplay(issuerMetadata.LocalizedIssuerDisplays, preferredLocale)

	return &ResolvedDisplayData{
		IssuerDisplay:      issuerOverview,
		CredentialDisplays: credentialDisplays,
	}, nil
}

// ResolveCredential resolves display information for some issued credentials based on an issuer's metadata.
func ResolveCredential(opts ...ResolveOpt) (*ResolvedData, error) {
	credentialConfigMappings, issuerMetadata, _, maskingString, err := processOpts(opts)
	if err != nil {
		return nil, err
	}

	rOpts := mergeOpts(opts)

	if maskingString == nil {
		defaultMaskingString := "•"
		maskingString = &defaultMaskingString
	}

	credentialDisplays, err := buildCredentialDisplaysAllLocale(credentialConfigMappings,
		*maskingString, rOpts.skipNonClaimData)
	if err != nil {
		return nil, err
	}

	issuerOverview := getIssuerDisplayAllLocale(issuerMetadata.LocalizedIssuerDisplays)

	return &ResolvedData{
		LocalizedIssuer: issuerOverview,
		Credential:      credentialDisplays,
	}, nil
}

// ResolveCredentialOffer resolves display information for some offered credentials based on an issuer's metadata.
// The CredentialDisplays in the returned ResolvedDisplayData object correspond to the offered credential types
// passed in and are in the same order.
func ResolveCredentialOffer(
	metadata *issuer.Metadata, offeredCredentialTypes [][]string, preferredLocale string,
) *ResolvedDisplayData {
	issuerOverview := getIssuerDisplay(metadata.LocalizedIssuerDisplays, preferredLocale)

	return &ResolvedDisplayData{
		IssuerDisplay: issuerOverview,
		CredentialDisplays: buildCredentialOfferingDisplays(offeredCredentialTypes,
			metadata.CredentialConfigurationsSupported, preferredLocale),
	}
}
