/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
)

// VerifiableCredential represents a Verifiable Credential per the VC Data Model spec:
// https://www.w3.org/TR/vc-data-model/.
// It wraps the VC type from aries-framework-go and provides gomobile-compatible methods.
type VerifiableCredential struct {
	VC *verifiable.Credential // Will be skipped in the gomobile bindings due to using an incompatible type
}

// NewVerifiableCredential creates a new VerifiableCredential.
// This function is only used internally in wallet-sdk-gomobile and is not available in the bindings due to it using
// unsupported types.
// To create a VC from a serialized format via the bindings, see the ParseVC method.
// This function will be skipped in the gomobile bindings due to using an incompatible type.
func NewVerifiableCredential(vc *verifiable.Credential) *VerifiableCredential {
	return &VerifiableCredential{
		VC: vc,
	}
}

// IssuerID returns the ID of this VC's issuer.
// While the ID is typically going to be a DID, the Verifiable Credential spec does not mandate this.
func (v *VerifiableCredential) IssuerID() string {
	return v.VC.Issuer.ID
}

// Serialize returns a JSON representation of this VC.
func (v *VerifiableCredential) Serialize() (string, error) {
	marshalledVC, err := v.VC.MarshalJSON()
	if err != nil {
		return "", err
	}

	return string(marshalledVC), nil
}

// VerifiableCredentialsArray represents an array of VerifiableCredentials.
// Since arrays and slices are not compatible with gomobile, this type acts as a wrapper around a Go array of VCs.
type VerifiableCredentialsArray struct {
	credentials []*VerifiableCredential
}

// NewVerifiableCredentialsArray creates new NewVerifiableCredentialsArray.
func NewVerifiableCredentialsArray() *VerifiableCredentialsArray {
	return &VerifiableCredentialsArray{}
}

// Add adds new VC to underlying array.
func (a *VerifiableCredentialsArray) Add(cred *VerifiableCredential) {
	a.credentials = append(a.credentials, cred)
}

// Length returns length of underlying array.
func (a *VerifiableCredentialsArray) Length() int {
	return len(a.credentials)
}

// AtIndex returns element from underlying array by index.
func (a *VerifiableCredentialsArray) AtIndex(index int) *VerifiableCredential {
	return a.credentials[index]
}
