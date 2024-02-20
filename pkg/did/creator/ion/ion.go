/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package ion contains a did:ion longform creator implementation.
package ion

import (
	"errors"

	"github.com/trustbloc/did-go/doc/did"
	jwktype "github.com/trustbloc/kms-go/doc/jose/jwk"
	longform "github.com/trustbloc/sidetree-go/pkg/vdr/sidetreelongform"

	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// ErrorModule is the error module name used for errors relating to did:ion creation.
const ErrorModule = "DIDION"

// CreateLongForm creates a new did:ion long-form document using the given JWK.
func CreateLongForm(jwk *jwktype.JWK) (*did.DocResolution, error) {
	if jwk == nil {
		return nil, walleterror.NewInvalidSDKUsageError(
			ErrorModule, errors.New("jwk object cannot be null/nil"))
	}

	vm, err := did.NewVerificationMethodFromJWK("#"+jwk.KeyID, "JsonWebKey2020", "",
		jwk)
	if err != nil {
		return nil, err
	}

	didDocArgument := &did.Doc{
		AssertionMethod: []did.Verification{{
			VerificationMethod: *vm,
			Relationship:       did.AssertionMethod,
			Embedded:           true,
		}},
		Authentication: []did.Verification{{
			VerificationMethod: *vm,
			Relationship:       did.Authentication,
			Embedded:           true,
		}},
	}

	vdr, err := longform.New()
	if err != nil {
		return nil, err
	}

	return vdr.Create(didDocArgument)
}
