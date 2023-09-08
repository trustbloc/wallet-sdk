/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package models contains models.
package models

import (
	"github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
)

// VerificationKey holds either a JWK or a raw public key.
type VerificationKey struct {
	JSONWebKey *jwk.JWK
	Raw        []byte
}

// VerificationMethod represents a DID verification method.
type VerificationMethod struct {
	ID   string
	Type string
	Key  VerificationKey
}

// VerificationMethodOption provides an optional public key for initializing a verification method.
type VerificationMethodOption func(vm *VerificationMethod)

// NewVerificationMethod creates VerificationMethod with optional public key.
func NewVerificationMethod(keyID, vmType string, opt ...VerificationMethodOption) *VerificationMethod {
	vm := &VerificationMethod{
		ID:   keyID,
		Type: vmType,
	}

	for _, option := range opt {
		option(vm)
	}

	return vm
}

// VerificationMethodFromDoc initializes a VerificationMethod from the DID Doc Verification Method type.
func VerificationMethodFromDoc(docVM *did.VerificationMethod) *VerificationMethod {
	jsonWebKey := docVM.JSONWebKey()
	if jsonWebKey != nil {
		return &VerificationMethod{
			ID:   docVM.ID,
			Type: docVM.Type,
			Key:  VerificationKey{JSONWebKey: jsonWebKey},
		}
	}

	return &VerificationMethod{
		ID:   docVM.ID,
		Type: docVM.Type,
		Key:  VerificationKey{Raw: docVM.Value},
	}
}

// WithRawKey initializes the VerificationMethod with public key in raw []byte format.
func WithRawKey(rawKey []byte) VerificationMethodOption {
	return func(vm *VerificationMethod) {
		vm.Key.Raw = rawKey
	}
}

// WithJWK initializes the VerificationMethod with public key in JWK format.
func WithJWK(pubKey *jwk.JWK) VerificationMethodOption {
	return func(vm *VerificationMethod) {
		vm.Key.JSONWebKey = pubKey
	}
}
