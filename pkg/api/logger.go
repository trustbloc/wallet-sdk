/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import "time"

// LogParams represents additional log data for a given log event, the contents of which will differ
// depending on the scenario.
type LogParams map[string]interface{}

// LogData represents specific details of a log entry.
type LogData struct {
	Client    string    `json:"client,omitempty"`
	Operation string    `json:"operation,omitempty"`
	Status    string    `json:"status,omitempty"`
	Params    LogParams `json:"params,omitempty"`
}

// A LogEntry represents a single log entry.
type LogEntry struct {
	ID   string     `json:"id,omitempty"`
	Type string     `json:"type,omitempty"`
	Time *time.Time `json:"time,omitempty"`
	Data *LogData   `json:"data,omitempty"`
}

const (
	// LogTypeCredentialActivity is the string used in log entries relating to credential operations.
	LogTypeCredentialActivity = "credential-activity" //nolint:gosec // false positive
	// LogStatusSuccess is the string used in log entries indicating a successful operation.
	LogStatusSuccess = "success"
	// LogStatusFailure is the string used in log entries indicating a failure/error of some sort.
	LogStatusFailure = "failure"
)

// A Logger is capable of writing log entries.
type Logger interface {
	// Log logs a single log entry.
	Log(log *LogEntry)
}
