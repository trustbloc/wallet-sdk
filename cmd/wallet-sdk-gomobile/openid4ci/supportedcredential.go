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
	supportedCredentials []issuer.SupportedCredential
}

// Length returns the number of SupportedCredentials contained within this object.
func (s *SupportedCredentials) Length() int {
	return len(s.supportedCredentials)
}

// AtIndex returns the SupportedCredential at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (s *SupportedCredentials) AtIndex(index int) *SupportedCredential {
	maxIndex := len(s.supportedCredentials) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &SupportedCredential{supportedCredential: &s.supportedCredentials[index]}
}

// SupportedCredential represents a specific credential (type and format) that an issuer can issue.
type SupportedCredential struct {
	supportedCredential *issuer.SupportedCredential
}

// Format returns this SupportedCredential's format.
func (s *SupportedCredential) Format() string {
	return s.supportedCredential.Format
}

// Types returns this SupportedCredential's types.
func (s *SupportedCredential) Types() *api.StringArray {
	return &api.StringArray{Strings: s.supportedCredential.Types}
}

// ID returns this SupportedCredential's ID.
func (s *SupportedCredential) ID() string {
	return s.supportedCredential.ID
}

// LocalizedDisplays returns an object that contains this SupportedCredential's display data in various locales.
func (s *SupportedCredential) LocalizedDisplays() *LocalizedCredentialDisplays {
	return &LocalizedCredentialDisplays{localizedCredentialDisplays: s.supportedCredential.LocalizedCredentialDisplays}
}
