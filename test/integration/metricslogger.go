/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package integration

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/metricslogger/stderr"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// MetricsLogger is a simple api.MetricsLogger implementation that saves all metrics events in memory and writes them to
// standard error.
type MetricsLogger struct {
	events              []*api.MetricsEvent
	stderrMetricsLogger *stderr.MetricsLogger
}

// NewMetricsLogger returns a new MetricsLogger.
func NewMetricsLogger() *MetricsLogger {
	return &MetricsLogger{
		events:              make([]*api.MetricsEvent, 0),
		stderrMetricsLogger: stderr.NewMetricsLogger(),
	}
}

// Log saves the given metrics events in memory and then writes it to standard error.
func (m *MetricsLogger) Log(metricsEvent *api.MetricsEvent) error {
	m.events = append(m.events, metricsEvent)

	return m.stderrMetricsLogger.Log(metricsEvent)
}
