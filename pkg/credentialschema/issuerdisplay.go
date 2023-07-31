/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema

import (
	"strings"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

func getIssuerDisplay(issuerDisplays []issuer.LocalizedIssuerDisplay, locale string) *ResolvedIssuerDisplay {
	if len(issuerDisplays) == 0 {
		return nil
	}

	if locale == "" {
		return &ResolvedIssuerDisplay{
			Name:   issuerDisplays[0].Name,
			Locale: issuerDisplays[0].Locale,
		}
	}

	for _, issuerDisplay := range issuerDisplays {
		if strings.EqualFold(issuerDisplay.Locale, locale) {
			return &ResolvedIssuerDisplay{
				Name:   issuerDisplay.Name,
				Locale: issuerDisplay.Locale,
			}
		}
	}

	return &ResolvedIssuerDisplay{
		Name:   issuerDisplays[0].Name,
		Locale: issuerDisplays[0].Locale,
	}
}
