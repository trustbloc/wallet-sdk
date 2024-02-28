/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package mock contains common mocks to be used on tests.
package mock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPClientMock used to mock http client in tests.
type HTTPClientMock struct {
	Response            string
	StatusCode          int
	Err                 error
	ExpectedEndpoint    string
	SentBody            []byte
	SentBodyUnMarshaled interface{}
}

// Do mocks call to http client Do function.
func (c *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	if c.ExpectedEndpoint != "" && c.ExpectedEndpoint != req.URL.String() {
		return nil, fmt.Errorf("requested endpoint %s does not match %s", req.URL.String(), c.ExpectedEndpoint)
	}

	if req.Body != nil {
		respBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		c.SentBody = respBytes
		if c.SentBodyUnMarshaled != nil {
			err = json.Unmarshal(respBytes, c.SentBodyUnMarshaled)
			if err != nil {
				return nil, err
			}
		}
	}

	if c.Err != nil {
		return nil, c.Err
	}

	return &http.Response{
		StatusCode: c.StatusCode,
		Body:       io.NopCloser(bytes.NewBufferString(c.Response)),
	}, nil
}
