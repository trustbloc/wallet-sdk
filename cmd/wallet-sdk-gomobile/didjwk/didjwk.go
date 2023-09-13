/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package didjwk contains a function that can be used to create did:jwk documents.
package didjwk

import (
	"errors"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	didjwk "github.com/trustbloc/wallet-sdk/pkg/did/creator/jwk"
	"github.com/trustbloc/wallet-sdk/pkg/did/creator/key"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// Create creates a new did:jwk document using the given JWK.
func Create(jwk *api.JSONWebKey) (*api.DIDDocResolution, error) {
	if jwk == nil {
		return nil, wrapper.ToMobileError(walleterror.NewInvalidSDKUsageError(
			key.ErrorModule, errors.New("jwk object cannot be null/nil")))
	}

	didDocResolution, err := didjwk.Create(jwk.JWK)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	didDocResolutionBytes, err := didDocResolution.JSONBytes()
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &api.DIDDocResolution{Content: string(didDocResolutionBytes)}, nil
}
