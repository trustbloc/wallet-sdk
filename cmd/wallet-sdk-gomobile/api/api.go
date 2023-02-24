/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package api defines gomobile-compatible wallet-sdk interfaces.
package api

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk"
)

// JSONWebKey holds a public key with associated metadata, in JWK format.
type JSONWebKey struct {
	JWK *jwk.JWK `json:"jwk,omitempty"`
}

// KeyWriter represents a type that is capable of performing operations related to key creation and storage within
// an underlying KMS.
type KeyWriter interface {
	// Create creates a keyset of the given keyType and then writes it to storage.
	// The public key JWK of the newly generated keyset is returned via the JSONWebKey object.
	Create(keyType string) (*JSONWebKey, error)
}

// KeyReader represents a type that is capable of performing operations related to reading keys from an underlying KMS.
type KeyReader interface {
	// ExportPubKey returns the public key associated with the given keyID as a JWK object.
	ExportPubKey(keyID string) (*JSONWebKey, error)
}

// CreateDIDOpts represents the various options for the DIDCreator.Create method.
type CreateDIDOpts struct {
	KeyID            string
	VerificationType string
	KeyType          string
}

// DIDCreator defines the method required for a type to create DID documents.
type DIDCreator interface {
	// Create creates a new DID Document using the given method.
	Create(method string, createDIDOpts *CreateDIDOpts) (*DIDDocResolution, error)
}

// DIDResolver defines the method required for a type to resolve DIDs.
type DIDResolver interface {
	// Resolve resolves a DID. It returns a DID document marshalled as JSON.
	Resolve(did string) ([]byte, error)
}

// A CredentialReader is capable of reading VCs from some underlying storage mechanism.
type CredentialReader interface {
	// Get retrieves a VC.
	Get(id string) (*VerifiableCredential, error)
	// GetAll retrieves all VCs.
	GetAll() (*VerifiableCredentialsArray, error)
}

// A CredentialWriter is capable of writing VCs to some underlying storage mechanism.
type CredentialWriter interface {
	// Add adds a VC.
	Add(vc *VerifiableCredential) error
	// Remove removes a VC.
	Remove(id string) error
}

// Crypto defines useful Crypto operations.
type Crypto interface {
	// Sign will sign msg using a matching signature primitive from key referenced by keyID
	// returns:
	// 		signature as []byte
	//		error in case of errors
	Sign(msg []byte, keyID string) ([]byte, error)
	// Verify will verify a signature for the given msg using a matching signature primitive from key referenced by keyID
	// returns:
	// 		error in case of errors or nil if signature verification was successful
	Verify(signature, msg []byte, keyID string) error
}

// LDDocument is linked domains document.
type LDDocument struct {
	DocumentURL string `json:"documentUrl,omitempty"`
	// bytes of json document.
	Document   []byte `json:"document,omitempty"`
	ContextURL string `json:"contextUrl,omitempty"`
}

// LDDocumentLoader is capable of loading linked domains documents.
type LDDocumentLoader interface {
	LoadDocument(u string) (*LDDocument, error)
}

// A Signer is capable of signing data.
type Signer interface {
	Sign(msg []byte) ([]byte, error)
}
