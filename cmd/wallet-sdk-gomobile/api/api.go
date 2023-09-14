/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package api defines gomobile-compatible wallet-sdk interfaces.
package api

// KeyWriter represents a type that is capable of performing operations related to key creation and storage within
// an underlying KMS.
// Deprecated: Not needed with the new DID creation pattern. This interface will be removed in a future version.
type KeyWriter interface {
	// Create creates a keyset of the given keyType and then writes it to storage.
	// The public key JWK of the newly generated keyset is returned via the JSONWebKey object.
	Create(keyType string) (*JSONWebKey, error)
}

// KeyReader represents a type that is capable of performing operations related to reading keys from an underlying KMS.
// Deprecated: Not needed with the new DID creation pattern. This interface will be removed in a future version.
type KeyReader interface {
	// ExportPubKey returns the public key associated with the given keyID as a JWK object.
	ExportPubKey(keyID string) (*JSONWebKey, error)
}

// DIDResolver defines the method required for a type to resolve DIDs.
type DIDResolver interface {
	// Resolve resolves a DID. It returns a DID document marshalled as JSON.
	Resolve(did string) ([]byte, error)
}

// Crypto defines useful Crypto operations.
type Crypto interface {
	// Sign will sign msg using a matching signature primitive from key referenced by keyID
	// returns:
	// 		signature as []byte
	//		error in case of errors
	Sign(msg []byte, keyID string) ([]byte, error)
}

// LDDocument is linked domains document.
type LDDocument struct {
	DocumentURL string `json:"documentURL,omitempty"`
	Document    string `json:"document,omitempty"` // Must be an LD document in JSON form.
	ContextURL  string `json:"contextURL,omitempty"`
}

// LDDocumentLoader is capable of loading linked domains documents.
type LDDocumentLoader interface {
	LoadDocument(url string) (*LDDocument, error)
}

// A Signer is capable of signing data.
type Signer interface {
	Sign(msg []byte) ([]byte, error)
}
