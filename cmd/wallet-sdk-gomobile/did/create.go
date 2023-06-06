/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import (
	"errors"

	arieskms "github.com/hyperledger/aries-framework-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	goapicreator "github.com/trustbloc/wallet-sdk/pkg/did/creator"
)

func create(method, keyID string, goAPICreator *goapicreator.Creator, opts *CreateOpts) (*api.DIDDocResolution, error) {
	if method == "" {
		return nil, errors.New("DID method must be provided")
	}

	if opts == nil {
		opts = NewCreateOpts()
	}

	goAPIOpts := &goapi.CreateDIDOpts{
		VerificationType: opts.verificationType,
		KeyType:          arieskms.KeyType(opts.keyType),
		MetricsLogger:    &wrapper.MobileMetricsLoggerWrapper{MobileAPIMetricsLogger: opts.metricsLogger},
		KeyID:            keyID,
	}

	didDocResolution, err := goAPICreator.Create(method, goAPIOpts)
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	didDocResolutionBytes, err := didDocResolution.JSONBytes()
	if err != nil {
		return nil, wrapper.ToMobileError(err)
	}

	return &api.DIDDocResolution{Content: string(didDocResolutionBytes)}, nil
}
