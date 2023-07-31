/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import "github.com/trustbloc/wallet-sdk/pkg/models/issuer"

// Logo represents display information for a logo.
type Logo struct {
	logo *issuer.Logo
}

// URL returns the URL where this logo's image can be fetched.
func (l *Logo) URL() string {
	return l.logo.URL
}

// AltText returns alt text for this logo.
func (l *Logo) AltText() string {
	return l.logo.AltText
}
