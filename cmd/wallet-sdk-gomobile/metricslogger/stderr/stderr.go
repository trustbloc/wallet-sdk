/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package stderr contains a simple api.MetricsLogger implementation that writes all metrics events to
// standard error.
package stderr

import (
	"fmt"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

// MetricsLogger is a simple api.MetricsLogger implementation that writes all metrics events to
// standard error.
type MetricsLogger struct{}

// NewMetricsLogger returns a new MetricsLogger.
// It writes metrics event logs to standard error.
func NewMetricsLogger() *MetricsLogger {
	return &MetricsLogger{}
}

// Log writes metrics event logs to standard error.
func (*MetricsLogger) Log(metricsEvent *api.MetricsEvent) error {
	message := "Metrics:"

	if metricsEvent.GoAPIMetricsEvent.ParentEvent != "" {
		message += fmt.Sprintf(" Parent Event: [%s]", metricsEvent.ParentEvent())
	}

	message += fmt.Sprintf(" Event=[%s] Duration=[%s]", metricsEvent.GoAPIMetricsEvent.Event,
		metricsEvent.GoAPIMetricsEvent.Duration.String())

	println(message)

	return nil
}
