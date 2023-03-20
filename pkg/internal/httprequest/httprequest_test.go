/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package httprequest_test

import (
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
			"", "")
		require.NoError(t, err)
	})

	t.Run("Invalid http method", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{}, noop.NewMetricsLogger())
		_, err := r.Do("\n\n", "url", "", nil,
			"", "")
		require.Contains(t, err.Error(), "invalid method")
	})

	t.Run("Failing metric logger", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{}, &failingMetricsLogger{})
		_, err := r.Do(http.MethodGet, "url", "test", nil,
			"", "")
		require.Contains(t, err.Error(), "failed to log event (Event=)")
	})

	t.Run("Invalid http code", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{}, noop.NewMetricsLogger())

		_, err := r.Do(http.MethodGet, "url", "", nil,
			"", "")
		require.Contains(t, err.Error(), "expected status code 200")
	})

	t.Run("Invalid http code", func(t *testing.T) {
		r := httprequest.New(&mock.HTTPClientMock{
			StatusCode: 200, Err: errors.New("request err"),
		}, noop.NewMetricsLogger())

		_, err := r.Do(http.MethodGet, "url", "", nil,
			"", "")
		require.Contains(t, err.Error(), "request err")
	})
}

type failingMetricsLogger struct{}

func (f *failingMetricsLogger) Log(metricsEvent *api.MetricsEvent) error {
	return fmt.Errorf("failed to log event (Event=%s)", metricsEvent.Event)
}
