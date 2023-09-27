/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package memstorage contains a credential storage implementation using in-memory storage only.
package memstorage

import (
	"errors"
	"fmt"

	"github.com/trustbloc/vc-go/verifiable"
)

// A Provider allows for credential storage and retrieval using in-memory storage only.
type Provider struct {
	credentialStore map[string]verifiable.Credential
}

// NewProvider returns a new Provider.
func NewProvider() *Provider {
	return &Provider{credentialStore: map[string]verifiable.Credential{}}
}

// Get returns a credential with the given id. An error is returned if no credential exists with the given id.
func (p *Provider) Get(id string) (*verifiable.Credential, error) {
	credential, exists := p.credentialStore[id]
	if !exists {
		return nil, fmt.Errorf("no credential with an id of %s was found", id)
	}

	return &credential, nil
}

// GetAll returns all stored credentials.
func (p *Provider) GetAll() ([]verifiable.Credential, error) {
	credentials := make([]verifiable.Credential, len(p.credentialStore))

	var counter int

	for i := range p.credentialStore {
		credentials[counter] = p.credentialStore[i]
		counter++
	}

	return credentials, nil
}

// Add stores the given credential.
func (p *Provider) Add(vc *verifiable.Credential) error {
	if vc == nil {
		return errors.New("VC cannot be nil")
	}

	p.credentialStore[vc.Contents().ID] = *vc

	return nil
}

// Remove removes the credential with the matching id, if it exists.
func (p *Provider) Remove(id string) error {
	delete(p.credentialStore, id)

	return nil
}
