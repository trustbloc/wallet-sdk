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

	"github.com/google/uuid"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/log/consolelogger"
	"github.com/trustbloc/wallet-sdk/pkg/models/issuer"
)

// Get gets an issuer's metadata by doing a lookup on its OpenID configuration endpoint.
// issuerURI is expected to be the base URL for the issuer.
// responseLogger is used only to log the response from the issuer's metadata endpoint if a successful status code is
// received. If not provided, then a console logger will be used.
// operationType is the operation type to be used in the log entry. It should reflect the context of where this function
// is being called from (e.g. as part of Request Credential, or as part of resolving display data, etc).
func Get(issuerURI string, responseLogger api.Logger, operationType string) (*issuer.Metadata, error) {
	if responseLogger == nil {
		responseLogger = consolelogger.NewConsoleLogger()
	}

	metadataEndpoint := issuerURI + "/.well-known/openid-configuration"

	response, err := http.Get(metadataEndpoint) //nolint: noctx,gosec
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

	logSuccess(responseLogger, operationType, responseBytes)

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

func logSuccess(logger api.Logger, operationType string, response []byte) {
	currentTime := time.Now()

	logger.Log(&api.LogEntry{
		ID:   uuid.New().String(),
		Type: api.LogTypeCredentialActivity,
		Time: &currentTime,
		Data: &api.LogData{
			Operation: operationType,
			Status:    api.LogStatusSuccess,
			Params:    map[string]interface{}{"metadata": string(response)},
		},
	})
}
