/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package vdr contains functionality for doing VDR operations.
package vdr

// VerifiableDataRegistry performs VDR operations.
type VerifiableDataRegistry struct{}

// Create does something // TODO: Complete this.
func (v *VerifiableDataRegistry) Create(didDocument []byte) ([]byte, error) {
	return didDocument, nil
}

// Resolve does something // TODO: Complete this.
func (v *VerifiableDataRegistry) Resolve(did string) ([]byte, error) {
	return []byte("DID document"), nil
}
