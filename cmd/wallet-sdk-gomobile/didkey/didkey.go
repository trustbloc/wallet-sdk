/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package didkey contains a function that can be used to create did:key documents.
package didkey

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/pkg/walleterror"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/did/creator/key"
)

// Create creates a new did:key document using the given JWK.
func Create(jwk *api.JSONWebKey) (*api.DIDDocResolution, error) {
	if jwk == nil {
		return nil, wrapper.ToMobileError(walleterror.NewInvalidSDKUsageError(
			key.ErrorModule, errors.New("jwk object cannot be null/nil")))
	}

	didDocResolution, err := key.Create(jwk.JWK)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	didDocResolutionBytes, err := didDocResolution.JSONBytes()
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &api.DIDDocResolution{Content: string(didDocResolutionBytes)}, nil
}
