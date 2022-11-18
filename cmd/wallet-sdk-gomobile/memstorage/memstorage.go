/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package memstorage contains a credential storage implementation using in-memory storage only.
package memstorage

import (
	"encoding/json"
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapimemstorage "github.com/trustbloc/wallet-sdk/pkg/memstorage"
)

// A Provider allows for credential storage and retrieval using in-memory storage only.
type Provider struct {
	goAPIProvider  *goapimemstorage.Provider
	documentLoader ld.DocumentLoader
}

// NewProvider returns a new Provider.
// It uses a network-based JSON-LD document loader.
// TODO: Support custom document loaders so that contexts can be preloaded.
func NewProvider() *Provider {
	return &Provider{
		goAPIProvider:  goapimemstorage.NewProvider(),
		documentLoader: ld.NewDefaultDocumentLoader(http.DefaultClient),
	}
}

// Get returns a credential with the given id. An error is returned if no credential exists with the given id.
func (p *Provider) Get(id string) (*api.JSONObject, error) {
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
func (p *Provider) GetAll() (*api.JSONArray, error) {
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
func (p *Provider) Add(vc *api.JSONObject) error {
	credential, err := verifiable.ParseCredential(vc.Data,
		verifiable.WithJSONLDDocumentLoader(p.documentLoader),
		verifiable.WithDisabledProofCheck())
	if err != nil {
		return err
	}

	return p.goAPIProvider.Add(credential)
}

// Remove removes the credential with the matching id, if it exists.
func (p *Provider) Remove(id string) error {
	return p.goAPIProvider.Remove(id)
}
