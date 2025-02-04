/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package walleterror defines the error model for the Go mobile API.
package walleterror

import (
	"encoding/json"
)

// Error represents error returned by go mobile api.
type Error struct {
	// A short, alphanumeric code that includes the Category.
	Code string `json:"code"`
	// A short descriptor of the general category of error. This will always be a pre-defined string.
	Category string `json:"category"`
	// A short message describing the error that occurred. Only present in certain cases.
	// It will be blank in all others.
	Message string `json:"message"`
	// The full, raw error message. Any lower-level details about the precise cause of the error will be captured here.
	Details string `json:"details"`
	// ID of Open Telemetry root trace. Can be used to trace API calls. Only present in certain errors.
	TraceID string `json:"trace_id"`
	// Server error code.
	ServerCode string `json:"server_code,omitempty"`
	// Server error message.
	ServerMessage string `json:"server_message,omitempty"`
}

// Parse used to parse exception message on mobile side.
func Parse(errorMessage string) *Error {
	walletErr := &Error{}

	err := json.Unmarshal([]byte(errorMessage), walletErr)
	if err != nil {
		return &Error{
			Code:     "UKN2-000",
			Category: "OTHER_ERROR",
			Details:  errorMessage,
		}
	}

	return walletErr
}
