/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package linkeddomains contains functionality for dealing with linked domains.
package linkeddomains

import (
	"encoding/json"
	"net/http"

	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// DocumentLoader represents a type that can help with linked domains.
type DocumentLoader struct {
	documentLoader ld.DocumentLoader
}

// NewDocumentLoader returns a new DocumentLoader instance.
func NewDocumentLoader() *DocumentLoader {
	return &DocumentLoader{
		documentLoader: ld.NewDefaultDocumentLoader(http.DefaultClient),
	}
}

// LoadDocument load linked document by url.
func (l *DocumentLoader) LoadDocument(u string) (*api.LDDocument, error) {
	doc, err := l.documentLoader.LoadDocument(u)
	if err != nil {
		return nil, err
	}

	wrappedDoc := &api.LDDocument{
		DocumentURL: doc.DocumentURL,
		ContextURL:  doc.ContextURL,
	}

	wrappedDoc.Document, err = json.Marshal(doc.Document)
	if err != nil {
		return nil, err
	}

	return wrappedDoc, nil
}
