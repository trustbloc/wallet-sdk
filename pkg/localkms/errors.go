/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	module                    = "LKM"
	InitialisationFailedError = "INITIALISATION_FAILED"
	CreateKeyFailedError      = "CREATE_KEY_FAILED"
)

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	InitialisationFailedCode = iota
	CreateKeyFailedCode
)
