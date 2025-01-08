/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package verifiable contains functionality related to Verifiable Credentials.
package verifiable

import (
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

const claimsTypesFieldName = "type"

// Credential represents a Verifiable Credential per the VC Data Model spec:
// https://www.w3.org/TR/vc-data-model/.
// It wraps the VC type from aries-framework-go and provides gomobile-compatible methods.
type Credential struct {
	VC *verifiable.Credential // Will be skipped in the gomobile bindings due to using an incompatible type
}

// NewCredential creates a new Credential.
// This function is only used internally in wallet-sdk-gomobile and is not available in the bindings due to it using
// unsupported types.
// To create a VC from a serialized format via the bindings, see the ParseCredential method.
func NewCredential(vc *verifiable.Credential) *Credential {
	return &Credential{
		VC: vc,
	}
}

// ID returns this VC's ID.
func (v *Credential) ID() string {
	return v.VC.Contents().ID
}

// Name returns this VC's name.
// If this VC doesn't provide a name, or the name isn't a string, then an empty string is returned.
func (v *Credential) Name() string {
	rawName := v.VC.CustomField("name")
	if rawName == nil {
		return ""
	}

	nameAsString, ok := rawName.(string)
	if !ok {
		return ""
	}

	return nameAsString
}

// IssuerID returns the ID of this VC's issuer.
// While the ID is typically going to be a DID, the Verifiable Credential spec does not mandate this.
func (v *Credential) IssuerID() string {
	return v.VC.Contents().Issuer.ID
}

// Types returns the types of this VC. At a minimum, one of the types will be "VerifiableCredential".
// There may be additional more specific credential types as well.
func (v *Credential) Types() *api.StringArray {
	return &api.StringArray{Strings: v.VC.Contents().Types}
}

// ClaimTypes returns the types specified in the claims of this VC.
// It first checks the selective disclosure claims for the types - if found, they are returned.
// Otherwise, the non-selective disclosure claims are checked and returned if found.
// When checking the non-selective disclosure claims, only the first subject is checked.
// If not found in either place, then nil is returned.
func (v *Credential) ClaimTypes() *api.StringArray {
	types := v.claimTypesFromSelectiveDisclosures()
	if types != nil {
		return types
	}

	types = v.claimTypesFromCredentialSubject()
	if types != nil {
		return types
	}

	return nil
}

func (v *Credential) claimTypesFromSelectiveDisclosures() *api.StringArray {
	if len(v.VC.SDJWTDisclosures()) > 0 {
		for _, disclosure := range v.VC.SDJWTDisclosures() {
			if disclosure.Name == claimsTypesFieldName {
				return rawTypesToStringArray(disclosure.Value)
			}
		}
	}

	return nil
}

func (v *Credential) claimTypesFromCredentialSubject() *api.StringArray {
	credentialSubjects := v.VC.Contents().Subject

	if len(credentialSubjects) == 0 {
		return nil
	}

	rawTypes, exists := credentialSubjects[0].CustomFields[claimsTypesFieldName]
	if !exists {
		return nil
	}

	return rawTypesToStringArray(rawTypes)
}

func rawTypesToStringArray(rawTypes interface{}) *api.StringArray {
	typesAsInterfaceArray, ok := rawTypes.([]interface{}) // This will be the type if the VC was parsed (unmarshalled)
	if ok {
		types := make([]string, len(typesAsInterfaceArray))
		for i := range typesAsInterfaceArray {
			types[i], ok = typesAsInterfaceArray[i].(string)
			if !ok {
				return nil
			}
		}

		return &api.StringArray{Strings: types}
	}

	typesAsStringArray, ok := rawTypes.([]string)
	if ok {
		return &api.StringArray{Strings: typesAsStringArray}
	}

	typeAsString, ok := rawTypes.(string)
	if ok {
		return &api.StringArray{Strings: []string{typeAsString}}
	}

	return nil
}

// IssuanceDate returns this VC's issuance date as a Unix timestamp.
func (v *Credential) IssuanceDate() int64 {
	if v.VC.Contents().Issued == nil {
		return 0
	}

	return v.VC.Contents().Issued.Unix()
}

// HasExpirationDate returns whether this VC has an expiration date.
func (v *Credential) HasExpirationDate() bool {
	return v.VC.Contents().Expired != nil
}

// ExpirationDate returns this VC's expiration date as a Unix timestamp.
// HasExpirationDate should be called first to ensure this VC has an expiration date before calling this method.
// This method returns an error if the VC has no expiration date.
func (v *Credential) ExpirationDate() int64 {
	if v.VC.Contents().Expired == nil {
		return 0
	}

	return v.VC.Contents().Expired.Unix()
}

// Serialize returns a JSON representation of this VC.
func (v *Credential) Serialize() (string, error) {
	marshalledVC, err := v.VC.MarshalJSON()
	if err != nil {
		return "", err
	}

	return string(marshalledVC), nil
}

// CredentialsArray represents an array of Credentials.
// Since arrays and slices are not compatible with gomobile, this type acts as a wrapper around a Go array of VCs.
type CredentialsArray struct {
	credentials []*Credential
}

// NewCredentialsArray creates new CredentialsArray.
func NewCredentialsArray() *CredentialsArray {
	return &CredentialsArray{}
}

// Add adds new VC to underlying array.
func (a *CredentialsArray) Add(cred *Credential) {
	a.credentials = append(a.credentials, cred)
}

// Length returns the number of Credentials contained within this CredentialsArray.
func (a *CredentialsArray) Length() int {
	return len(a.credentials)
}

// AtIndex returns the Credential at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (a *CredentialsArray) AtIndex(index int) *Credential {
	maxIndex := len(a.credentials) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return a.credentials[index]
}

// CredentialsArrayV2 represents an array of Credentials with associated array of config IDs.
// Each config ID maps to CredentialConfigurationSupported in the issuer's metadata.
type CredentialsArrayV2 struct {
	credentials []*Credential
	configIDs   []string
}

// NewCredentialsArrayV2 creates a new CredentialsArrayV2.
func NewCredentialsArrayV2() *CredentialsArrayV2 {
	return &CredentialsArrayV2{}
}

// Add adds a new Credential with associated credential config ID.
func (a *CredentialsArrayV2) Add(credential *Credential, configID string) {
	a.credentials = append(a.credentials, credential)
	a.configIDs = append(a.configIDs, configID)
}

// Length returns the number of Credentials contained within this CredentialsArrayV2.
func (a *CredentialsArrayV2) Length() int {
	return len(a.credentials)
}

// AtIndex returns the Credential at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (a *CredentialsArrayV2) AtIndex(index int) *Credential {
	if index < 0 || index >= len(a.credentials) {
		return nil
	}

	return a.credentials[index]
}

// ConfigIDAtIndex returns the config ID for the Credential at the given index.
// If the index is out of bounds, it returns an empty string.
func (a *CredentialsArrayV2) ConfigIDAtIndex(index int) string {
	if index < 0 || index >= len(a.credentials) {
		return ""
	}

	return a.configIDs[index]
}
