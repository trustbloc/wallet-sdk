/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package ld contains functionality for dealing with linked domains.
package ld

import (
	"encoding/json"
	"net/http"

	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// DocLoader represents a type that can help with linked domains.
type DocLoader struct {
	documentLoader ld.DocumentLoader
}

// NewDocLoader returns a new DocLoader instance.
func NewDocLoader() *DocLoader {
	return &DocLoader{
		documentLoader: ld.NewDefaultDocumentLoader(http.DefaultClient),
	}
}

// LoadDocument load linked document by url.
func (l *DocLoader) LoadDocument(u string) (*api.LDDocument, error) {
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
