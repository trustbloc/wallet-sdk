/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package did contains functionality related to DIDs.
package did

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapicreator "github.com/trustbloc/wallet-sdk/pkg/did/creator"
)

const (
	// DIDMethodKey is the name recognized by the Create method for the did:key method.
	DIDMethodKey = goapicreator.DIDMethodKey
	// Ed25519VerificationKey2018 is a supported DID verification type.
	Ed25519VerificationKey2018 = goapicreator.Ed25519VerificationKey2018
	// JSONWebKey2020 is a supported DID verification type.
	JSONWebKey2020 = goapicreator.JSONWebKey2020
)

// A Creator is used for creating DID Documents using supported DID methods.
type Creator struct {
	goAPICreator *goapicreator.Creator
}

// NewCreator returns a new DID document Creator. Any keys needed for DID creation will be generated and
// stored using the given KeyWriter.
// Deprecated: The standalone Create functions specific to each DID method should be used instead.
func NewCreator(keyWriter api.KeyWriter) (*Creator, error) {
	if keyWriter == nil {
		return nil, errors.New("a KeyWriter must be specified")
	}

	gomobileKeyWriterWrapper := &gomobileKeyWriterWrapper{keyWriter: keyWriter}

	goAPICreator, err := goapicreator.NewCreatorWithKeyWriter(gomobileKeyWriterWrapper)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &Creator{
		goAPICreator: goAPICreator,
	}, nil
}

// Create creates a DID document using the given DID method.
// A default key type and verification type will be used if they aren't specified in opts.
// Deprecated: The standalone Create functions specific to each DID method should be used instead.
func (d *Creator) Create(method string, opts *CreateOpts) (*api.DIDDocResolution, error) {
	if method == "" {
		return nil, errors.New("DID method must be provided")
	}

	return create(method, "", d.goAPICreator, opts)
}
