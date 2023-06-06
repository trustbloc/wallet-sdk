/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package metricslogger

import (
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/metricslogger/stderr"
)

// MetricsLogger is a simple api.MetricsLogger implementation that saves all metrics Events in memory and writes them to
// standard error.
type MetricsLogger struct {
	Events              []*api.MetricsEvent
	stderrMetricsLogger *stderr.MetricsLogger
}

// NewMetricsLogger returns a new MetricsLogger.
func NewMetricsLogger() *MetricsLogger {
	return &MetricsLogger{
		Events:              make([]*api.MetricsEvent, 0),
		stderrMetricsLogger: stderr.NewMetricsLogger(),
	}
}

// Log saves the given metrics Events in memory and then writes it to standard error.
func (m *MetricsLogger) Log(metricsEvent *api.MetricsEvent) error {
	m.Events = append(m.Events, metricsEvent)

	return m.stderrMetricsLogger.Log(metricsEvent)
}
