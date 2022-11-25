/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema

import (
	"strings"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

func getIssuerDisplay(credentialIssuer *issuer.CredentialIssuer,
	locale string,
) *ResolvedCredentialIssuerDisplay {
	if credentialIssuer == nil || len(credentialIssuer.Displays) == 0 {
		return nil
	}

	if locale == "" {
		return &ResolvedCredentialIssuerDisplay{
			Name:   credentialIssuer.Displays[0].Name,
			Locale: credentialIssuer.Displays[0].Locale,
		}
	}

	for _, issuerDisplay := range credentialIssuer.Displays {
		if strings.EqualFold(issuerDisplay.Locale, locale) {
			return &ResolvedCredentialIssuerDisplay{
				Name:   issuerDisplay.Name,
				Locale: issuerDisplay.Locale,
			}
		}
	}

	return &ResolvedCredentialIssuerDisplay{
		Name:   credentialIssuer.Displays[0].Name,
		Locale: credentialIssuer.Displays[0].Locale,
	}
}
