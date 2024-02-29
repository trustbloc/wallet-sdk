/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

// IssuerMetadata represents metadata about an issuer as obtained from their .well-known OpenID configuration.
type IssuerMetadata struct {
	issuerMetadata *issuer.Metadata
}

// CredentialIssuer returns the issuer's identifier.
func (i *IssuerMetadata) CredentialIssuer() string {
	return i.issuerMetadata.CredentialIssuer
}

// SupportedCredentials returns an object that can be used to determine the types of
// credentials that the issuer supports issuance of.
func (i *IssuerMetadata) SupportedCredentials() *SupportedCredentials {
	supportedCredentials := &SupportedCredentials{
		credentialConfigurations: i.issuerMetadata.CredentialConfigurationsSupported,
	}

	for _, credentialConf := range i.issuerMetadata.CredentialConfigurationsSupported {
		supportedCredentials.credentialConfigurationsList = append(
			supportedCredentials.credentialConfigurationsList, credentialConf)
	}

	return supportedCredentials
}

// LocalizedIssuerDisplays returns an object that contains display information for the issuer in various locales.
func (i *IssuerMetadata) LocalizedIssuerDisplays() *LocalizedIssuerDisplays {
	return &LocalizedIssuerDisplays{localizedIssuerDisplays: i.issuerMetadata.LocalizedIssuerDisplays}
}

// IssuerMetadataToGoImpl unwrap original issuer.Metadata from IssuerMetadata wrapper.
func IssuerMetadataToGoImpl(wrapped *IssuerMetadata) *issuer.Metadata {
	return wrapped.issuerMetadata
}

// IssuerMetadataFromGoImpl wrap original issuer.Metadata into IssuerMetadata wrapper.
func IssuerMetadataFromGoImpl(goAPIIssuerMetadata *issuer.Metadata) *IssuerMetadata {
	return &IssuerMetadata{issuerMetadata: goAPIIssuerMetadata}
}

// IssuerTrustInfo represent issuer trust information.
type IssuerTrustInfo struct {
	DID              string
	Domain           string
	CredentialOffers []*CredentialOffer
}

// CredentialOffer contains data related to a credential type being offered in an issuance request.
type CredentialOffer struct {
	CredentialType             string
	CredentialFormat           string
	ClientAttestationRequested bool
}

// OfferLength returns the number offers.
func (i *IssuerTrustInfo) OfferLength() int {
	return len(i.CredentialOffers)
}

// OfferAtIndex returns offer by index.
func (i *IssuerTrustInfo) OfferAtIndex(index int) *CredentialOffer {
	return i.CredentialOffers[index]
}
