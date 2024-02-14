/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

// CredentialConfigurationsSupported represents the credentials (types and formats) that an issuer can issue.
type CredentialConfigurationsSupported struct {
	credentialConfigurations map[issuer.CredentialConfigurationID]*issuer.CredentialConfigurationSupported
}

// Length returns the number of CredentialConfigurationsSupported contained within this object.
func (s *CredentialConfigurationsSupported) Length() int {
	return len(s.credentialConfigurations)
}

// CredentialConfigurationSupported returns the CredentialConfigurationSupported by given credentialConfigurationID.
// If credentialConfigurationID is unknown, then nil is returned.
func (s *CredentialConfigurationsSupported) CredentialConfigurationSupported(
	credentialConfigurationID issuer.CredentialConfigurationID,
) *CredentialConfigurationSupported {
	credentialConf, ok := s.credentialConfigurations[credentialConfigurationID]
	if !ok {
		return nil
	}

	return &CredentialConfigurationSupported{credentialConfigurationSupported: credentialConf}
}

// CredentialConfigurationSupported represents a specific credential (type and format) that an issuer can issue.
type CredentialConfigurationSupported struct {
	credentialConfigurationSupported *issuer.CredentialConfigurationSupported
}

// Format returns this CredentialConfigurationSupported's format.
func (s *CredentialConfigurationSupported) Format() string {
	return s.credentialConfigurationSupported.Format
}

// Types returns this CredentialConfigurationSupported's types.
func (s *CredentialConfigurationSupported) Types() *api.StringArray {
	return &api.StringArray{Strings: s.credentialConfigurationSupported.CredentialDefinition.Type}
}

// LocalizedDisplays returns an object that contains this CredentialConfigurationSupported's
// display data in various locales.
func (s *CredentialConfigurationSupported) LocalizedDisplays() *LocalizedCredentialDisplays {
	return &LocalizedCredentialDisplays{
		localizedCredentialDisplays: s.credentialConfigurationSupported.LocalizedCredentialDisplays,
	}
}
