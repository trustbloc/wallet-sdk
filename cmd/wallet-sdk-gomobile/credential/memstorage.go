/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credential contains an in-memory credential storage implementation.
// It also contains a type that can be used to query for credentials using a presentation definition.
package credential

import (
	"net/http"

	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapimemstorage "github.com/trustbloc/wallet-sdk/pkg/memstorage"
)

// A DB allows for credential storage and retrieval using in-memory storage only.
type DB struct {
	goAPIProvider  *goapimemstorage.Provider
	documentLoader ld.DocumentLoader
}

// NewInMemoryDB returns a new in-memory credential DB.
// It uses a network-based JSON-LD document loader.
// TODO: https://github.com/trustbloc/wallet-sdk/issues/160 Support custom document
// loaders so that contexts can be preloaded.
func NewInMemoryDB() *DB {
	return &DB{
		goAPIProvider:  goapimemstorage.NewProvider(),
		documentLoader: ld.NewDefaultDocumentLoader(http.DefaultClient),
	}
}

// Get returns a credential with the given id. An error is returned if no credential exists with the given id.
func (p *DB) Get(id string) (*api.VerifiableCredential, error) {
	vc, err := p.goAPIProvider.Get(id)
	if err != nil {
		return nil, err
	}

	return api.NewVerifiableCredential(vc), nil
}

// GetAll returns all stored credentials.
func (p *DB) GetAll() (*api.VerifiableCredentialsArray, error) {
	vcs, err := p.goAPIProvider.GetAll()
	if err != nil {
		return nil, err
	}

	gomobileVCs := api.NewVerifiableCredentialsArray()

	for i := range vcs {
		gomobileVCs.Add(api.NewVerifiableCredential(&vcs[i]))
	}

	return gomobileVCs, nil
}

// Add stores the given credential.
func (p *DB) Add(vc *api.VerifiableCredential) error {
	return p.goAPIProvider.Add(vc.VC)
}

// Remove removes the credential with the matching id, if it exists.
func (p *DB) Remove(id string) error {
	return p.goAPIProvider.Remove(id)
}
