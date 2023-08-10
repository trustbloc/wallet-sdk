/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "github.com/trustbloc/wallet-sdk/pkg/models/issuer"

// LocalizedIssuerDisplays represents display information for an issuer in various locales.
type LocalizedIssuerDisplays struct {
	localizedIssuerDisplays []issuer.LocalizedIssuerDisplay
}

// AtIndex returns the LocalizedIssuerDisplays at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (l *LocalizedIssuerDisplays) AtIndex(index int) *LocalizedIssuerDisplay {
	maxIndex := len(l.localizedIssuerDisplays) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &LocalizedIssuerDisplay{localizedIssuerDisplay: &l.localizedIssuerDisplays[index]}
}

// Length returns the number of LocalizedIssuerDisplays contained within this object.
func (l *LocalizedIssuerDisplays) Length() int {
	return len(l.localizedIssuerDisplays)
}

// LocalizedIssuerDisplay represents display information for an issuer in a specific locale.
type LocalizedIssuerDisplay struct {
	localizedIssuerDisplay *issuer.LocalizedIssuerDisplay
}

// Name returns this LocalizedIssuerDisplay's name.
func (l *LocalizedIssuerDisplay) Name() string {
	return l.localizedIssuerDisplay.Name
}

// Locale returns this LocalizedIssuerDisplay's Locale.
func (l *LocalizedIssuerDisplay) Locale() string {
	return l.localizedIssuerDisplay.Locale
}

// URL returns this LocalizedIssuerDisplay's URL.
func (l *LocalizedIssuerDisplay) URL() string {
	return l.localizedIssuerDisplay.URL
}

// Logo returns this LocalizedIssuerDisplay's logo.
// If it has no logo, then nil/null is returned instead.
func (l *LocalizedIssuerDisplay) Logo() *Logo {
	if l.localizedIssuerDisplay.Logo == nil {
		return nil
	}

	return &Logo{logo: l.localizedIssuerDisplay.Logo}
}

// BackgroundColor returns this LocalizedIssuerDisplay's background color.
func (l *LocalizedIssuerDisplay) BackgroundColor() string {
	return l.localizedIssuerDisplay.BackgroundColor
}

// TextColor returns this LocalizedIssuerDisplay's text color.
func (l *LocalizedIssuerDisplay) TextColor() string {
	return l.localizedIssuerDisplay.TextColor
}
