/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"encoding/json"
	"fmt"

	"github.com/piprate/json-gold/ld"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// gomobileDocumentLoaderWrapper wraps a gomobile-compatible version of a LDDocumentLoader and translates
// methods calls to their corresponding Go API versions.
type gomobileDocumentLoaderWrapper struct {
	ldDocumentLoader api.LDDocumentLoader
}

func (l *gomobileDocumentLoaderWrapper) LoadDocument(u string) (*ld.RemoteDocument, error) {
	doc, err := l.ldDocumentLoader.LoadDocument(u)
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
