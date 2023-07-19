/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package issuermetadata contains a function for fetching issuer metadata.
package issuermetadata

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	"github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

const fetchIssuerMetadataViaGETReqEventText = "Fetch issuer metadata via an HTTP GET request to %s"

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Get gets an issuer's metadata by doing a lookup on its OpenID configuration endpoint.
// issuerURI is expected to be the base URL for the issuer.
func Get(issuerURI string, httpClient httpClient, metricsLogger api.MetricsLogger, parentEvent string,
) (*issuer.Metadata, error) {
	if metricsLogger == nil {
		metricsLogger = noop.NewMetricsLogger()
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: api.DefaultHTTPTimeout}
	}

	metadataEndpoint := issuerURI + "/.well-known/openid-credential-issuer"

	responseBytes, err := httprequest.New(httpClient, metricsLogger).Do(
		http.MethodGet, metadataEndpoint, "", nil,
		fmt.Sprintf(fetchIssuerMetadataViaGETReqEventText, metadataEndpoint), parentEvent, nil)
	if err != nil {
		return nil, fmt.Errorf("openid configuration endpoint: %w", err)
	}

	var metadata issuer.Metadata

	err = json.Unmarshal(responseBytes, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from the issuer's "+
			"OpenID configuration endpoint: %w", err)
	}

	return &metadata, err
}
