/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httprequest_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/internal/httprequest"
	"github.com/trustbloc/wallet-sdk/pkg/internal/mock"
	"github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
)

func Test_doHTTPRequest(t *testing.T) {
	t.Run("Invalid http method", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{StatusCode: 200}, noop.NewMetricsLogger())
		_, err := r.Do(http.MethodGet, "url", "test", nil,
			"", "", nil)
		require.NoError(t, err)
	})

	t.Run("Invalid http method", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{}, noop.NewMetricsLogger())
		_, err := r.Do("\n\n", "url", "", nil,
			"", "", nil)
		require.Contains(t, err.Error(), "invalid method")
	})

	t.Run("Failing metric logger", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{}, &failingMetricsLogger{})
		_, err := r.Do(http.MethodGet, "url", "test", nil,
			"", "", nil)
		require.Contains(t, err.Error(), "failed to log event (Event=)")
	})

	t.Run("Invalid http code", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{}, noop.NewMetricsLogger())

		_, err := r.Do(http.MethodGet, "url", "", nil,
			"", "", nil)
		require.Contains(t, err.Error(), "expected status code 200")
	})

	t.Run("Invalid http code", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{
			StatusCode: 200, Err: errors.New("request err"),
		}, noop.NewMetricsLogger())

		_, err := r.Do(http.MethodGet, "url", "", nil,
			"", "", nil)
		require.Contains(t, err.Error(), "request err")
	})
}

func Test_DoContextHTTPRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{StatusCode: 200}, noop.NewMetricsLogger())

		additionalHeaders := http.Header{}
		additionalHeaders.Add("X-Header", "12345")

		_, err := r.DoContext(context.Background(), http.MethodGet, "url", "test", additionalHeaders,
			nil, "", "", []int{200, 201}, nil)
		require.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{StatusCode: 400}, noop.NewMetricsLogger())

		additionalHeaders := http.Header{}
		additionalHeaders.Add("X-Header", "12345")

		_, err := r.DoContext(context.Background(), http.MethodGet, "url", "test", additionalHeaders,
			nil, "", "", []int{200, 201}, nil)
		require.Error(t, err)
	})
}

func Test_DoAndParseHTTPRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{
			StatusCode: 200,
			Response:   `"response"`,
		},
			noop.NewMetricsLogger(),
		)

		additionalHeaders := http.Header{}
		additionalHeaders.Add("X-Header", "12345")

		var response string

		err := r.DoAndParse(http.MethodGet, "url", "test", nil,
			"", "", nil, &response)
		require.NoError(t, err)
		require.Equal(t, "response", response)
	})
	t.Run("Failure: unexpected status code", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{
			StatusCode: 400,
			Response:   `"response"`,
		},
			noop.NewMetricsLogger())

		additionalHeaders := http.Header{}
		additionalHeaders.Add("X-Header", "12345")

		var response string

		err := r.DoAndParse(http.MethodGet, "url", "test", nil,
			"", "", nil, &response)
		require.Error(t, err)
		require.ErrorContains(t, err, "expected status code 200 but got status code 400")
	})
}

type failingMetricsLogger struct{}

func (f *failingMetricsLogger) Log(metricsEvent *api.MetricsEvent) error {
	return fmt.Errorf("failed to log event (Event=%s)", metricsEvent.Event)
}
