/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package didion contains a function that can be used to create did:ion documents.
package didion

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	didion "github.com/trustbloc/wallet-sdk/pkg/did/creator/ion"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// CreateLongForm creates a new did:ion long-form document using the given JWK.
func CreateLongForm(jwk *api.JSONWebKey) (*api.DIDDocResolution, error) {
	if jwk == nil {
		return nil, wrapper.ToMobileError(walleterror.NewInvalidSDKUsageError(
			didion.ErrorModule, errors.New("jwk object cannot be null/nil")))
	}

	didDocResolution, err := didion.CreateLongForm(jwk.JWK)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	didDocResolutionBytes, err := didDocResolution.JSONBytes()
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &api.DIDDocResolution{Content: string(didDocResolutionBytes)}, nil
}
