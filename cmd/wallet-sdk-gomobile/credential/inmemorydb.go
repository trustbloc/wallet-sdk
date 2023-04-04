/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credential contains an in-memory credential storage implementation.
// It also contains a type that can be used to query for credentials using a presentation definition.
package credential

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	goapimemstorage "github.com/trustbloc/wallet-sdk/pkg/memstorage"
)

// A InMemoryDB allows for credential storage and retrieval using in-memory storage only.
type InMemoryDB struct {
	goAPIProvider *goapimemstorage.Provider
}

// NewInMemoryDB returns a new InMemoryDB.
func NewInMemoryDB() *InMemoryDB {
	return &InMemoryDB{
		goAPIProvider: goapimemstorage.NewProvider(),
	}
}

// Get returns a credential with the given id. An error is returned if no credential exists with the given id.
func (p *InMemoryDB) Get(id string) (*verifiable.Credential, error) {
	vc, err := p.goAPIProvider.Get(id)
	if err != nil {
		return nil, err
	}

	return verifiable.NewCredential(vc), nil
}

// GetAll returns all stored credentials.
func (p *InMemoryDB) GetAll() (*verifiable.CredentialsArray, error) {
	vcs, err := p.goAPIProvider.GetAll()
	if err != nil {
		return nil, err
	}

	gomobileVCs := verifiable.NewCredentialsArray()

	for i := range vcs {
		gomobileVCs.Add(verifiable.NewCredential(&vcs[i]))
	}

	return gomobileVCs, nil
}

// Add stores the given credential.
func (p *InMemoryDB) Add(vc *verifiable.Credential) error {
	return p.goAPIProvider.Add(vc.VC)
}

// Remove removes the credential with the matching id, if it exists.
func (p *InMemoryDB) Remove(id string) error {
	return p.goAPIProvider.Remove(id)
}
