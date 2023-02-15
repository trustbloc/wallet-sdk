/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package api defines wallet-sdk APIs.
package api

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose"
	"github.com/hyperledger/aries-framework-go/pkg/doc/jose/jwk"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
)

// KeyWriter represents a type that is capable of performing operations related to key creation and storage within
// an underlying KMS.
type KeyWriter interface {
	// Create creates a keyset of the given keyType and then writes it to storage.
	// The keyID and public key JWK for the newly generated keyset are returned.
	Create(keyType kms.KeyType) (string, *jwk.JWK, error)
}

// KeyReader represents a type that is capable of performing operations related to reading keys from an underlying KMS.
type KeyReader interface {
	// ExportPubKey returns the public key associated with the given keyID as a JWK.
	ExportPubKey(keyID string) (*jwk.JWK, error)
}

// CreateDIDOpts represents the various options for the DIDCreator.Create method.
type CreateDIDOpts struct {
	KeyID            string
	VerificationType string
	KeyType          kms.KeyType
}

// DIDCreator defines the method required for a type to create DID documents.
type DIDCreator interface {
	// Create creates a new DID Document using the given method.
	Create(method string, createDIDOpts *CreateDIDOpts) (*did.DocResolution, error)
}

// DIDResolver defines DID resolution APIs.
type DIDResolver interface {
	// Resolve resolves a DID.
	Resolve(did string) (*did.DocResolution, error)
}

// A CredentialReader is capable of reading VCs from some underlying storage mechanism.
type CredentialReader interface {
	// Get retrieves a VC.
	Get(id string) (*verifiable.Credential, error)
	// GetAll retrieves all VCs.
	GetAll() ([]*verifiable.Credential, error)
}

// A CredentialWriter is capable of writing VCs to some underlying storage mechanism.
type CredentialWriter interface {
	// Add adds a VC.
	Add(vc *verifiable.Credential) error
	// Remove removes a VC.
	Remove(id string) error
}

// Crypto defines various crypto operations that may be used with wallet-sdk APIs.
type Crypto interface {
	// Sign will sign msg using a matching signature primitive from key referenced by keyID
	Sign(msg []byte, keyID string) ([]byte, error)
	// Verify will verify a signature for the given msg using a matching signature primitive from key referenced by keyID
	Verify(signature, msg []byte, keyID string) error
}

// JWTSigner defines interface for JWT signing operation.
type JWTSigner interface {
	GetKeyID() string
	Sign(data []byte) ([]byte, error)
	Headers() jose.Headers
}
