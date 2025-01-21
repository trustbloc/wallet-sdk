/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package httprequest is utility package to simplify work with http requests.
package httprequest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
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

// Do executes the request in the background context and reads the response body.
// If a status other than 200 is received from the endpoint, then errorResponseHandler is called to generate the
// error that gets returned. If errorResponseHandler is nil, then a generic error response handler will be used.
func (r *Request) Do(method, endpointURL, contentType string, body io.Reader,
	event, parentEvent string, errorResponseHandler func(statusCode int, responseBody []byte) error,
) ([]byte, error) {
	return r.DoContext(context.Background(), method, endpointURL, contentType,
		nil, body, event, parentEvent, nil, errorResponseHandler)
}

//nolint:gochecknoglobals
var defaultAcceptableStatuses = []int{http.StatusOK}

// DoContext is the same as Do, but also accept context and headers.
//
//nolint:gocyclo
func (r *Request) DoContext(ctx context.Context, method, endpointURL, contentType string,
	additionalHeaders http.Header, body io.Reader, event, parentEvent string, acceptableStatuses []int,
	errorResponseHandler func(statusCode int, responseBody []byte) error,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpointURL, body)
	if err != nil {
		return nil, err
	}

	for header, values := range additionalHeaders {
		for _, value := range values {
			req.Header.Add(header, value)
		}
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
			fmt.Printf("failed to close response body: %s\n", errClose.Error())
		}
	}()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	statuses := acceptableStatuses
	if statuses == nil {
		statuses = defaultAcceptableStatuses
	}

	if !slices.Contains(statuses, resp.StatusCode) {
		if errorResponseHandler == nil {
			errorResponseHandler = genericErrorResponseHandler(statuses)
		}

		return nil, errorResponseHandler(resp.StatusCode, respBytes)
	}

	return respBytes, nil
}

// DoAndParse executes the request in the background context and reads the response body.
// If a status other than 200 is received from the endpoint, then errorResponseHandler is called to generate the
// error that gets returned. If errorResponseHandler is nil, then a generic error response handler will be used.
func (r *Request) DoAndParse(method, endpointURL, contentType string, body io.Reader,
	event, parentEvent string, errorResponseHandler func(statusCode int, responseBody []byte) error,
	response interface{},
) error {
	respBytes, err := r.Do(method, endpointURL, contentType, body,
		event, parentEvent, errorResponseHandler)
	if err != nil {
		return err
	}

	return json.Unmarshal(respBytes, response)
}

func genericErrorResponseHandler(expectedStatusCodes []int) func(statusCode int, respBytes []byte) error {
	return func(statusCode int, respBytes []byte) error {
		if len(expectedStatusCodes) == 1 {
			return fmt.Errorf(
				"expected status code %d but got status code %d with response body %s instead",
				expectedStatusCodes[0], statusCode, respBytes)
		}

		return fmt.Errorf(
			"expected status codes %v but got status code %d with response body %s instead",
			expectedStatusCodes, statusCode, respBytes)
	}
}
