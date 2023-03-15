/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package display contains functionality that can be used to resolve display values per the OpenID4CI spec.
package display

import (
	"encoding/json"
	"errors"

	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
)

// Data represents display information for some issued credentials based on an issuer's metadata.
type Data struct {
	resolvedDisplayData *goapicredentialschema.ResolvedDisplayData
}

// ParseData parses the given serialized display data and returns a display Data object.
func ParseData(displayData string) (*Data, error) {
	var parsedDisplayData goapicredentialschema.ResolvedDisplayData

	err := json.Unmarshal([]byte(displayData), &parsedDisplayData)
	if err != nil {
		return nil, err
	}

	return &Data{resolvedDisplayData: &parsedDisplayData}, nil
}

// Serialize serializes this display Data object into JSON.
func (d *Data) Serialize() (string, error) {
	resolvedDisplayDataBytes, err := json.Marshal(d.resolvedDisplayData)

	return string(resolvedDisplayDataBytes), err
}

// IssuerDisplay returns the issuer display object.
func (d *Data) IssuerDisplay() *IssuerDisplay {
	return &IssuerDisplay{issuerDisplay: d.resolvedDisplayData.IssuerDisplay}
}

// CredentialDisplaysLength returns the number of credential displays contained within this display Data object.
func (d *Data) CredentialDisplaysLength() int {
	return len(d.resolvedDisplayData.CredentialDisplays)
}

// CredentialDisplayAtIndex returns the credential display object at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (d *Data) CredentialDisplayAtIndex(index int) *CredentialDisplay {
	maxIndex := len(d.resolvedDisplayData.CredentialDisplays) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &CredentialDisplay{credentialDisplay: &d.resolvedDisplayData.CredentialDisplays[index]}
}

// IssuerDisplay represents display information about the issuer of some credential(s).
type IssuerDisplay struct {
	issuerDisplay *goapicredentialschema.ResolvedIssuerDisplay
}

// ParseIssuerDisplay parses the given serialized issuer display data and returns an IssuerDisplay object.
func ParseIssuerDisplay(issuerDisplay string) (*IssuerDisplay, error) {
	var parsedIssuerDisplay goapicredentialschema.ResolvedIssuerDisplay

	err := json.Unmarshal([]byte(issuerDisplay), &parsedIssuerDisplay)
	if err != nil {
		return nil, err
	}

	return &IssuerDisplay{issuerDisplay: &parsedIssuerDisplay}, nil
}

// Serialize serializes this IssuerDisplay object into JSON.
func (d *IssuerDisplay) Serialize() (string, error) {
	issuerDisplayBytes, err := json.Marshal(d.issuerDisplay)

	return string(issuerDisplayBytes), err
}

// Name returns the issuer's display name.
func (d *IssuerDisplay) Name() string {
	return d.issuerDisplay.Name
}

// Locale returns the locale corresponding to this issuer's display name.
// The locale is determined during the ResolveDisplay call based on the preferred locale passed in and what
// localizations were provided in the issuer's metadata.
func (d *IssuerDisplay) Locale() string {
	return d.issuerDisplay.Locale
}

// CredentialDisplay represents display data for a credential.
// Display data for specific claims (e.g. first name, date of birth, etc.) are in the CredentialSubject objects.
type CredentialDisplay struct {
	credentialDisplay *goapicredentialschema.CredentialDisplay
}

// ParseCredentialDisplay parses the given serialized credential display data and returns a CredentialDisplay object.
func ParseCredentialDisplay(credentialDisplay string) (*CredentialDisplay, error) {
	var parsedCredentialDisplay goapicredentialschema.CredentialDisplay

	err := json.Unmarshal([]byte(credentialDisplay), &parsedCredentialDisplay)
	if err != nil {
		return nil, err
	}

	return &CredentialDisplay{credentialDisplay: &parsedCredentialDisplay}, nil
}

// Serialize serializes this CredentialDisplay object into JSON.
func (c *CredentialDisplay) Serialize() (string, error) {
	credentialDisplayBytes, err := json.Marshal(c.credentialDisplay)

	return string(credentialDisplayBytes), err
}

