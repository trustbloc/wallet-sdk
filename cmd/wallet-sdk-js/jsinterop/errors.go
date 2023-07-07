/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package jsinterop implements interop between Wallet-SDK and JS.
package jsinterop

// Errors for creator modules.
const (
	Module                    = "AGENT"
	InvalidArgumentsError     = "INVALID_ARGUMENTS"
	InitializationFailedError = "INITIALIZATION_FAILED"
)

// Category codes for creator module.
const (
	InvalidArgumentsCode = iota
	InitializationFailedCode
)
