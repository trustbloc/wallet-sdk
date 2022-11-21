/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialquery allows querying credentials using presentation definition.
package credentialquery

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/presexch"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/piprate/json-gold/ld"
)

// Instance implements querying credentials using presentation definition.
type Instance struct {
	documentLoader ld.DocumentLoader
}

// NewInstance returns new Instance.
func NewInstance(documentLoader ld.DocumentLoader) *Instance {
	return &Instance{documentLoader: documentLoader}
}

// Query returns credentials that match PresentationDefinition.
func (c *Instance) Query(
	query *presexch.PresentationDefinition,
	contents []*verifiable.Credential,
) (*verifiable.Presentation, error) {
	return query.CreateVP(contents, c.documentLoader, verifiable.WithDisabledProofCheck(),
		verifiable.WithJSONLDDocumentLoader(c.documentLoader))
}
