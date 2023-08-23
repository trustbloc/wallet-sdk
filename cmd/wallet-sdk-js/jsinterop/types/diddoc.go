//go:build js && wasm

/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"fmt"
	"syscall/js"

	"github.com/trustbloc/vc-go/did"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-js/jsinterop/jssupport"
)

const (
	DIDDocIDFld      = "id"
	DIDDocContentFld = "content"
)

func SerializeDIDDoc(didDoc *did.DocResolution) (map[string]interface{}, error) {
	content, err := didDoc.JSONBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize did:%w", err)
	}

	return map[string]interface{}{
		DIDDocIDFld:      didDoc.DIDDocument.ID,
		DIDDocContentFld: string(content),
	}, nil
}

func DeserializeDIDDoc(jsDIDDoc js.Value) (*did.DocResolution, error) {
	content, err := jssupport.EnsureString(jssupport.GetNamedProperty(jsDIDDoc, DIDDocContentFld))
	if err != nil {
		return nil, fmt.Errorf("failed to parse did:%w", err)
	}

	parsed, err := did.ParseDocumentResolution([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse did:%w", err)
	}

	return parsed, nil
}
