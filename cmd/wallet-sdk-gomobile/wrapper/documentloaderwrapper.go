/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

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
func (l *DocumentLoaderWrapper) LoadDocument(url string) (*ld.RemoteDocument, error) {
	doc, err := l.DocumentLoader.LoadDocument(url)
	if err != nil {
		return nil, err
	}

	wrappedDoc := &ld.RemoteDocument{
		DocumentURL: doc.DocumentURL,
		ContextURL:  doc.ContextURL,
	}

	err = json.Unmarshal([]byte(doc.Document), &wrappedDoc.Document)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal ld document bytes: %w", err)
	}

	return wrappedDoc, nil
}
