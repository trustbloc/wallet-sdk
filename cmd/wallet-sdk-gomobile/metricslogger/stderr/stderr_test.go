/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package stderr_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/metricslogger/stderr"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

func TestMetricsLogger_Log(t *testing.T) {
	metricsEvent := &api.MetricsEvent{
		GoAPIMetricsEvent: &goapi.MetricsEvent{
			Event:    "Event",
			Duration: time.Second,
		},
	}

	metricsLogger := stderr.NewMetricsLogger()

	t.Run("Without parent event", func(t *testing.T) {
		err := metricsLogger.Log(metricsEvent)
		require.NoError(t, err)
	})
	t.Run("With parent event", func(t *testing.T) {
		metricsEvent.GoAPIMetricsEvent.ParentEvent = "ParentEvent"

		err := metricsLogger.Log(metricsEvent)
		require.NoError(t, err)
	})
}
