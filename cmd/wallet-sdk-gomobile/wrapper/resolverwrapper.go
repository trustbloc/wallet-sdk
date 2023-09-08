/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	"fmt"

	"github.com/trustbloc/did-go/doc/did"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// VDRResolverWrapper wraps a gomobile-compatible version of a DIDResolver and translates methods calls to
// their corresponding Go API versions.
type VDRResolverWrapper struct {
	DIDResolver api.DIDResolver
}

// Resolve wraps Resolve of  api.DIDResolver.
func (a *VDRResolverWrapper) Resolve(didID string) (*did.DocResolution, error) {
	docBytes, err := a.DIDResolver.Resolve(didID)
	if err != nil {
		return nil, err
	}

	doc, err := did.ParseDocumentResolution(docBytes)
	if err != nil {
		return nil, fmt.Errorf("document resolution parsing failed: %w", err)
	}

	return doc, nil
}
