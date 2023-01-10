/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credential contains an in-memory credential storage implementation.
// It also contains a type that can be used to query for credentials using a presentation definition.
package credential

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"
	"github.com/trustbloc/wallet-sdk/pkg/common"

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
		documentLoader: ld.NewDefaultDocumentLoader(common.DefaultHTTPClient()),
	}
}

// Get returns a credential with the given id. An error is returned if no credential exists with the given id.
func (p *DB) Get(id string) (*api.JSONObject, error) {
	vc, err := p.goAPIProvider.Get(id)
	if err != nil {
		return nil, err
	}

	vcBytes, err := json.Marshal(vc)
	if err != nil {
		return nil, err
	}

	return &api.JSONObject{Data: vcBytes}, nil
}

// GetAll returns all stored credentials.
func (p *DB) GetAll() (*api.JSONArray, error) {
	vcs, err := p.goAPIProvider.GetAll()
	if err != nil {
		return nil, err
	}

	vcsBytes, err := json.Marshal(vcs)
	if err != nil {
		return nil, err
	}

	return &api.JSONArray{Data: vcsBytes}, nil
}

// Add stores the given credential.
func (p *DB) Add(vc *api.JSONObject) error {
	credential, err := verifiable.ParseCredential(vc.Data,
		verifiable.WithJSONLDDocumentLoader(p.documentLoader),
		verifiable.WithDisabledProofCheck())
	if err != nil {
		return err
	}

	return p.goAPIProvider.Add(credential)
}

// Remove removes the credential with the matching id, if it exists.
func (p *DB) Remove(id string) error {
	return p.goAPIProvider.Remove(id)
}
