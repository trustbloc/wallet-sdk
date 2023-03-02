/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package noop contains a metrics logger implementation that does nothing except satisfy the api.MetricsLogger
// interface. Using this effectively disables metrics logging. This metrics logger is the default.
package noop

import (
	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// MetricsLogger is a "no-op" activity logger implementation that does nothing.
// It's the default implementation that gets used if a Wallet-SDK API is called with no activity logger specified.
type MetricsLogger struct{}

// NewMetricsLogger returns a new MetricsLogger.
func NewMetricsLogger() *MetricsLogger {
	return &MetricsLogger{}
}

// Log does nothing. It exists only to satisfy the api.MetricsLogger interface.
func (*MetricsLogger) Log(*api.MetricsEvent) error {
	return nil
}
