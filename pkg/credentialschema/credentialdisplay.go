/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialschema

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

var (
	errNoClaimDisplays        error
	errClaimValueNotFoundInVC error
)

func buildCredentialDisplays(vcs []*verifiable.Credential, credentialsSupported []issuer.SupportedCredential,
	preferredLocale string,
) ([]CredentialDisplay, error) {
	var credentialDisplays []CredentialDisplay

	for _, vc := range vcs {
		// The call below creates a copy of the VC with the selective disclosures merged into the credential subject.
		displayVC, err := vc.CreateDisplayCredential(verifiable.DisplayAllDisclosures())
		if err != nil {
			return nil, err
		}

		subject, err := getSubject(displayVC) // Will contain both selective and non-selective disclosures.
		if err != nil {
			return nil, err
		}

		var foundMatchingType bool

		for i := range credentialsSupported {
			if !haveMatchingTypes(&credentialsSupported[i], displayVC) {
				continue
			}

			credentialDisplay, err := buildCredentialDisplay(&credentialsSupported[i], subject, preferredLocale)
			if err != nil {
				return nil, err
			}

			credentialDisplays = append(credentialDisplays, *credentialDisplay)

			foundMatchingType = true

			break
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

		for _, typeFromSupportedCredential := range supportedCredential.Types {
			if strings.EqualFold(typeFromVC, typeFromSupportedCredential) {
				return true
			}
		}
	}

	return false
}

func buildCredentialDisplay(supportedCredential *issuer.SupportedCredential, subject *verifiable.Subject,
	preferredLocale string,
) (*CredentialDisplay, error) {
	resolvedClaims, err := resolveClaims(supportedCredential, subject, preferredLocale)
	if err != nil {
		return nil, err
	}

	overview := *getOverviewDisplay(supportedCredential, preferredLocale)

	return &CredentialDisplay{Overview: &overview, Claims: resolvedClaims}, nil
}

func buildDefaultCredentialDisplay(vcID string, subject *verifiable.Subject) *CredentialDisplay {
	var claims []ResolvedClaim

	id := subject.ID

	if id != "" {
		claims = append(claims, ResolvedClaim{
			RawID:     "id",
			RawValue:  id,
			ValueType: "string",
		})
	}

	for name, untypedValue := range subject.CustomFields {
		value := fmt.Sprintf("%v", untypedValue)

		claims = append(claims, ResolvedClaim{
			RawID:    name,
			RawValue: value,
		})
	}

	credentialOverview := CredentialOverview{Name: vcID}

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
) ([]ResolvedClaim, error) {
	var resolvedClaims []ResolvedClaim

	for fieldName, claim := range supportedCredential.CredentialSubject {
		claim := claim // Resolves implicit memory aliasing warning from linter

		resolvedClaim, err := resolveClaim(fieldName, &claim, credentialSubject, preferredLocale)
		if err != nil && !errors.Is(err, errNoClaimDisplays) && !errors.Is(err, errClaimValueNotFoundInVC) {
			return nil, err
		}

		if resolvedClaim != nil {
			resolvedClaims = append(resolvedClaims, *resolvedClaim)
		}
	}

	return resolvedClaims, nil
}

func resolveClaim(fieldName string, claim *issuer.Claim, credentialSubject *verifiable.Subject,
	preferredLocale string,
) (*ResolvedClaim, error) {
	if len(claim.Displays) == 0 {
		return nil, errNoClaimDisplays
	}

	label, labelLocale := getLocalizedLabel(preferredLocale, claim)

	untypedValue := getMatchingClaimValue(credentialSubject, fieldName)
	if untypedValue == nil {
		return nil, errClaimValueNotFoundInVC
	}

	rawValue := fmt.Sprintf("%v", untypedValue)

	var value string

	if isMaskPattern(claim.Pattern) {
		maskedValue, err := getMaskedValue(rawValue, claim.Pattern)
		if err != nil {
			return nil, err
		}

		value = maskedValue
	}

	return &ResolvedClaim{
		RawID:     fieldName,
		Label:     label,
		ValueType: claim.ValueType,
		Order:     claim.Order,
		RawValue:  rawValue,
		Value:     value,
		Pattern:   claim.Pattern,
		Locale:    labelLocale,
	}, nil
}

func isMaskPattern(pattern string) bool {
	patternSplit := strings.Split(pattern, ":")

	if len(patternSplit) == 2 &&
		patternSplit[0] == "mask" &&
		strings.HasPrefix(patternSplit[1], "regex(") &&
		strings.HasSuffix(patternSplit[1], ")") {
		return true
	}

	return false
}

func getMaskedValue(rawValue, pattern string) (string, error) {
	patternSplitByColons := strings.Split(pattern, ":")

	regexPattern := patternSplitByColons[1]

	// Trim "regex(" from the beginning and ")" from the end
	regex := regexPattern[6 : len(regexPattern)-1]

	r, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}

	// Always use the first submatch.
	valueToBeMasked := r.ReplaceAllString(rawValue, "$1")

	maskedValue := strings.ReplaceAll(rawValue, valueToBeMasked, strings.Repeat("*", len(valueToBeMasked)))

	return maskedValue, nil
}

