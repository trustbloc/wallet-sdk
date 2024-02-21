/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package key contains a did:key creator implementation.
package key

import (
	"errors"

	"github.com/trustbloc/did-go/doc/did"
	"github.com/trustbloc/did-go/method/key"
	jwktype "github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// ErrorModule is the error module name used for errors relating to did:key creation.
const ErrorModule = "DIDKEY"

// Create creates a new did:key document using the given verification method.
func Create(jwk *jwktype.JWK) (*did.DocResolution, error) {
	if jwk == nil {
		return nil, walleterror.NewInvalidSDKUsageError(
			ErrorModule, errors.New("jwk object cannot be nil"))
	}

	var vm *did.VerificationMethod

	if jwk.Crv == "Ed25519" {
		// Workaround: when the did:key VDR creates a DID for ed25519, Ed25519VerificationKey2018 is the expected
		// verification method.
		publicKeyBytes, err := jwk.PublicKeyBytes()
		if err != nil {
			return nil, err
		}

		vm = &did.VerificationMethod{Value: publicKeyBytes, Type: "Ed25519VerificationKey2018"}
	} else {
		var err error

		vm, err = did.NewVerificationMethodFromJWK("", "JsonWebKey2020", "", jwk)
		if err != nil {
			return nil, err
		}
	}

	didDocArgument := &did.Doc{VerificationMethod: []did.VerificationMethod{*vm}}

	return key.New().Create(didDocArgument)
}
