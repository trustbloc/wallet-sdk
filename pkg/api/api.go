/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package api defines wallet-sdk APIs.
package api

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
)

// KeyHandle represents a key with associated metadata.
type KeyHandle struct {
	Key     []byte `json:"key,omitempty"` // Raw bytes
	KeyType string `json:"keyType,omitempty"`
}

// KeyHandleWriter defines key handler writer APIs.
type KeyHandleWriter interface {
	// Create a new key/keyset/key handle for the type kt
	// Some key types may require additional attributes described in `opts`
	// Returns:
	//  - keyID of the handle
	//  - handle instance (to private key)
	//  - error if failure
	Create(kt kms.KeyType, opts ...kms.KeyOpts) (string, interface{}, error)
	// Rotate a key referenced by keyID and return a new handle of a keyset including old key and
	// new key with type kt. It also returns the updated keyID as the first return value
	// Some key types may require additional attributes described in `opts`
	// Returns:
	//  - new KeyID
	//  - handle instance (to private key)
	//  - error if failure
	Rotate(kt kms.KeyType, keyID string, opts ...kms.KeyOpts) (string, interface{}, error)
	// Import will import privateKey into the KMS storage for the given keyType then returns the new key id and
	// the newly persisted Handle.
	// 'privateKey' possible types are: *ecdsa.PrivateKey and ed25519.PrivateKey
	// 'kt' possible types are signing key types only (ECDSA keys or Ed25519)
	// 'opts' allows setting the keysetID of the imported key using WithKeyID() option. If the ID is already used,
	// then an error is returned.
	// Returns:
	//  - keyID of the handle
	//  - handle instance (to private key)
	//  - error if import failure (key empty, invalid, doesn't match keyType, unsupported keyType or storing key failed)
	Import(privateKey interface{}, kt kms.KeyType, opts ...kms.PrivateKeyOpts) (string, interface{}, error)
}

// KeyHandleReader defines key handler reader APIs.
type KeyHandleReader interface {
	// GetKeyHandle returns a key handle stored under the given keyID.
	GetKeyHandle(keyID string) (*KeyHandle, error)
	// Export is not yet fully defined. TODO: Define this
	Export(keyID string) ([]byte, error)
}

// CreateDIDOpts represents the various options for the DIDCreator.Create method.
type CreateDIDOpts struct {
	KeyID string
}

// DIDCreator defines the method required for a type to create DID documents.
type DIDCreator interface {
	// Create creates a new DID Document using the given method.
	Create(method string, createDIDOpts *CreateDIDOpts) (*did.DocResolution, error)
}

// DIDResolver defines DID resolution APIs.
type DIDResolver interface {
	// Resolve resolves a did.
	Resolve(did string) (*did.DocResolution, error)
}

// CredentialReader defines credential reader APIs.
type CredentialReader interface {
	// Get retrieves a VC.
	Get(id string) (*verifiable.Credential, error)
	// GetAll retrieves all VCs.
	GetAll() ([]*verifiable.Credential, error)
}

// CredentialWriter defines credential write APIs.
type CredentialWriter interface {
	// Remove removes a VC.
	Remove(id string) error
	// Add adds a VC.
	Add(vc *verifiable.Credential) error
}

// Crypto defines various crypto operations that may be used with wallet-sdk APIs.
type Crypto interface {
	// Sign will sign msg using a matching signature primitive in kh key handle of a private key
	// returns:
	// 		signature in []byte
	//		error in case of errors
	Sign(msg []byte, kh interface{}) ([]byte, error)
	// Verify will verify a signature for the given msg using a matching signature primitive in kh key handle of
	// a public key
	// returns:
	// 		error in case of errors or nil if signature verification was successful
	Verify(signature, msg []byte, kh interface{}) error
}

// ActivityLog defines activity log related APIs.
type ActivityLog interface {
	// Log logs an activity.
	Log(message string)
}
