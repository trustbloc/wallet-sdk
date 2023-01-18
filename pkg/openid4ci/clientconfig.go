/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"errors"

	"github.com/hyperledger/aries-framework-go/pkg/doc/util/didsignjwt"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/walleterror"

	diddoc "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdr"
)

type didResolverWrapper struct {
	didResolver api.DIDResolver
}

func (d *didResolverWrapper) Resolve(did string, _ ...vdr.DIDMethodOption) (*diddoc.DocResolution, error) {
	return d.didResolver.Resolve(did)
}

// ClientConfig contains the various required parameters for an OpenID4CI Interaction.
// TODO: https://github.com/trustbloc/wallet-sdk/issues/163 refactor to instead require a key ID and a signer.
type ClientConfig struct {
	UserDID        string
	ClientID       string
	SignerProvider didsignjwt.SignerGetter
	DIDResolver    api.DIDResolver
}

func validateClientConfig(config *ClientConfig) error {
	if config == nil {
		return walleterror.NewValidationError(
			module,
			NoClientConfigProvidedCode,
			NoClientConfigProvidedError,
			errors.New("no client config provided"))
	}

	if config.UserDID == "" {
		return walleterror.NewValidationError(
			module,
			ClientConfigNoUserDidProvidedCode,
			ClientConfigNoUserDidProvidedError,
			errors.New("no user DID provided"))
	}

	if config.ClientID == "" {
		return walleterror.NewValidationError(
			module,
			ClientConfigNoClientIDProvidedCode,
			ClientConfigNoClientIDProvidedError,
			errors.New("no client ID provided"))
	}

	if config.SignerProvider == nil {
		return walleterror.NewValidationError(
			module,
			ClientConfigNoSignerProviderProvidedCode,
			ClientConfigNoSignerProviderProvidedError,
			errors.New("no signer provider provided"))
	}

	if config.DIDResolver == nil {
		return walleterror.NewValidationError(
			module,
			ClientConfigNoDIDResolverProvidedCode,
			ClientConfigNoDIDResolverProvidedError,
			errors.New("no DID resolver provided"))
	}

	return nil
}
