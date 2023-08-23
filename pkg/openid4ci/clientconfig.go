/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

import (
	"errors"
	"net/http"
	"time"

	"github.com/trustbloc/kms-crypto-go/spi/kms"
	"github.com/trustbloc/vc-go/dataintegrity/suite/ecdsa2019"

	"github.com/piprate/json-gold/ld"
	diddoc "github.com/trustbloc/vc-go/did"
	"github.com/trustbloc/vc-go/spi/vdr"

	noopactivitylogger "github.com/trustbloc/wallet-sdk/pkg/activitylogger/noop"
	"github.com/trustbloc/wallet-sdk/pkg/api"
	noopmetricslogger "github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
)

type didResolverWrapper struct {
	didResolver api.DIDResolver
}

func (d *didResolverWrapper) Resolve(did string, _ ...vdr.DIDMethodOption) (*diddoc.DocResolution, error) {
	return d.didResolver.Resolve(did)
}

// Header represents an HTTP header.
type Header struct {
	Name  string
	Value string
}

// ClientConfig contains the various parameters for an OpenID4CI Interaction.
type ClientConfig struct {
	DIDResolver                      api.DIDResolver
	ActivityLogger                   api.ActivityLogger // If not specified, then activities won't be logged.
	MetricsLogger                    api.MetricsLogger  // If not specified, then metrics events won't be logged.
	DisableVCProofChecks             bool
	DocumentLoader                   ld.DocumentLoader // If not specified, then a network-based loader will be used.
	NetworkDocumentLoaderHTTPTimeout *time.Duration    // Only used if the default network-based loader is used.
	HTTPClient                       *http.Client
	// If both of the below fields are set, then VC data integrity proof checks for JSON-LD VCs will be enabled.
	Verifier ecdsa2019.Verifier
	KMS      kms.KeyManager
}

func validateRequiredParameters(config *ClientConfig) error {
	if config == nil {
		return errors.New("no client config provided")
	}

	if config.DIDResolver == nil {
		return errors.New("no DID resolver provided")
	}

	return nil
}

func setDefaults(config *ClientConfig) {
	if config.ActivityLogger == nil {
		config.ActivityLogger = noopactivitylogger.NewActivityLogger()
	}

	if config.MetricsLogger == nil {
		config.MetricsLogger = noopmetricslogger.NewMetricsLogger()
	}

	if config.DocumentLoader == nil {
		httpClient := &http.Client{}

		if config.NetworkDocumentLoaderHTTPTimeout != nil {
			httpClient.Timeout = *config.NetworkDocumentLoaderHTTPTimeout
		} else {
			httpClient.Timeout = api.DefaultHTTPTimeout
		}

		config.DocumentLoader = ld.NewDefaultDocumentLoader(httpClient)
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: api.DefaultHTTPTimeout}
	}
}
