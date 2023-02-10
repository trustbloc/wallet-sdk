/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"encoding/json"

	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
)

// DisplayData represents display information for some issued credentials based on an issuer's metadata.
type DisplayData struct {
	resolvedDisplayData *goapicredentialschema.ResolvedDisplayData
}

// ParseDisplayData parses the given serialized display data and returns a DisplayData object.
func ParseDisplayData(displayData string) (*DisplayData, error) {
	var parsedDisplayData goapicredentialschema.ResolvedDisplayData

	err := json.Unmarshal([]byte(displayData), &parsedDisplayData)
	if err != nil {
		return nil, err
	}

	return &DisplayData{resolvedDisplayData: &parsedDisplayData}, nil
}

// Serialize serializes this DisplayData object into JSON.
func (d *DisplayData) Serialize() (string, error) {
	resolvedDisplayDataBytes, err := json.Marshal(d.resolvedDisplayData)

	return string(resolvedDisplayDataBytes), err
}

// IssuerDisplay returns the issuer display object.
func (d *DisplayData) IssuerDisplay() *IssuerDisplay {
	return &IssuerDisplay{issuerDisplay: d.resolvedDisplayData.IssuerDisplay}
}

// CredentialDisplaysLength returns the number of credential displays contained within this DisplayData object.
func (d *DisplayData) CredentialDisplaysLength() int {
	return len(d.resolvedDisplayData.CredentialDisplays)
}

// CredentialDisplayAtIndex returns the credential display object at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (d *DisplayData) CredentialDisplayAtIndex(index int) *CredentialDisplay {
	maxIndex := len(d.resolvedDisplayData.CredentialDisplays) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &CredentialDisplay{credentialDisplay: &d.resolvedDisplayData.CredentialDisplays[index]}
}

// IssuerDisplay represents display information about the issuer of some credential(s).
type IssuerDisplay struct {
	issuerDisplay *goapicredentialschema.ResolvedCredentialIssuerDisplay
}

// ParseIssuerDisplay parses the given serialized issuer display data and returns an IssuerDisplay object.
func ParseIssuerDisplay(issuerDisplay string) (*IssuerDisplay, error) {
	var parsedIssuerDisplay goapicredentialschema.ResolvedCredentialIssuerDisplay

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
// Display data for specific claims (e.g. first name, date of birth, etc.) are in the Claims objects.
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
func (c *Claim) Value() string {
	return c.claim.Value
}

// Locale returns the locale corresponding to this claim's display data.
// The locale is determined during the ResolveDisplay call based on the preferred locale passed in and what
// localizations were provided in the issuer's metadata.
func (c *Claim) Locale() string {
	return c.claim.Locale
}
