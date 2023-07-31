/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "github.com/trustbloc/wallet-sdk/pkg/models/issuer"

// LocalizedCredentialDisplays represents display information for a credential in various locales.
type LocalizedCredentialDisplays struct {
	localizedCredentialDisplays []issuer.LocalizedCredentialDisplay
}

// AtIndex returns the LocalizedCredentialDisplays at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (l *LocalizedCredentialDisplays) AtIndex(index int) *LocalizedCredentialDisplay {
	maxIndex := len(l.localizedCredentialDisplays) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &LocalizedCredentialDisplay{localizedCredentialDisplay: &l.localizedCredentialDisplays[index]}
}

// Length returns the number of LocalizedIssuerDisplays contained within this object.
func (l *LocalizedCredentialDisplays) Length() int {
	return len(l.localizedCredentialDisplays)
}

// LocalizedCredentialDisplay represents display information for a credential in a specific locale.
type LocalizedCredentialDisplay struct {
	localizedCredentialDisplay *issuer.LocalizedCredentialDisplay
}

// Name returns this LocalizedCredentialDisplay's name.
func (l *LocalizedCredentialDisplay) Name() string {
	return l.localizedCredentialDisplay.Name
}

// Locale returns this LocalizedCredentialDisplay's locale.
func (l *LocalizedCredentialDisplay) Locale() string {
	return l.localizedCredentialDisplay.Locale
}

// Logo returns this LocalizedCredentialDisplay's logo.
func (l *LocalizedCredentialDisplay) Logo() *Logo {
	return &Logo{logo: l.localizedCredentialDisplay.Logo}
}

// BackgroundColor returns this LocalizedCredentialDisplay's background color.
func (l *LocalizedCredentialDisplay) BackgroundColor() string {
	return l.localizedCredentialDisplay.BackgroundColor
}

// TextColor returns this LocalizedCredentialDisplay's text color.
func (l *LocalizedCredentialDisplay) TextColor() string {
	return l.localizedCredentialDisplay.TextColor
}
