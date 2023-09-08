/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package ion contains a did:ion longform creator implementation.
package ion

import (
	"fmt"

	"github.com/trustbloc/did-go/doc/did"
	longform "github.com/trustbloc/did-go/method/sidetreelongform"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	"github.com/trustbloc/kms-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// Creator is used for creating did:ion longform DID Documents.
type Creator struct {
	kw  api.KeyWriter
	vdr *longform.VDR
}

// NewCreator returns a new did:ion longform document Creator.
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
