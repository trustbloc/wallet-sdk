/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package display contains functionality that can be used to resolve display values per the OpenID4CI spec.
package display

import (
	"encoding/json"
	"errors"

	goapicredentialschema "github.com/trustbloc/wallet-sdk/pkg/credentialschema"
)

// Issuer represents display information about the issuer of some credential(s).
type Issuer struct {
	issuerDisplay *goapicredentialschema.ResolvedIssuerDisplay
}

// Serialize serializes this IssuerDisplay object into JSON.
func (d *Issuer) Serialize() (string, error) {
	issuerDisplayBytes, err := json.Marshal(d.issuerDisplay)

	return string(issuerDisplayBytes), err
}

// Name returns the issuer's display name.
func (d *Issuer) Name() string {
	return d.issuerDisplay.Name
}

// Locale returns the locale corresponding to this issuer's display name.
// The locale is determined during the ResolveDisplay call based on the preferred locale passed in and what
// localizations were provided in the issuer's metadata.
func (d *Issuer) Locale() string {
	return d.issuerDisplay.Locale
}

// URL returns this IssuerDisplay's URL.
func (d *Issuer) URL() string {
	return d.issuerDisplay.URL
}

// Logo returns this IssuerDisplay's logo.
// If it has no logo, then nil/null is returned instead.
func (d *Issuer) Logo() *Logo {
	if d.issuerDisplay.Logo == nil {
		return nil
	}

	return &Logo{logo: d.issuerDisplay.Logo}
}

// BackgroundColor returns this LocalizedIssuerDisplay's background color.
func (d *Issuer) BackgroundColor() string {
	return d.issuerDisplay.BackgroundColor
}

// TextColor returns this IssuerDisplay's text color.
func (d *Issuer) TextColor() string {
	return d.issuerDisplay.TextColor
}

// Resolved represents display information for all locales for the issued credentials based on an issuer's metadata.
type Resolved struct {
	resolvedDisplayData *goapicredentialschema.ResolvedData
}

// ParseResolvedData parses the given serialized display data and returns a display Data object.
func ParseResolvedData(displayData string) (*Resolved, error) {
	var parsedDisplayData goapicredentialschema.ResolvedData

	err := json.Unmarshal([]byte(displayData), &parsedDisplayData)
	if err != nil {
		return nil, err
	}

	return &Resolved{resolvedDisplayData: &parsedDisplayData}, nil
}

// Serialize serializes this display Data object into JSON.
func (d *Resolved) Serialize() (string, error) {
	resolvedDisplayDataBytes, err := json.Marshal(d.resolvedDisplayData)

	return string(resolvedDisplayDataBytes), err
}

// LocalizedIssuersLength returns the number of different locales supported for the issuer displays.
func (d *Resolved) LocalizedIssuersLength() int {
	return len(d.resolvedDisplayData.LocalizedIssuer)
}

// LocalizedIssuerAtIndex returns the issuer display object at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (d *Resolved) LocalizedIssuerAtIndex(index int) *Issuer {
	maxIndex := len(d.resolvedDisplayData.LocalizedIssuer) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &Issuer{issuerDisplay: &d.resolvedDisplayData.LocalizedIssuer[index]}
}

// CredentialsLength returns the number of credential displays contained within this display Data object.
func (d *Resolved) CredentialsLength() int {
	return len(d.resolvedDisplayData.Credential)
}

// CredentialAtIndex returns the credential display object at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (d *Resolved) CredentialAtIndex(index int) *Credential {
	maxIndex := len(d.resolvedDisplayData.Credential) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &Credential{credentialDisplay: &d.resolvedDisplayData.Credential[index]}
}

// Overview represents display data for a credential as a whole.
type Overview struct {
	overview *goapicredentialschema.CredentialOverview
}

// Name returns the display name for the credential.
func (c *Overview) Name() string {
	return c.overview.Name
}

// Logo returns display logo data for the credential.
func (c *Overview) Logo() *Logo {
	return &Logo{logo: c.overview.Logo}
}

// BackgroundColor returns the background color that should be used when displaying this credential.
func (c *Overview) BackgroundColor() string {
	return c.overview.BackgroundColor
}

// TextColor returns the text color that should be used when displaying this credential.
func (c *Overview) TextColor() string {
	return c.overview.TextColor
}

