/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package issuermetadata contains a function for fetching issuer metadata.
package issuermetadata

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"

	"github.com/trustbloc/wallet-sdk/pkg/api"

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

const fetchIssuerMetadataViaGETReqEventText = "Fetch issuer metadata via an HTTP GET request to %s"

// Get gets an issuer's metadata by doing a lookup on its OpenID configuration endpoint.
// issuerURI is expected to be the base URL for the issuer.
func Get(issuerURI string, metricsLogger api.MetricsLogger, parentEvent string) (*issuer.Metadata, error) {
	if metricsLogger == nil {
		metricsLogger = noop.NewMetricsLogger()
	}

	metadataEndpoint := issuerURI + "/.well-known/openid-credential-issuer"

	timeStartHTTPRequest := time.Now()

	response, err := http.Get(metadataEndpoint) //nolint: noctx,gosec
	if err != nil {
		return nil, err
	}

	err = metricsLogger.Log(&api.MetricsEvent{
		Event:       fmt.Sprintf(fetchIssuerMetadataViaGETReqEventText, metadataEndpoint),
		ParentEvent: parentEvent,
		Duration:    time.Since(timeStartHTTPRequest),
	})
	if err != nil {
		return nil, err
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code [%d] with body [%s] from issuer's "+
			"OpenID credential issuer endpoint", response.StatusCode, string(responseBytes))
	}

	defer func() {
		errClose := response.Body.Close()
		if errClose != nil {
			println(fmt.Sprintf("failed to close response body: %s", errClose.Error()))
		}
	}()

	var metadata issuer.Metadata

	err = json.Unmarshal(responseBytes, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response from the issuer's "+
			"OpenID configuration endpoint: %w", err)
	}

	return &metadata, err
}
