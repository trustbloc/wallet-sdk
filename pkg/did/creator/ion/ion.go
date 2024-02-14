/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package ion contains a did:ion longform creator implementation.
package ion

import (
	"errors"
	"fmt"

	"github.com/trustbloc/did-go/doc/did"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	jwktype "github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/spi/kms"
	longform "github.com/trustbloc/sidetree-go/pkg/vdr/sidetreelongform"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// ErrorModule is the error module name used for errors relating to did:ion creation.
const ErrorModule = "DIDION"

// Creator is used for creating did:ion longform DID Documents.
type Creator struct {
	kw  api.KeyWriter
	vdr *longform.VDR
}

// NewCreator returns a new did:ion longform document Creator.
// Deprecated: The standalone Create function should be used instead.
func NewCreator(kw api.KeyWriter) (*Creator, error) {
	v, err := longform.New()
	if err != nil {
		return nil, err
	}

	return &Creator{
		vdr: v,
		kw:  kw,
	}, nil
}

// Create creates a new did:ion longform document using the given Verification Method.
// Deprecated: The standalone Create function should be used instead.
func (d *Creator) Create(vm *did.VerificationMethod) (*did.DocResolution, error) {
	updatePK, err := d.makeKey()
	if err != nil {
		return nil, fmt.Errorf("failed to create update key: %w", err)
	}

	recoveryPK, err := d.makeKey()
	if err != nil {
		return nil, fmt.Errorf("failed to create recovery key: %w", err)
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

	return d.vdr.Create(
		didDocArgument,
		vdrapi.WithOption(longform.UpdatePublicKeyOpt, updatePK),
		vdrapi.WithOption(longform.RecoveryPublicKeyOpt, recoveryPK),
	)
}

func (d *Creator) makeKey() (interface{}, error) {
	_, pkJWK, err := d.kw.Create(kms.ECDSAP256TypeIEEEP1363)
	if err != nil {
		return nil, err
	}

	return pkJWK.Key, nil
}

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
