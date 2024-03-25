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

	"github.com/PaesslerAG/jsonpath"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

var (
	errNoClaimDisplays        error
	errClaimValueNotFoundInVC error
)

func buildCredentialDisplays(
	vcs []*verifiable.Credential,
	credentialConfigurationsSupported map[issuer.CredentialConfigurationID]*issuer.CredentialConfigurationSupported,
	preferredLocale, maskingString string,
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

		for _, credentialConfigurationSupported := range credentialConfigurationsSupported {
			if !haveMatchingTypes(credentialConfigurationSupported, displayVC.Contents().Types) {
				continue
			}

			credentialDisplay, err := buildCredentialDisplay(credentialConfigurationSupported, subject, preferredLocale,
				maskingString)
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
			credentialDisplay := buildDefaultCredentialDisplay(vc.Contents().ID, subject)

			credentialDisplays = append(credentialDisplays, *credentialDisplay)
		}
	}

	return credentialDisplays, nil
}

func buildCredentialOfferingDisplays(offeringTypes [][]string,
	credentialConfigurationsSupported map[issuer.CredentialConfigurationID]*issuer.CredentialConfigurationSupported,
	preferredLocale string,
) []CredentialDisplay {
	var credentialDisplays []CredentialDisplay

	for _, vcTypes := range offeringTypes {
		for _, credentialConfiguration := range credentialConfigurationsSupported {
			if !haveMatchingTypes(credentialConfiguration, vcTypes) {
				continue
			}

			credentialDisplay := &CredentialDisplay{Overview: getOverviewDisplay(credentialConfiguration, preferredLocale)}

			credentialDisplays = append(credentialDisplays, *credentialDisplay)

			break
		}
	}

	return credentialDisplays
}

// The VC is considered to be a match for the supportedCredential if the VC has at least one type that's the same as
// the type specified by the supportCredential (excluding the "VerifiableCredential" type that all VCs have).
func haveMatchingTypes(credentialConfSupported *issuer.CredentialConfigurationSupported, vcTypes []string) bool {
	for _, typeFromVC := range vcTypes {
		// We expect the types in the VC and SupportedCredential to always include VerifiableCredential,
		// so we skip this case.
		if strings.EqualFold(typeFromVC, "VerifiableCredential") {
			continue
		}

		for _, typeFromSupportedCredential := range credentialConfSupported.CredentialDefinition.Type {
			if strings.EqualFold(typeFromVC, typeFromSupportedCredential) {
				return true
			}
		}
	}

	return false
}

func buildCredentialDisplay(
	credentialConfigurationSupported *issuer.CredentialConfigurationSupported,
	subject *verifiable.Subject,
	preferredLocale, maskingString string,
) (*CredentialDisplay, error) {
	resolvedClaims, err := resolveClaims(credentialConfigurationSupported, subject, preferredLocale, maskingString)
	if err != nil {
		return nil, err
	}

	overview := *getOverviewDisplay(credentialConfigurationSupported, preferredLocale)

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
	credentialSubjects := vc.Contents().Subject

	if len(credentialSubjects) != 1 {
		return nil, errors.New("only VCs with one credential subject are supported")
	}

	return &credentialSubjects[0], nil
}

func resolveClaims(
	credentialConfigurationSupported *issuer.CredentialConfigurationSupported,
	credentialSubject *verifiable.Subject,
	preferredLocale, maskingString string,
) ([]ResolvedClaim, error) {
	var resolvedClaims []ResolvedClaim

	for fieldName, claim := range credentialConfigurationSupported.CredentialDefinition.CredentialSubject {
		resolvedClaim, err := resolveClaim(
			fieldName, claim, credentialSubject, credentialConfigurationSupported, preferredLocale, maskingString)
		if err != nil && !errors.Is(err, errNoClaimDisplays) && !errors.Is(err, errClaimValueNotFoundInVC) {
			return nil, err
		}

		if resolvedClaim != nil {
			resolvedClaims = append(resolvedClaims, *resolvedClaim)
		}
	}

	return resolvedClaims, nil
}

