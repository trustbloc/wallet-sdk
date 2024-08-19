/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// AccessDeniedErrorResponse is returned in "error" of Authorization Error Response when no consent is provided or
	// no credentials match found.
	AccessDeniedErrorResponse = "access_denied"
	// NoConsentErrorDescription is returned in "error_description" of Authorization Error Response when no consent is provided.
	NoConsentErrorDescription = "no_consent"
	// NoMatchFoundErrorDescription is returned in "error_description" of Authorization Error Response when no credentials match found.
	NoMatchFoundErrorDescription = "no_match_found"
)

// Acknowledgment holds data needed to acknowledge the verifier.
type Acknowledgment struct {
	ResponseURI string `json:"response_uri"`
	State       string `json:"state"`
}

// AcknowledgeVerifier sends acknowledgment to the verifier.
func (a *Acknowledgment) AcknowledgeVerifier(error, desc string, httpClient httpClient) error {
	// https://openid.github.io/OpenID4VP/openid-4-verifiable-presentations-wg-draft.html#section-6.2-16
	v := url.Values{}
	v.Set("error", error)
	v.Set("error_description", desc)
	v.Set("state", a.State)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, a.ResponseURI,
		bytes.NewBufferString(v.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
