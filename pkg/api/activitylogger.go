/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"time"

	"github.com/google/uuid"
)

// Params represents additional parameters which may be required for wallet applications in the future.
// As such, this is currently a placeholder.
type Params map[string]interface{}

// Data represents specific details of an activity.
type Data struct {
	Client    string `json:"client,omitempty"`
	Operation string `json:"operation,omitempty"`
	Status    string `json:"status,omitempty"`
	Params    Params `json:"params,omitempty"`
}

// An Activity represents a single activity.
type Activity struct {
	ID   uuid.UUID `json:"id,omitempty"`
	Type string    `json:"type,omitempty"`
	Time time.Time `json:"time"`
	Data Data      `json:"data"`
}

const (
	// LogTypeCredentialActivity is the string used in an activity to indicate that it is related to a
	// credential operation.
	LogTypeCredentialActivity = "credential-activity" //nolint:gosec // false positive
	// ActivityLogStatusSuccess is the string used in log entries indicating a successful operation.
	ActivityLogStatusSuccess = "success"
)

// An ActivityLogger logs activities.
type ActivityLogger interface {
	// Log logs a single activity.
	Log(credential *Activity) error
}