func resolveClaim(
	fieldName string,
	claim *issuer.Claim,
	credentialSubject *verifiable.Subject,
	credentialConfigurationSupported *issuer.CredentialConfigurationSupported,
	preferredLocale, maskingString string,
) (*ResolvedClaim, error) {
	if len(claim.LocalizedClaimDisplays) == 0 {
		return nil, errNoClaimDisplays
	}

	label, labelLocale := getLocalizedLabel(preferredLocale, claim)

	untypedValue := getMatchingClaimValue(credentialSubject, fieldName)
	if untypedValue == nil {
		return nil, errClaimValueNotFoundInVC
	}

	rawValue := fmt.Sprintf("%v", untypedValue)

	var value *string

	if claim.Mask != "" {
		maskedValue, err := getMaskedValue(rawValue, claim.Mask, maskingString)
		if err != nil {
			return nil, err
		}

		value = &maskedValue
	}

	var order *int

	orderAsInt, err := credentialConfigurationSupported.ClaimOrderAsInt(fieldName)
	if err == nil {
		order = &orderAsInt
	}

	return &ResolvedClaim{
		RawID:     fieldName,
		Label:     label,
		ValueType: claim.ValueType,
		Order:     order,
		RawValue:  rawValue,
		Value:     value,
		Pattern:   claim.Pattern,
		Mask:      claim.Mask,
		Locale:    labelLocale,
	}, nil
}

func getMaskedValue(rawValue, maskingPattern, maskingString string) (string, error) {
	// Trim "regex(" from the beginning and ")" from the end
	regex := maskingPattern[6 : len(maskingPattern)-1]

	r, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}

	// Always use the first submatch.
	valueToBeMasked := r.ReplaceAllString(rawValue, "$1")

	maskedValue := strings.ReplaceAll(rawValue, valueToBeMasked, strings.Repeat(maskingString, len(valueToBeMasked)))

	return maskedValue, nil
}

// Returns the localized name and the actual locale used (which may differ from the user's preferred locale, depending
// on what is available). If no preferred locale is specified, then the first available locale is used.
func getLocalizedLabel(preferredLocale string, claim *issuer.Claim) (string, string) {
	if preferredLocale == "" {
		return claim.LocalizedClaimDisplays[0].Name, claim.LocalizedClaimDisplays[0].Locale
	}

	for _, claimDisplay := range claim.LocalizedClaimDisplays {
		if strings.EqualFold(preferredLocale, claimDisplay.Locale) {
			return claimDisplay.Name, claimDisplay.Locale
		}
	}

	return claim.LocalizedClaimDisplays[0].Name, claim.LocalizedClaimDisplays[0].Locale
}

// Returns nil if no matching claim value could be found.
func getMatchingClaimValue(credentialSubject *verifiable.Subject, fieldName string) interface{} {
	if strings.EqualFold(fieldName, "ID") {
		if credentialSubject.ID == "" {
			return nil
		}

		return credentialSubject.ID
	}

	if strings.HasPrefix(fieldName, "$.") {
		value := findMatchingClaimUsingJSONPath(credentialSubject.CustomFields, fieldName)
		if value != nil {
			return value
		}
	} else {
		value, exists := credentialSubject.CustomFields[fieldName]
		if exists {
			return value
		}

		value = findMatchingClaimValueInMap(credentialSubject.CustomFields, fieldName)
		if value != nil {
			return value
		}
	}

	return nil
}

func findMatchingClaimUsingJSONPath(claims map[string]interface{}, fieldName string) interface{} {
	value, err := jsonpath.Get(fieldName, claims)
	if err != nil {
		return nil
	}

	return value
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

func getOverviewDisplay(
	credentialConfigurationSupported *issuer.CredentialConfigurationSupported,
	preferredLocale string,
) *CredentialOverview {
	if preferredLocale == "" {
		return issuerCredentialDisplayToResolvedCredentialOverview(
			&credentialConfigurationSupported.LocalizedCredentialDisplays[0],
		)
	}

	for i := range credentialConfigurationSupported.LocalizedCredentialDisplays {
		if strings.EqualFold(preferredLocale, credentialConfigurationSupported.LocalizedCredentialDisplays[i].Locale) {
			return issuerCredentialDisplayToResolvedCredentialOverview(
				&credentialConfigurationSupported.LocalizedCredentialDisplays[i],
			)
		}
	}

	return issuerCredentialDisplayToResolvedCredentialOverview(
		&credentialConfigurationSupported.LocalizedCredentialDisplays[0],
	)
}

func issuerCredentialDisplayToResolvedCredentialOverview(
	issuerCredentialOverview *issuer.LocalizedCredentialDisplay,
) *CredentialOverview {
	resolvedCredentialOverview := &CredentialOverview{
		Name:            issuerCredentialOverview.Name,
		Locale:          issuerCredentialOverview.Locale,
		BackgroundColor: issuerCredentialOverview.BackgroundColor,
		TextColor:       issuerCredentialOverview.TextColor,
		Logo:            convertLogo(issuerCredentialOverview.Logo),
	}

	return resolvedCredentialOverview
}

func convertLogo(logo *issuer.Logo) *Logo {
	if logo != nil {
		return &Logo{
			URL:     logo.URL,
			AltText: logo.AltText,
		}
	}

	return nil
}
