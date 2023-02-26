/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

func TestMetrics(t *testing.T) {
	metricsEvent := api.MetricsEvent{
		GoAPIMetricsEvent: &goapi.MetricsEvent{
			Event:       "Event",
			ParentEvent: "ParentEvent",
			Duration:    time.Second,
		},
	}

	require.Equal(t, "Event", metricsEvent.Event())
	require.Equal(t, "ParentEvent", metricsEvent.ParentEvent())
	require.Equal(t, int64(1000000000), metricsEvent.DurationNanoseconds())
}
