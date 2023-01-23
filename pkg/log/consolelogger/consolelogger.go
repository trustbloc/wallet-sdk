/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package consolelogger contains a simple logger implementation that writes log entries directly to the console.
package consolelogger

import (
	"encoding/json"
	"fmt"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// ConsoleLogger is a simple logger implementation that writes log entries directly to the console.
// It's used as the default logger in many places in the SDK.
type ConsoleLogger struct{}

// NewConsoleLogger returns a new ConsoleLogger.
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

// Log marshals the given log entry and writes the resulting JSON to the console.
func (c *ConsoleLogger) Log(log *api.LogEntry) {
	marshalledLog, err := json.Marshal(log)
	if err != nil {
		println(fmt.Sprintf("failed to write a log message (type:%s)", log.Type))
	}

	println(string(marshalledLog))
}
