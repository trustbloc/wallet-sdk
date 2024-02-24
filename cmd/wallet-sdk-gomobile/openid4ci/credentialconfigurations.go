/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

// SupportedCredentials represents the credentials (types and formats) that an issuer can issue.
type SupportedCredentials struct {
	credentialConfigurations map[issuer.CredentialConfigurationID]*issuer.CredentialConfigurationSupported
}

// Length returns the number of SupportedCredentials contained within this object.
func (s *SupportedCredentials) Length() int {
	return len(s.credentialConfigurations)
}

// CredentialConfigurationSupported returns the SupportedCredential by given credentialConfigurationID.
// If credentialConfigurationID is unknown, then nil is returned.
func (s *SupportedCredentials) CredentialConfigurationSupported(
	credentialConfigurationID issuer.CredentialConfigurationID,
) *SupportedCredential {
	credentialConf, ok := s.credentialConfigurations[credentialConfigurationID]
	if !ok {
		return nil
	}

	return &SupportedCredential{credentialConfigurationSupported: credentialConf}
}

// SupportedCredential represents a specific credential (type and format) that an issuer can issue.
type SupportedCredential struct {
	credentialConfigurationSupported *issuer.CredentialConfigurationSupported
}

// Format returns this SupportedCredential's format.
func (s *SupportedCredential) Format() string {
	return s.credentialConfigurationSupported.Format
}

// Types returns this SupportedCredential's types.
func (s *SupportedCredential) Types() *api.StringArray {
	return &api.StringArray{Strings: s.credentialConfigurationSupported.CredentialDefinition.Type}
}

// LocalizedDisplays returns an object that contains this SupportedCredential's
// display data in various locales.
func (s *SupportedCredential) LocalizedDisplays() *LocalizedCredentialDisplays {
	return &LocalizedCredentialDisplays{
		localizedCredentialDisplays: s.credentialConfigurationSupported.LocalizedCredentialDisplays,
	}
}
