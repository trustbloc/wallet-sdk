/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package api defines wallet-sdk APIs.
package api

import (
	"time"

	"github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/kms-go/doc/jose"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/jwt"
	"github.com/trustbloc/vc-go/verifiable"
)

// DefaultHTTPTimeout is the default timeout used across Wallet-SDK for HTTP calls.
const DefaultHTTPTimeout = time.Second * 30

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
	KeyType          kms.KeyType // Not used for DID:ion's update and recovery key types (they're hardcoded)
	MetricsLogger    MetricsLogger
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
}

// JWTSigner defines interface for JWT signing operation.
type JWTSigner interface {
	GetKeyID() string
	SignJWT(sigParams jwt.SignParameters, data []byte) ([]byte, error)
	CreateJWTHeaders(sigParams jwt.SignParameters) (jose.Headers, error)
}

// JSONWebKeySet represents a JWK Set object.
// It uses the JWK type from aries-framework-go.
type JSONWebKeySet struct {
	JWKs []jwk.JWK `json:"keys"`
}
