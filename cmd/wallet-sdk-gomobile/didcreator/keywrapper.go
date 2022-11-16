/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didcreator

import (
	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// gomobileKeyWriterWrapper wraps a gomobile-compatible version of a KeyWriter and translates methods calls to
// their corresponding Go API versions.
type gomobileKeyWriterWrapper struct {
	keyWriter api.KeyWriter
}

func (g *gomobileKeyWriterWrapper) Create(keyType arieskms.KeyType) (string, []byte, error) {
	keyHandle, err := g.keyWriter.Create(string(keyType))
	if err != nil {
		return "", nil, err
	}

	return keyHandle.KeyID, keyHandle.PubKey, nil
}

// gomobileKeyReaderWrapper wraps a gomobile-compatible version of a KeyReader and translates methods calls to their
// corresponding Go API versions.
type gomobileKeyReaderWrapper struct {
	keyReader api.KeyReader
}

func (g *gomobileKeyReaderWrapper) ExportPubKey(keyID string) ([]byte, error) {
	return g.keyReader.ExportPubKey(keyID)
}

func (g *gomobileKeyReaderWrapper) GetSignAlgorithm(keyID string) (string, error) {
	return g.keyReader.GetSignAlgorithm(keyID)
}
