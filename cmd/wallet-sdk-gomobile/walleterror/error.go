/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package walleterror defines the error model for the Go mobile API.
package walleterror

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// Error represents error returned by go mobile api.
type Error struct {
	Code     string `json:"code"`
	Category string `json:"category"`
	Details  string `json:"details"`
}

// ToMobileError translates go api errors to go mobile errors.
func ToMobileError(err error) error {
	if err == nil {
		return nil
	}

	var result *Error

	var walletError *walleterror.Error

	if errors.As(err, &walletError) {
		result = &Error{
			Code:     walletError.Code,
			Category: walletError.Scenario,
			Details:  walletError.ParentError,
		}
	} else {
		result = &Error{
			Code:     "UKN2-000",
			Category: "UNEXPECTED_ERROR",
			Details:  err.Error(),
		}
	}

	formatted, fmtErr := json.Marshal(result)
	if fmtErr != nil {
		return fmt.Errorf("unexpected format error: %w", fmtErr)
	}

	return errors.New(string(formatted))
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