// Returns the localized name and the actual locale used (which may differ from the user's preferred locale, depending
// on what is available). If no preferred locale is specified, then the first available locale is used.
func getLocalizedLabel(preferredLocale string, claim *issuer.Claim) (string, string) {
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

// Returns nil if no matching claim value could be found.
func getMatchingClaimValue(credentialSubject *verifiable.Subject, fieldName string) interface{} {
	if strings.EqualFold(fieldName, "ID") {
		if credentialSubject.ID == "" {
			return nil
		}

		return credentialSubject.ID
	}

	value, exists := credentialSubject.CustomFields[fieldName]
	if exists {
		return value
	}

	value = findMatchingClaimValueInMap(credentialSubject.CustomFields, fieldName)
	if value != nil {
		return value
	}

	return nil
}

// If nil is returned, then no matching claim was found.
func findMatchingClaimValueInMap(claims map[string]interface{}, fieldName string) interface{} {
	claim, found := claims[fieldName]
	if found {
		return claim
	}

	for _, value := range claims {
		valueAsMap, ok := value.(map[string]interface{})
		if ok {
			return findMatchingClaimValueInMap(valueAsMap, fieldName)
		}
	}

	return nil
}

func getOverviewDisplay(supportedCredential *issuer.SupportedCredential,
	preferredLocale string,
) *CredentialOverview {
	if preferredLocale == "" {
		return issuerCredentialDisplayToResolvedCredentialOverview(&supportedCredential.Overview[0])
	}

	for i := range supportedCredential.Overview {
		if strings.EqualFold(preferredLocale, supportedCredential.Overview[i].Locale) {
			return issuerCredentialDisplayToResolvedCredentialOverview(&supportedCredential.Overview[i])
		}
	}

	return issuerCredentialDisplayToResolvedCredentialOverview(&supportedCredential.Overview[0])
}

func issuerCredentialDisplayToResolvedCredentialOverview(
	issuerCredentialOverview *issuer.CredentialOverview,
) *CredentialOverview {
	resolvedCredentialOverview := &CredentialOverview{
		Name:            issuerCredentialOverview.Name,
		Locale:          issuerCredentialOverview.Locale,
		BackgroundColor: issuerCredentialOverview.BackgroundColor,
		TextColor:       issuerCredentialOverview.TextColor,
	}

	if issuerCredentialOverview.Logo != nil {
		resolvedCredentialOverview.Logo = &Logo{
			URL:     issuerCredentialOverview.Logo.URL,
			AltText: issuerCredentialOverview.Logo.AltText,
		}
	}

	return resolvedCredentialOverview
}
