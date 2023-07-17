/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package errors describe possible errors in JS implementation of walled-sdk.
package errors

// Errors for creator modules.
const (
	Module                      = "AGENT"
	InvalidArgumentsError       = "INVALID_ARGUMENTS"
	InitializationFailedError   = "INITIALIZATION_FAILED"
	MissedRequiredPropertyError = "MISSED_REQUIRED_PROPERTY"
	InvalidDisplayDataError     = "INVALID_DISPLAY_DATA"
)

// Category codes for creator module.
const (
	InvalidArgumentsCode = iota
	InitializationFailedCode
	MissedRequiredPropertyCode
	InvalidDisplayDataCode
)
