/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package noop contains an activity logger implementation that does nothing except satisfy the api.ActivityLogger
// interface. Using this effectively disabled activity logging. This activity logger is the default.
package noop

import (
	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// ActivityLogger is a "no-op" activity logger implementation that does nothing.
// It's the default implementation that gets used if a Wallet-SDK API is called with no activity logger specified.
type ActivityLogger struct{}

// NewActivityLogger returns a new ActivityLogger.
func NewActivityLogger() *ActivityLogger {
	return &ActivityLogger{}
}

// Log does nothing. It exists only to satisfy the api.ActivityLogger interface.
func (n *ActivityLogger) Log(*api.Activity) error {
	return nil
}
