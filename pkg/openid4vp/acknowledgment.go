/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	"github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
)

const (
	// AccessDeniedErrorResponse is returned in "error" of Authorization Error Response when no consent is provided or
	// no credentials match found.
	AccessDeniedErrorResponse = "access_denied"
	// NoConsentErrorDescription is returned in "error_description" of Authorization Error Response when
	// no consent is provided.
	NoConsentErrorDescription = "no_consent"
	// NoMatchFoundErrorDescription is returned in "error_description" of Authorization Error Response when
	// no credentials match found.
	NoMatchFoundErrorDescription = "no_match_found"
)

// Acknowledgment holds data needed to acknowledge the verifier.
type Acknowledgment struct {
	ResponseURI        string                 `json:"response_uri"`
	State              string                 `json:"state"`
	InteractionDetails map[string]interface{} `json:"interaction_details,omitempty"`
}

// AcknowledgeVerifier sends acknowledgment to the verifier.
func (a *Acknowledgment) AcknowledgeVerifier(errStr, desc string, httpClient httpClient) error {
	// https://openid.github.io/OpenID4VP/openid-4-verifiable-presentations-wg-draft.html#section-6.2-16
	v := url.Values{}
	v.Set("error", errStr)
	v.Set("error_description", desc)
	v.Set("state", a.State)

	if a.InteractionDetails != nil {
		interactionDetailsBytes, e := json.Marshal(a.InteractionDetails)
		if e != nil {
			return fmt.Errorf("encode interaction details: %w", e)
		}

		v.Add("interaction_details", base64.StdEncoding.EncodeToString(interactionDetailsBytes))
	}

	_, err := httprequest.New(httpClient, noop.NewMetricsLogger()).Do(http.MethodPost, a.ResponseURI,
		"application/x-www-form-urlencoded", bytes.NewBufferString(v.Encode()), "", "", nil)
	if err != nil {
		return err
	}

	return nil
}
