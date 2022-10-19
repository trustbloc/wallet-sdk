/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package storage contains a storage implementation.
package storage

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

// A Provider allows for credential storage and retrieval.
type Provider struct {
	credentialStore map[string][]byte
}

// NewProvider returns a new Provider.
func NewProvider() *Provider {
	return &Provider{credentialStore: map[string][]byte{}}
}

// Get returns a credential with the given id. An error is returned if no credential exists with the given id.
func (p *Provider) Get(id string) ([]byte, error) {
	credential, exists := p.credentialStore[id]
	if !exists {
		return nil, fmt.Errorf("no credential with an id of %s was found", id)
	}

	return credential, nil
}

// GetAll returns all stored credentials.
func (p *Provider) GetAll() ([]byte, error) {
	credentials := make([][]byte, len(p.credentialStore))

	var counter int

	for _, credential := range p.credentialStore {
		credentials[counter] = credential
		counter++
	}

	credentialsAsJSONArray, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}

	return credentialsAsJSONArray, nil
}

// Remove removes the credential with the matching id, if it exists.
func (p *Provider) Remove(id string) error {
	delete(p.credentialStore, id)

	return nil
}

// Add stores the given credential.
func (p *Provider) Add(vc []byte) error {
	vcID := gjson.GetBytes(vc, "id")

	p.credentialStore[vcID.String()] = vc

	return nil
}
