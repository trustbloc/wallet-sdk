/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package httprequest is utility package to simplify work with http requests.
package httprequest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Request is utility class used to do http requests.
type Request struct {
	httpClient    httpClient
	metricsLogger api.MetricsLogger
}

// New returns new Request.
func New(httpClient httpClient, metricsLogger api.MetricsLogger) *Request {
	return &Request{
		httpClient:    httpClient,
		metricsLogger: metricsLogger,
	}
}

// Do executes request in background context and read response body.
func (r *Request) Do(method, endpointURL, contentType string, body io.Reader,
	event, parentEvent string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, endpointURL, body)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	timeStartHTTPRequest := time.Now()

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	err = r.metricsLogger.Log(&api.MetricsEvent{
		Event:       event,
		ParentEvent: parentEvent,
		Duration:    time.Since(timeStartHTTPRequest),
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		errClose := resp.Body.Close()
		if errClose != nil {
			println(fmt.Sprintf("failed to close response body: %s", errClose.Error()))
		}
	}()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"expected status code %d but got status code %d with response body %s instead",
			http.StatusOK, resp.StatusCode, respBytes)
	}

	return respBytes, nil
}
