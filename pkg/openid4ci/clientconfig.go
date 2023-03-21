/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"errors"
	"net/http"

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

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientConfig contains the various parameters for an OpenID4CI Interaction.
type ClientConfig struct {
	ClientID             string
	DIDResolver          api.DIDResolver
	ActivityLogger       api.ActivityLogger // If not specified, then activities won't be logged.
	MetricsLogger        api.MetricsLogger  // If not specified, then metrics events won't be logged.
	DisableVCProofChecks bool
	HTTPClient           httpClient
}

func validateClientConfig(config *ClientConfig) error {
	if config == nil {
		return walleterror.NewValidationError(
			module,
			NoClientConfigProvidedCode,
			NoClientConfigProvidedError,
			errors.New("no client config provided"))
	}

	if config.ClientID == "" {
		return walleterror.NewValidationError(
			module,
			ClientConfigNoClientIDProvidedCode,
			ClientConfigNoClientIDProvidedError,
			errors.New("no client ID provided"))
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
