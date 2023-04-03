/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapicreator "github.com/trustbloc/wallet-sdk/pkg/did/creator"
)

// CreatorWithKeyReader is a DID creator that allows you to create DIDs using your own already-generated keys.
type CreatorWithKeyReader struct {
	goAPICreator *goapicreator.Creator
}

// NewCreatorWithKeyReader returns a new DID Creator. A Creator created with this function can be used to
// create DID documents using your own already-generated keys from the given KeyReader.
func NewCreatorWithKeyReader(keyReader api.KeyReader) (*CreatorWithKeyReader, error) {
	if keyReader == nil {
		return nil, errors.New("a KeyReader must be specified")
	}

	gomobileKeyReaderWrapper := &gomobileKeyReaderWrapper{keyReader: keyReader}

	goAPIDIDCreator, err := goapicreator.NewCreatorWithKeyReader(gomobileKeyReaderWrapper)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &CreatorWithKeyReader{
		goAPICreator: goAPIDIDCreator,
	}, nil
}

// Create creates a DID document using the given DID method.
// A verification type must be specified in opts.
// The key type specified in opts is not used for this method.
func (d *CreatorWithKeyReader) Create(method, keyID string, opts *CreateOpts) (*api.DIDDocResolution, error) {
	if keyID == "" {
		return nil, errors.New("key ID must be provided")
	}

	return create(method, keyID, d.goAPICreator, opts)
}
