/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package walleterror defines the error model for the Go API.
package walleterror

import "fmt"

const (
	validationError     = 0
	executionError      = 1
	systemError         = 2
	incorrectUsageError = 3
)

// Error represents an error returned by the Go API.
type Error struct {
	Code        string
	Scenario    string
	ParentError string
}

// NewValidationError creates validation error.
func NewValidationError(module string, code int, errorName string, parentError error) *Error {
	return &Error{
		Code:        getErrorCode(module, validationError, code),
		Scenario:    errorName,
		ParentError: parentError.Error(),
	}
}

// NewExecutionError creates execution error.
func NewExecutionError(module string, code int, scenario string, cause error) *Error {
	return &Error{
		Code:        getErrorCode(module, executionError, code),
		Scenario:    scenario,
		ParentError: cause.Error(),
	}
}

// NewSystemError creates system error.
func NewSystemError(module string, code int, errorName string, parentError error) *Error {
	return &Error{
		Code:        getErrorCode(module, systemError, code),
		Scenario:    errorName,
		ParentError: parentError.Error(),
	}
}

// NewInvalidSDKUsageError creates a new invalid SDK usage error, used to indicate when a Wallet-SDK API is used
// incorrectly.
func NewInvalidSDKUsageError(module string, parentError error) *Error {
	return &Error{
		Code:        getErrorCode(module, incorrectUsageError, 0),
		Scenario:    "INVALID_SDK_USAGE",
		ParentError: parentError.Error(),
	}
}

// Error returns string representation of error.
func (e *Error) Error() string {
	return fmt.Sprintf("%s(%s):%s", e.Scenario, e.Code, e.ParentError)
}

func getErrorCode(module string, errType, code int) string {
	return fmt.Sprintf("%s%d-%04d", module, errType, code)
}
