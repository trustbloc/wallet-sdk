/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package gomobilewrappers contains wrappers that wraps mobile interfaces to go interfaces.
package gomobilewrappers

import (
	"encoding/json"
	"fmt"

	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// DocumentLoaderWrapper wraps a gomobile-compatible version of a LDDocumentLoader and translates
// methods calls to their corresponding Go API versions.
type DocumentLoaderWrapper struct {
	DocumentLoader api.LDDocumentLoader
}

// LoadDocument wraps LoadDocument of api.LDDocumentLoader.
func (l *DocumentLoaderWrapper) LoadDocument(u string) (*ld.RemoteDocument, error) {
	doc, err := l.DocumentLoader.LoadDocument(u)
	if err != nil {
		return nil, err
	}

	wrappedDoc := &ld.RemoteDocument{
		DocumentURL: doc.DocumentURL,
		ContextURL:  doc.ContextURL,
	}

	err = json.Unmarshal(doc.Document, &wrappedDoc.Document)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal ld document bytes: %w", err)
	}

	return wrappedDoc, nil
}