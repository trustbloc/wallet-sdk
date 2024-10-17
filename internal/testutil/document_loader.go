/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package testutil implements common test tasks.
package testutil

import (
	_ "embed" //nolint:gci // required for go:embed
	"testing"

	"github.com/stretchr/testify/require"
	ldcontext "github.com/trustbloc/did-go/doc/ld/context"
	lddocloader "github.com/trustbloc/did-go/doc/ld/documentloader"
	mockldstore "github.com/trustbloc/did-go/doc/ld/mock"
	ldstore "github.com/trustbloc/did-go/doc/ld/store"
)

var (
	//go:embed contexts/credentials-examples_v1.jsonld
	credentialExamples []byte
	//go:embed contexts/examples_v1.jsonld
	vcExamples []byte
	//go:embed contexts/examples_v2.jsonld
	vcExamplesV2 []byte
	//go:embed contexts/odrl.jsonld
	odrl []byte
	//go:embed contexts/citizenship_v1.jsonld
	citizenship []byte
	//go:embed contexts/lds-jws2020-v1.jsonld
	jws2020 []byte
)

type mockLDStoreProvider struct {
	ContextStore        ldstore.ContextStore
	RemoteProviderStore ldstore.RemoteProviderStore
}

func (p *mockLDStoreProvider) JSONLDContextStore() ldstore.ContextStore {
	return p.ContextStore
}

func (p *mockLDStoreProvider) JSONLDRemoteProviderStore() ldstore.RemoteProviderStore {
	return p.RemoteProviderStore
}

// DocumentLoader returns a document loader with preloaded test contexts.
func DocumentLoader(t *testing.T, extraContexts ...ldcontext.Document) *lddocloader.DocumentLoader {
	t.Helper()

	ldStore := &mockLDStoreProvider{
		ContextStore:        mockldstore.NewMockContextStore(),
		RemoteProviderStore: mockldstore.NewMockRemoteProviderStore(),
	}

	testContexts := []ldcontext.Document{
		{
			URL:     "https://www.w3.org/2018/credentials/examples/v1",
			Content: credentialExamples,
		},
		{
			URL:     "https://trustbloc.github.io/context/vc/examples-v1.jsonld",
			Content: vcExamples,
		},
		{
			URL:     "https://www.w3.org/ns/credentials/examples/v2",
			Content: vcExamplesV2,
		},
		{
			URL:     "https://www.w3.org/ns/odrl.jsonld",
			Content: odrl,
		},
		{
			URL:         "https://w3id.org/citizenship/v1",
			DocumentURL: "https://w3c-ccg.github.io/citizenship-vocab/contexts/citizenship-v1.jsonld",
			Content:     citizenship,
		},
		{
			URL:     "https://w3c-ccg.github.io/lds-jws2020/contexts/lds-jws2020-v1.json",
			Content: jws2020,
		},
	}

	loader, err := lddocloader.NewDocumentLoader(ldStore,
		lddocloader.WithExtraContexts(
			append(testContexts, extraContexts...)...,
		),
	)
	require.NoError(t, err)

	return loader
}
