/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package api defines gomobile-compatible wallet-sdk interfaces.
package api

// KeyHandle represents a key with associated metadata.
type KeyHandle struct {
	Key     []byte `json:"key,omitempty"`
	KeyType string `json:"keyType,omitempty"`
}

// A KeyHandleWriter represents a type that is capable of performing certain operations related to writing key data.
type KeyHandleWriter interface {
	// Create a new key/keyset/key handle of type keyType
	// Some key types may require additional attributes described in `opts`. // TODO: Format of opts to be determined.
	// Returns a key handle, which contains both the key ID and actual private key bytes
	Create(keyType, opts string) (*KeyHandle, error)
	// Rotate a key referenced by keyID and return a new handle of a keyset including old key and
	// new key of type keyType. It also returns the updated keyID as the first return value
	// Some key types may require additional attributes described in `opts` // TODO: Format of opts to be determined.
	// Returns: a key handle, which contains both the new key ID and actual private key bytes
	Rotate(keyType, keyID string, opts string) (*KeyHandle, error)
	// Import will import privKey into the KMS storage for the given keyType then returns the new key id and
	// the newly persisted KeyHandle.
	// privKey possible types are: *ecdsa.PrivateKey and ed25519.PrivateKey.
	// TODO: Determine how these restrictions work.
	// kt possible types are signing key types only (ECDSA keys or Ed25519)
	// opts allows setting the keysetID of the imported key using WithKeyID() option. If the ID is already used,
	// then an error is returned. // TODO: Format of opts to be determined.
	// Returns: a key handle, which contains both the new key ID and actual private key bytes.
	// An error/exception will be returned/thrown if there is an import failure (key empty, invalid, doesn't match
	// keyType, unsupported keyType or storing of key failed)
	// TODO: Consider renaming this method to avoid keyword collision and automatic renaming to _import in Java
	Import(privateKey *KeyHandle, keyType, opts string) (*KeyHandle, error)
}

// A KeyHandleReader represents a type that is capable of performing certain operations related to reading key data.
type KeyHandleReader interface {
	// GetKeyHandle return the key handle for the given keyID
	// Returns:
	//  - The private key handle
	//  - Error if failure
	GetKeyHandle(keyID string) (*KeyHandle, error)
	// Export will fetch a key referenced by id then gets its public key in raw bytes and returns it.
	// The key must be an asymmetric key.
	// Returns:
	//  - A key handle, which contains both the key type and public key bytes
	//  - Error if it fails to export the public key bytes
	Export(keyID string) (*KeyHandle, error)
}

// CreateDIDOpts represents the various options for the DIDCreator.Create method.
type CreateDIDOpts struct {
	KeyID string
}

// DIDCreator defines the method required for a type to create DID documents.
type DIDCreator interface {
	// Create creates a new DID Document.
	// It returns a DID Document Resolution.
	Create(method string, createDIDOpts *CreateDIDOpts) ([]byte, error)
}

// DIDResolver defines the method required for a type to resolve DIDs.
type DIDResolver interface {
	// Resolve resolves a did.
	// It returns a DID Document Resolution.
	Resolve(did string) ([]byte, error)
}

// A CredentialReader is capable of reading VCs from some underlying storage mechanism.
type CredentialReader interface {
	// Get retrieves a VC.
	Get(id string) ([]byte, error)
	// GetAll retrieves all VCs.
	GetAll() ([]byte, error)
}

// A CredentialWriter is capable of writing VCs to some underlying storage mechanism.
type CredentialWriter interface {
	// Remove removes a VC.
	Remove(id string) error
	// Add adds a VC.
	Add(vc []byte) error
}

// Crypto defines useful Crypto operations.
// TODO: Define more precisely the input and output formats.
type Crypto interface {
	// Sign will sign msg using a matching signature primitive in kh key handle of a private key
	// returns:
	// 		signature as []byte
	//		error in case of errors
	Sign(msg []byte, kh *KeyHandle) ([]byte, error)
	// Verify will verify a signature for the given msg using a matching signature primitive in kh key handle of
	// a public key
	// returns:
	// 		error in case of errors or nil if signature verification was successful
	Verify(signature, msg []byte, kh *KeyHandle) error
}

// ActivityLog defines logging functionality.
type ActivityLog interface {
	// Log logs an activity.
	Log(message string)
}
