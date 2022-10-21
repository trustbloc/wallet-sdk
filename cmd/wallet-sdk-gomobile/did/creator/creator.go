/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package creator is a gomobile-compatible version of github.com/trustbloc/wallet-sdk/pkg/did/creator.
// It takes care of the necessary wrapping/conversion between the Go SDK and the gomobile sdk
package creator

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	goapididcreator "github.com/trustbloc/wallet-sdk/pkg/did/creator"
)

// DIDMethodKey is the name used by the Create method for the did:key method.
const DIDMethodKey = goapididcreator.DIDMethodKey

// DIDCreator is used for creating DID Documents using supported DID methods (currently only did:key).
type DIDCreator struct {
	goAPIDIDCreator *goapididcreator.DIDCreator
}

// gomobileKHRWrapper wraps a gomobile-compatible version of a KeyHandleReader and translates methods calls to their
// respective Go API versions.
type gomobileKHRWrapper struct {
	keyHandleReader api.KeyHandleReader
}

func (g *gomobileKHRWrapper) GetKeyHandle(keyID string) (*goapi.KeyHandle, error) {
	keyHandle, err := g.keyHandleReader.GetKeyHandle(keyID)
	if err != nil {
		return nil, err
	}

	return &goapi.KeyHandle{
		Key:     keyHandle.Key,
		KeyType: keyHandle.KeyType,
	}, nil
}

func (g *gomobileKHRWrapper) Export(keyID string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

// NewDIDCreator returns a new DIDCreator.
// keyHandleReader is optional. If provided, then it will be used to fetch a key (based on the KeyID in the options)
// in the Create method. If not provided, then a new key will be generated and used for DID creation.
func NewDIDCreator(keyHandleReader api.KeyHandleReader) *DIDCreator {
	goAPIKHRWrapper := &gomobileKHRWrapper{keyHandleReader: keyHandleReader}

	return &DIDCreator{goAPIDIDCreator: goapididcreator.NewDIDCreator(goAPIKHRWrapper)}
}

// Create creates a DID document for the given DID method.
// The usage of createDIDOpts depends on the DID method you're using.
func (d *DIDCreator) Create(method string, createDIDOpts *api.CreateDIDOpts) ([]byte, error) {
	goAPIOpts := convertToGoAPIOpts(createDIDOpts)

	didDocResolution, err := d.goAPIDIDCreator.Create(method, goAPIOpts)
	if err != nil {
		return nil, err
	}

	return didDocResolution.JSONBytes()
}

func convertToGoAPIOpts(createDIDOpts *api.CreateDIDOpts) *goapi.CreateDIDOpts {
	return &goapi.CreateDIDOpts{KeyID: createDIDOpts.KeyID}
}
