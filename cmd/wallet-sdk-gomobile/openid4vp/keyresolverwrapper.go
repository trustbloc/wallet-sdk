/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// gomobileVDRKeyResolverAdapter wraps a gomobile-compatible version of a DIDResolver and translates methods calls to
// their corresponding Go API versions.
type gomobileVDRKeyResolverAdapter struct {
	didResolver api.DIDResolver
}

func (a *gomobileVDRKeyResolverAdapter) Resolve(didID string) (*did.DocResolution, error) {
	docBytes, err := a.didResolver.Resolve(didID)
	if err != nil {
		return nil, err
	}

	doc, err := did.ParseDocumentResolution(docBytes)
	if err != nil {
		return nil, fmt.Errorf("document resolution parsing failed: %w", err)
	}

	return doc, nil
}
