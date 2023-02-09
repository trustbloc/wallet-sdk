/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package common

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	module                    = "COM"
	UnsupportedAlgorithmError = "UNSUPPORTED_ALGORITHM"
	NoCryptoProvidedError     = "NO_CRYPTO_PROVIDED"
)

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	UnsupportedAlgorithmCode = iota
	NoCryptoProvidedCode
)
