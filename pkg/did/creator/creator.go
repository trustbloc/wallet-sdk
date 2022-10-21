/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package creator defines a type that can be used to create DIDs using various supported DID methods.
package creator

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	didkeycreator "github.com/trustbloc/wallet-sdk/pkg/did/key"
)

const (
	// DIDMethodKey is the name used by the Create method for the did:key method.
	DIDMethodKey = "key"

	didMethodPrefix            = "did:"
	ed25519VerificationKey2018 = "Ed25519VerificationKey2018"
)

// DIDCreator is used for creating DID Documents using supported DID methods.
type DIDCreator struct {
	keyHandleReader api.KeyHandleReader
}

// NewDIDCreator returns a new DIDCreator.
// keyHandleReader is optional. If provided, then it will be used to fetch a key (based on the KeyID in the options)
// in the Create method. If not provided, then a new key will be generated and used for DID creation.
func NewDIDCreator(keyHandleReader api.KeyHandleReader) *DIDCreator {
	return &DIDCreator{
		keyHandleReader: keyHandleReader,
	}
}

// Create creates a DID document for the given DID method.
// The usage of createDIDOpts depends on the DID method you're using.
func (d *DIDCreator) Create(method string, createDIDOpts *api.CreateDIDOpts) (*did.DocResolution, error) {
	if method == DIDMethodKey || method == fmt.Sprintf("%s%s", didMethodPrefix, DIDMethodKey) {
		return d.createKeyDIDDoc(createDIDOpts)
	}

	return nil, fmt.Errorf("DID method %s not supported", method)
}

func (d *DIDCreator) createKeyDIDDoc(createDIDOpts *api.CreateDIDOpts) (*did.DocResolution, error) {
	var key []byte

	var keyType string

	if createDIDOpts.KeyID == "" {
		var err error

		key, _, err = ed25519.GenerateKey(nil)
		if err != nil {
			return nil, err
		}

		keyType = ed25519VerificationKey2018
	} else {
		if d.keyHandleReader == nil {
			return nil, errors.New("key ID specified but no key handle reader set up")
		}

		keyHandle, err := d.keyHandleReader.GetKeyHandle(createDIDOpts.KeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get key handle: %w", err)
		}

		key = keyHandle.Key
		keyType = keyHandle.KeyType
	}

	return didkeycreator.NewDIDKeyCreator().Create(key, keyType)
}
