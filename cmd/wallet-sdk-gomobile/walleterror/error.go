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
	Code     string `json:"code"`
	Category string `json:"category"`
	Details  string `json:"details"`
	TraceID  string `json:"trace_id"`
}

// Parse used to parse exception message on mobile side.
func Parse(message string) *Error {
	walletErr := &Error{}

	err := json.Unmarshal([]byte(message), walletErr)
	if err != nil {
		return &Error{
			Code:     "GNR2-000",
			Category: "GENERAL_ERROR",
			Details:  message,
		}
	}

	return walletErr
}
