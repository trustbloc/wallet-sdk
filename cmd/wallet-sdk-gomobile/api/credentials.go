/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/wallet-sdk/pkg/common"
)

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

// IssuerID returns the ID of this VC's issuer.
// While the ID is typically going to be a DID, the Verifiable Credential spec does not mandate this.
func (v *VerifiableCredential) IssuerID() (string, error) {
	vc, err := verifiable.ParseCredential(v.Content, verifiable.WithDisabledProofCheck(),
		verifiable.WithJSONLDDocumentLoader(ld.NewDefaultDocumentLoader(common.DefaultHTTPClient())))
	if err != nil {
		return "", err
	}

	return vc.Issuer.ID, nil
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