// Overview returns the credential overview display object.
func (c *CredentialDisplay) Overview() *CredentialOverview {
	return &CredentialOverview{c.credentialDisplay.Overview}
}

// ClaimsLength returns the number of claims displays contained within this CredentialDisplay object.
func (c *CredentialDisplay) ClaimsLength() int {
	return len(c.credentialDisplay.Claims)
}

// ClaimAtIndex returns the claim display object at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (c *CredentialDisplay) ClaimAtIndex(index int) *Claim {
	maxIndex := len(c.credentialDisplay.Claims) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &Claim{claim: &c.credentialDisplay.Claims[index]}
}

// CredentialOverview represents display data for a credential as a whole.
type CredentialOverview struct {
	overview *goapicredentialschema.CredentialOverview
}

// Name returns the display name for the credential.
func (c *CredentialOverview) Name() string {
	return c.overview.Name
}

// Logo returns display logo data for the credential.
func (c *CredentialOverview) Logo() *Logo {
	return &Logo{logo: c.overview.Logo}
}

// BackgroundColor returns the background color that should be used when displaying this credential.
func (c *CredentialOverview) BackgroundColor() string {
	return c.overview.BackgroundColor
}

// TextColor returns the text color that should be used when displaying this credential.
func (c *CredentialOverview) TextColor() string {
	return c.overview.TextColor
}

// Locale returns the locale corresponding to this credential overview's display data.
// The locale is determined during the ResolveDisplay call based on the preferred locale passed in and what
// localizations were provided in the issuer's metadata.
func (c *CredentialOverview) Locale() string {
	return c.overview.Locale
}

// Logo represents display information for a logo.
type Logo struct {
	logo *goapicredentialschema.Logo
}

// URL returns the URL where this logo's image can be fetched.
func (l *Logo) URL() string {
	return l.logo.URL
}

// AltText returns alt text for this logo.
func (l *Logo) AltText() string {
	return l.logo.AltText
}

// Claim represents display data for a specific claim.
type Claim struct {
	claim *goapicredentialschema.ResolvedClaim
}

// RawID returns the raw field name (key) from the VC associated with this claim.
// It's not localized or formatted for display.
func (c *Claim) RawID() string {
	return c.claim.RawID
}

// Label returns the display label for this claim.
// For example, if the UI were to display "Given Name: Alice", then the Label would be "Given Name".
func (c *Claim) Label() string {
	return c.claim.Label
}

// ValueType returns the display value type for this claim.
// For example:  "string", "number", "image", etc.
func (c *Claim) ValueType() string {
	return c.claim.ValueType
}

// Value returns the display value for this claim.
// For example, if the UI were to display "Given Name: Alice", then the Value would be "Alice".
// If no special formatting was applied to the display value, then this method will be equivalent to calling RawValue.
func (c *Claim) Value() string {
	if c.claim.Value == "" {
		return c.claim.RawValue
	}

	return c.claim.Value
}

// RawValue returns the raw display value for this claim without any formatting.
// For example, if this claim is masked, this method will return the unmasked version.
// If no special formatting was applied to the display value, then this method will be equivalent to calling Value.
func (c *Claim) RawValue() string {
	return c.claim.RawValue
}

// IsMasked indicates whether this claim's value is masked. If this method returns true, then the Value method
// will return the masked value while the RawValue method will return the unmasked version.
func (c *Claim) IsMasked() bool {
	return c.claim.Mask != ""
}

// Pattern returns the pattern information for this claim.
func (c *Claim) Pattern() string {
	return c.claim.Pattern
}

// HasOrder returns whether this Claim has a specified order in it.
func (c *Claim) HasOrder() bool {
	return c.claim.Order != nil
}

// Order returns the display order for this claim.
// HasOrder should be called first to ensure this claim has a specified order before calling this method.
// This method returns an error if the claim has no specified order.
func (c *Claim) Order() (int, error) {
	if c.claim.Order == nil {
		return -1, errors.New("claim has no specified order")
	}

	return *c.claim.Order, nil
}

// Locale returns the locale corresponding to this claim's display data.
// The locale is determined during the ResolveDisplay call based on the preferred locale passed in and what
// localizations were provided in the issuer's metadata.
func (c *Claim) Locale() string {
	return c.claim.Locale
}
