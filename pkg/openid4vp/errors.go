/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	module                                = "OVP"
	RequestObjectFetchFailedError         = "REQUEST_OBJECT_FETCH_FAILED"
	VerifyAuthorizationRequestFailedError = "VERIFY_AUTHORIZATION_REQUEST_FAILED"
	CreateAuthorizedResponseFailedError   = "CREATE_AUTHORIZED_RESPONSE"
	SendAuthorizedResponseFailedError     = "SEND_AUTHORIZED_RESPONSE"
)

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	RequestObjectFetchFailedCode = iota
	VerifyAuthorizationRequestFailedCode
	CreateAuthorizedResponseFailedCode
	SendAuthorizedResponseFailedCode
)
