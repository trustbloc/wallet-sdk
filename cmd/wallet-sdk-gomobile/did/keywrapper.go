/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import (
	"github.com/trustbloc/kms-crypto-go/doc/jose/jwk"
	arieskms "github.com/trustbloc/kms-crypto-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// gomobileKeyWriterWrapper wraps a gomobile-compatible version of a KeyWriter and translates methods calls to
// their corresponding Go API versions.
type gomobileKeyWriterWrapper struct {
	keyWriter api.KeyWriter
}

func (g *gomobileKeyWriterWrapper) Create(keyType arieskms.KeyType) (string, *jwk.JWK, error) {
	keyHandle, err := g.keyWriter.Create(string(keyType))
	if err != nil {
		return "", nil, err
	}

	return keyHandle.ID(), keyHandle.JWK, nil
}

// gomobileKeyReaderWrapper wraps a gomobile-compatible version of a KeyReader and translates methods calls to their
// corresponding Go API versions.
type gomobileKeyReaderWrapper struct {
	keyReader api.KeyReader
}

func (g *gomobileKeyReaderWrapper) ExportPubKey(keyID string) (*jwk.JWK, error) {
	kh, err := g.keyReader.ExportPubKey(keyID)
	if err != nil {
		return nil, err
	}

	return kh.JWK, nil
}
