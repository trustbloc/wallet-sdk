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

	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

// Get gets an issuer's metadata by doing a lookup on its OpenID configuration endpoint.
// issuerURI is expected to be the base URL for the issuer.
func Get(issuerURI string) (*issuer.Metadata, error) {
	metadataEndpoint := issuerURI + "/.well-known/openid-configuration"

	// TODO: Implement trusted list type of mechanism? The gosec warning (correctly) warns about using a variable URL.
	response, err := http.Get(metadataEndpoint) //nolint: noctx,gosec // TODO: To be re-evaluated later
	if err != nil {
		return nil, err
	}

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received status code [%d] with body [%s] from issuer's "+
			"OpenID configuration endpoint", response.StatusCode, string(responseBytes))
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
