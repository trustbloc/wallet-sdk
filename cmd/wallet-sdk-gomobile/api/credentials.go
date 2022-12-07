/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

// VerifiableCredential typed wrapper around verifiable credentials content.
type VerifiableCredential struct {
	Content []byte
}

// NewVerifiableCredential creates a new VerifiableCredential.
func NewVerifiableCredential(content []byte) *VerifiableCredential {
	return &VerifiableCredential{
		Content: content,
	}
}

// VerifiableCredentialsArray is a wrapper around go array of VerifiableCredential to overcome limitations of gomobile.
type VerifiableCredentialsArray struct {
	credentials []*VerifiableCredential
}

// NewVerifiableCredentialsArray creates new NewVerifiableCredentialsArray.
func NewVerifiableCredentialsArray() *VerifiableCredentialsArray {
	return &VerifiableCredentialsArray{}
}

// Add adds new VerifiableCredential to underlying array.
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
