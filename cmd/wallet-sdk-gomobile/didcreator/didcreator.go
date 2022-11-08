/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package didcreator contains a DID document creator that can be used to create DIDs using various supported DID
// methods.
package didcreator

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	goapicreator "github.com/trustbloc/wallet-sdk/pkg/did/creator"
)

const (
	// DIDMethodKey is the name recognized by the Create method for the did:key method.
	DIDMethodKey = goapicreator.DIDMethodKey
	// Ed25519VerificationKey2018 is a supported DID verification type.
	Ed25519VerificationKey2018 = goapicreator.Ed25519VerificationKey2018
)

// Creator is used for creating DID Documents using supported DID methods.
type Creator struct {
	goAPICreator *goapicreator.Creator
}

// NewCreatorWithKeyWriter returns a new DID document Creator. A Creator created with this function will automatically
// generate keys for you when creating new DID documents. Those keys will be generated and stored using the given
// KeyWriter. See the Create method for more information.
func NewCreatorWithKeyWriter(keyWriter api.KeyWriter) (*Creator, error) {
	gomobileKeyWriterWrapper := &gomobileKeyWriterWrapper{keyWriter: keyWriter}

	goAPICreator, err := goapicreator.NewCreatorWithKeyWriter(gomobileKeyWriterWrapper)
	if err != nil {
		return nil, err
	}

	return &Creator{
		goAPICreator: goAPICreator,
	}, nil
}

// NewCreatorWithKeyReader returns a new DID document Creator. A Creator created with this function can be used to
// create DID documents using your own already-generated keys from the given KeyReader.
// At least one of keyHandleCreator and keyReader must be provided. See the Create method for more information.
func NewCreatorWithKeyReader(keyReader api.KeyReader) (*Creator, error) {
	gomobileKeyReaderWrapper := &gomobileKeyReaderWrapper{keyReader: keyReader}

	goAPIDIDCreator, err := goapicreator.NewCreatorWithKeyReader(gomobileKeyReaderWrapper)
	if err != nil {
		return nil, err
	}

	return &Creator{
		goAPICreator: goAPIDIDCreator,
	}, nil
}

// Create creates a DID document using the given DID method.
// The usage of createDIDOpts depends on the DID method you're using.
//
// For creating did:key documents, there are two relevant options that can be set in the createDIDOpts object: KeyID and
// VerificationType.
//
//	If the Creator was created using the NewCreatorWithKeyWriter function, then both of those options are ignored.
//	An ED25519 key will be generated and saved automatically, and the verification type will automatically be set to
//	Ed25519VerificationKey2018. TODO (#51): Support more key types. ED25519 is the chosen default for now.
//	If the Creator was created using the NewCreatorWithKeyReader function, then you must specify the KeyID and also
//	the VerificationType in the createDIDOpts object to use for the creation of the DID document.
func (d *Creator) Create(method string, createDIDOpts *api.CreateDIDOpts) ([]byte, error) {
	goAPIOpts := convertToGoAPIOpts(createDIDOpts)

	didDocResolution, err := d.goAPICreator.Create(method, goAPIOpts)
	if err != nil {
		return nil, err
	}

	return didDocResolution.JSONBytes()
}

func convertToGoAPIOpts(createDIDOpts *api.CreateDIDOpts) *goapi.CreateDIDOpts {
	return &goapi.CreateDIDOpts{KeyID: createDIDOpts.KeyID}
}