// Locale returns the locale corresponding to this credential overview's display data.
// The locale is determined during the ResolveDisplay call based on the preferred locale passed in and what
// localizations were provided in the issuer's metadata.
func (c *Overview) Locale() string {
	return c.overview.Locale
}

// Credential represents display data for a credential.
type Credential struct {
	credentialDisplay *goapicredentialschema.Credential
}

// LocalizedOverviewsLength returns the number of different locales supported for the credential displays.
func (c *Credential) LocalizedOverviewsLength() int {
	return len(c.credentialDisplay.LocalizedOverview)
}

// LocalizedOverviewAtIndex returns the number of different locales supported for the issuer displays.
// If the index passed in is out of bounds, then nil is returned.
func (c *Credential) LocalizedOverviewAtIndex(index int) *Overview {
	maxIndex := len(c.credentialDisplay.LocalizedOverview) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &Overview{overview: &c.credentialDisplay.LocalizedOverview[index]}
}

// SubjectsLength returns the number of credential subject displays contained within this Credential object.
func (c *Credential) SubjectsLength() int {
	return len(c.credentialDisplay.Subject)
}

// SubjectAtIndex returns the credential subject display object at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (c *Credential) SubjectAtIndex(index int) *Subject {
	maxIndex := len(c.credentialDisplay.Subject) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &Subject{claim: &c.credentialDisplay.Subject[index]}
}

// Subject represents display data for a specific credential subject including all locales for the label.
type Subject struct {
	claim *goapicredentialschema.Subject
}

// RawID returns the raw field name (key) from the VC associated with this claim.
// It's not localized or formatted for display.
func (c *Subject) RawID() string {
	return c.claim.RawID
}

// LocalizedLabelsLength returns the number of different locales supported for the credential subject.
func (c *Subject) LocalizedLabelsLength() int {
	return len(c.claim.LocalizedLabels)
}

// LocalizedLabelAtIndex returns the label at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (c *Subject) LocalizedLabelAtIndex(index int) *Label {
	maxIndex := len(c.claim.LocalizedLabels) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &Label{label: &c.claim.LocalizedLabels[index]}
}

// ValueType returns the display value type for this claim.
// For example:  "string", "number", "image", "attachment" etc.
// For type=attachment, ignore the RawValue()  and Value(), instead use Attachment() method.
func (c *Subject) ValueType() string {
	return c.claim.ValueType
}

// Value returns the display value for this claim.
// For example, if the UI were to display "Given Name: Alice", then the Value would be "Alice".
// If no special formatting was applied to the display value, then this method will be equivalent to calling RawValue.
func (c *Subject) Value() string {
	if c.claim.Value == nil {
		return c.claim.RawValue
	}

	return *c.claim.Value
}

// RawValue returns the raw display value for this claim without any formatting.
// For example, if this claim is masked, this method will return the unmasked version.
// If no special formatting was applied to the display value, then this method will be equivalent to calling Value.
func (c *Subject) RawValue() string {
	return c.claim.RawValue
}

// IsMasked indicates whether this claim's value is masked. If this method returns true, then the Value method
// will return the masked value while the RawValue method will return the unmasked version.
func (c *Subject) IsMasked() bool {
	return c.claim.Mask != ""
}

// Pattern returns the pattern information for this claim.
func (c *Subject) Pattern() string {
	return c.claim.Pattern
}

// HasOrder returns whether this Claim has a specified order in it.
func (c *Subject) HasOrder() bool {
	return c.claim.Order != nil
}

// Order returns the display order for this claim.
// HasOrder should be called first to ensure this claim has a specified order before calling this method.
// This method returns an error if the claim has no specified order.
func (c *Subject) Order() (int, error) {
	if c.claim.Order == nil {
		return -1, errors.New("claim has no specified order")
	}

	return *c.claim.Order, nil
}

// Attachment returns the attachment data. Check this field if the claim type is "attachment", instead of value field.
func (c *Subject) Attachment() *Attachment {
	return &Attachment{attachment: c.claim.Attachment}
}

// Label represents localized name and locale..
type Label struct {
	label *goapicredentialschema.Label
}

// Name in a locale.
func (c *Label) Name() string {
	return c.label.Name
}

// Locale locale value.
func (c *Label) Locale() string {
	return c.label.Locale
}
