/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package creator contains a DID document creator that can be used to create DIDs using various supported DID methods.
package creator

import (
	"errors"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	didkeycreator "github.com/trustbloc/wallet-sdk/pkg/did/creator/key"
)

const (
	// DIDMethodKey is the name recognized by the Create method for the did:key method.
	DIDMethodKey = "key"
	// Ed25519VerificationKey2018 is a supported DID verification type.
	Ed25519VerificationKey2018 = "Ed25519VerificationKey2018"
)

// Creator is used for creating DID Documents using supported DID methods.
type Creator struct {
	keyWriter api.KeyWriter
	keyReader api.KeyReader
}

// NewCreatorWithKeyWriter returns a new DID document Creator. A Creator created with this function will automatically
// generate keys for you when creating new DID documents. Those keys will be generated and stored using the given
// KeyWriter. See the Create method for more information.
func NewCreatorWithKeyWriter(keyWriter api.KeyWriter) (*Creator, error) {
	if keyWriter == nil {
		return nil, errors.New("a KeyWriter must be specified")
	}

	return &Creator{
		keyWriter: keyWriter,
	}, nil
}

// NewCreatorWithKeyReader returns a new DID document Creator. A Creator created with this function can be used to
// create DID documents using your own already-generated keys from the given KeyReader.
// At least one of keyHandleCreator and keyReader must be provided. See the Create method for more information.
func NewCreatorWithKeyReader(keyReader api.KeyReader) (*Creator, error) {
	if keyReader == nil {
		return nil, errors.New("a KeyReader must be specified")
	}

	return &Creator{
		keyReader: keyReader,
	}, nil
}

// Create creates a DID document using the given DID method.
// The usage of createDIDOpts depends on the DID method you're using.
//
// For creating did:key documents, there are two relevant options that can be set in the createDIDOpts object: KeyID and
// VerificationType.
//
//	If the Creator was created using the NewCreatorWithKeyWriter function, then both of those options are ignored.
//	ED25519 key will be generated and saved automatically, and the verification type will automatically be set to
//	Ed25519VerificationKey2018. TODO (#51): Support more key types. ED25519 is the chosen default for now.
//	If the Creator was created using the NewCreatorWithKeyReader function, then you must specify the KeyID and also
//	the VerificationType in the createDIDOpts object to use for the creation of the DID document.
func (d *Creator) Create(method string, createDIDOpts *api.CreateDIDOpts) (*did.DocResolution, error) {
	if method == DIDMethodKey {
		return d.createDIDKeyDoc(createDIDOpts)
	}

	return nil, fmt.Errorf("DID method %s not supported", method)
}

func (d *Creator) createDIDKeyDoc(createDIDOpts *api.CreateDIDOpts) (*did.DocResolution, error) {
	var key []byte

	var verificationType string

	var err error

	if d.keyReader == nil { // Generate a key and set the verification type in behalf of the caller.
		_, key, err = d.keyWriter.Create(arieskms.ED25519Type)
		if err != nil {
			return nil, err
		}

		verificationType = Ed25519VerificationKey2018
	} else { // Use the caller's chosen key and verification type.
		if createDIDOpts.VerificationType == "" {
			return nil, errors.New("no verification type specified")
		}

		key, err = d.keyReader.ExportPubKey(createDIDOpts.KeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get key handle: %w", err)
		}

		verificationType = createDIDOpts.VerificationType
	}

	return didkeycreator.NewCreator().Create(key, verificationType)
}
