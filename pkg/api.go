/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pkg

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/hyperledger/aries-framework-go/pkg/kms"
)

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
	// Import will import privKey into the KMS storage for the given keyType then returns the new key id and
	// the newly persisted Handle.
	// 'privKey' possible types are: *ecdsa.PrivateKey and ed25519.PrivateKey
	// 'kt' possible types are signing key types only (ECDSA keys or Ed25519)
	// 'opts' allows setting the keysetID of the imported key using WithKeyID() option. If the ID is already used,
	// then an error is returned.
	// Returns:
	//  - keyID of the handle
	//  - handle instance (to private key)
	//  - error if import failure (key empty, invalid, doesn't match keyType, unsupported keyType or storing key failed)
	Import(privKey interface{}, kt kms.KeyType, opts ...kms.PrivateKeyOpts) (string, interface{}, error)
}

type KeyHandleReader interface {
	// GetKeyHandle key handle for the given keyID
	// Returns:
	//  - handle instance (to private key)
	//  - error if failure
	GetKeyHandle(keyID string) (interface{}, error)
	// Export will fetch a key referenced by id then gets its public key in raw bytes and returns it.
	// The key must be an asymmetric key.
	// Returns:
	//  - marshalled public key []byte
	//  - error if it fails to export the public key bytes
	Export(keyID string) ([]byte, kms.KeyType, error)
}

type DIDCreator interface {
	// Create creates a new DID Document.
	Create(didDocument *did.Doc) (*did.DocResolution, error)
}

type DIDResolver interface {
	// Resolve resolves a did.
	Resolve(did string) (*did.DocResolution, error)
}

type CredentialReader interface {
	// Get retrieves a VC.
	Get(id string) (*verifiable.Credential, error)
	// GetAll retrieves all VCs.
	GetAll() ([]*verifiable.Credential, error)
}

type CredentialWriter interface {
	// Remove removes a VC.
	Remove(id string) error
	// Add adds a VC.
	Add(vc *verifiable.Credential) error
}

type ActivityLog interface {
	// Log logs an activity.
	Log(message string)
}
