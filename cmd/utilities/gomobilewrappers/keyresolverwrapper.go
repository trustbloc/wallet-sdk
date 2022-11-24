/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gomobilewrappers

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// VDRKeyResolverWrapper wraps a gomobile-compatible version of a DIDResolver and translates methods calls to
// their corresponding Go API versions.
type VDRKeyResolverWrapper struct {
	DIDResolver api.DIDResolver
}

// Resolve wraps Resolve of  api.DIDResolver.
func (a *VDRKeyResolverWrapper) Resolve(didID string) (*did.DocResolution, error) {
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
