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

// SupportedCredentials returns an object that can be used to determine the types of credentials that the issuer
// supports issuance of.
func (i *IssuerMetadata) SupportedCredentials() *SupportedCredentials {
	return &SupportedCredentials{supportedCredentials: i.issuerMetadata.CredentialsSupported}
}

// LocalizedIssuerDisplays returns an object that contains display information for the issuer in various locales.
func (i *IssuerMetadata) LocalizedIssuerDisplays() *LocalizedIssuerDisplays {
	return &LocalizedIssuerDisplays{localizedIssuerDisplays: i.issuerMetadata.LocalizedIssuerDisplays}
}

// IssuerTrustInfo represent issuer trust information.
type IssuerTrustInfo struct {
	DID              string
	Domain           string
	CredentialType   string
	CredentialFormat string
}
