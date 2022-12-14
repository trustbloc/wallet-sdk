/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

const localeNotApplicable = "N/A"

func buildCredentialDisplays(vcs []*verifiable.Credential, credentialsSupported map[string]issuer.SupportedCredential,
	preferredLocale string,
) ([]CredentialDisplay, error) {
	var credentialDisplays []CredentialDisplay

	for _, vc := range vcs {
		subject, err := getSubject(vc)
		if err != nil {
			return nil, err
		}

		var foundMatchingType bool

		// Note that the actual ID here isn't important here - what matters are the types listed within the
		// supported credential object.
		for id := range credentialsSupported {
			supportedCredential := credentialsSupported[id]

			if haveMatchingTypes(&supportedCredential, vc) {
				credentialDisplay := buildCredentialDisplay(&supportedCredential, subject, preferredLocale)

				credentialDisplays = append(credentialDisplays, *credentialDisplay)

				foundMatchingType = true

				break
			}
		}

		if !foundMatchingType {
			// In case the issuer's metadata doesn't contain display info for this type of credential for some
			// reason, we build up a default/generic type of credential display based only on information in the VC.
			// It'll be functional, but won't be pretty.
			credentialDisplay := buildDefaultCredentialDisplay(vc.ID, subject)

			credentialDisplays = append(credentialDisplays, *credentialDisplay)
		}
	}

	return credentialDisplays, nil
}

// The VC is considered to be a match for the supportedCredential if the VC has at least one type that's the same as
// the type specified by the supportCredential (excluding the "VerifiableCredential" type that all VCs have).
func haveMatchingTypes(supportedCredential *issuer.SupportedCredential, vc *verifiable.Credential) bool {
	for _, typeFromVC := range vc.Types {
		// We expect the types in the VC and SupportedCredential to always include VerifiableCredential,
		// so we skip this case.
		if strings.EqualFold(typeFromVC, "VerifiableCredential") {
			continue
		}

		for _, typeFromSupportedCredentials := range supportedCredential.Types {
			if strings.EqualFold(typeFromVC, typeFromSupportedCredentials) {
				return true
			}
		}
	}

	return false
}

func buildCredentialDisplay(supportedCredential *issuer.SupportedCredential, subject *verifiable.Subject,
	preferredLocale string,
) *CredentialDisplay {
	resolvedClaims := resolveClaims(supportedCredential, subject, preferredLocale)

	overview := *getOverviewDisplay(supportedCredential, preferredLocale)

	return &CredentialDisplay{Overview: &overview, Claims: resolvedClaims}
}

func buildDefaultCredentialDisplay(vcID string, subject *verifiable.Subject) *CredentialDisplay {
	var claims []ResolvedClaim

	id := subject.ID

	if id != "" {
		claims = append(claims, ResolvedClaim{
			Label:  "ID",
			Value:  id,
			Locale: localeNotApplicable,
		})
	}

	for name, rawValue := range subject.CustomFields {
		value := fmt.Sprintf("%v", rawValue)

		claims = append(claims, ResolvedClaim{
			Label:  name,
			Value:  value,
			Locale: localeNotApplicable,
		})
	}

	credentialOverview := issuer.CredentialDisplay{Name: vcID, Locale: localeNotApplicable}

	return &CredentialDisplay{Overview: &credentialOverview, Claims: claims}
}

func getSubject(vc *verifiable.Credential) (*verifiable.Subject, error) {
	credentialSubjects, ok := vc.Subject.([]verifiable.Subject)
	if !ok {
		return nil, errors.New("unsupported vc subject type")
	}

	if len(credentialSubjects) != 1 {
		return nil, errors.New("only VCs with one credential subject are supported")
	}

	return &credentialSubjects[0], nil
}

func resolveClaims(supportedCredential *issuer.SupportedCredential, credentialSubject *verifiable.Subject,
	preferredLocale string,
) []ResolvedClaim {
	var resolvedClaims []ResolvedClaim

	for fieldName, claim := range supportedCredential.CredentialSubject {
		claim := claim // Resolves implicit memory aliasing warning from linter

		resolvedClaim := resolveClaim(fieldName, &claim, credentialSubject, preferredLocale)

		if resolvedClaim != nil {
			resolvedClaims = append(resolvedClaims, *resolvedClaim)
		}
	}

	return resolvedClaims
}

func resolveClaim(fieldName string, claim *issuer.Claim, credentialSubject *verifiable.Subject,
	preferredLocale string,
) *ResolvedClaim {
	if len(claim.Displays) == 0 {
		return nil
	}

	name, nameLocale := getLocalizedName(preferredLocale, claim)

	rawValue, exists := getValue(credentialSubject, fieldName)
	if !exists {
		return &ResolvedClaim{
			Label:  name,
			Value:  "",
			Locale: nameLocale,
		}
	}

	value := fmt.Sprintf("%v", rawValue)

	return &ResolvedClaim{
		Label:  name,
		Value:  value,
		Locale: nameLocale,
	}
}

// Returns the localized name and the actual locale used (which may differ from the user's preferred locale, depending
// on what is available). If no preferred locale is specified, then the first available locale is used.
func getLocalizedName(preferredLocale string, claim *issuer.Claim) (string, string) {
	if preferredLocale == "" {
		return claim.Displays[0].Name, claim.Displays[0].Locale
	}

	for _, claimDisplay := range claim.Displays {
		if strings.EqualFold(preferredLocale, claimDisplay.Locale) {
			return claimDisplay.Name, claimDisplay.Locale
		}
	}

	return claim.Displays[0].Name, claim.Displays[0].Locale
}

func getValue(credentialSubject *verifiable.Subject, fieldName string) (interface{}, bool) {
	if strings.EqualFold(fieldName, "ID") {
		if credentialSubject.ID == "" {
			return "", false
		}

		return credentialSubject.ID, true
	}

	value, exists := credentialSubject.CustomFields[fieldName]

	return value, exists
}

func getOverviewDisplay(supportedCredential *issuer.SupportedCredential,
	preferredLocale string,
) *issuer.CredentialDisplay {
	if preferredLocale == "" {
		return &supportedCredential.Displays[0]
	}

	for _, credentialDisplay := range supportedCredential.Displays {
		if strings.EqualFold(preferredLocale, credentialDisplay.Locale) {
			return &credentialDisplay
		}
	}

	return &supportedCredential.Displays[0]
}
