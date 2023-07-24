/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	ErrorModule                                 = "OVP"
	RequestObjectFetchFailedError               = "REQUEST_OBJECT_FETCH_FAILED"
	VerifyAuthorizationRequestFailedError       = "VERIFY_AUTHORIZATION_REQUEST_FAILED"
	CreateAuthorizedResponseFailedError         = "CREATE_AUTHORIZED_RESPONSE_FAILED"
	InvalidScopeError                           = "INVALID_SCOPE"
	InvalidRequestError                         = "INVALID_REQUEST"
	InvalidClientError                          = "INVALID_CLIENT"
	VPFormatsNotSupportedError                  = "VP_FORMATS_NOT_SUPPORTED"
	InvalidPresentationDefinitionURIError       = "INVALID_PRESENTATION_DEFINITION_URI"
	InvalidPresentationDefinitionReferenceError = "INVALID_PRESENTATION_DEFINITION_REFERENCE"
	OtherAuthorizationResponseError             = "OTHER_AUTHORIZATION_RESPONSE_ERROR"
)

// Constants' names and reasons are obvious, so they do not require additional comments.
// nolint:golint,nolintlint
const (
	RequestObjectFetchFailedCode                    = 0
	VerifyAuthorizationRequestFailedCode            = 1
	CreateAuthorizedResponseFailedCode              = 2
	InvalidScopeErrorCode                           = 3
	InvalidRequestErrorCode                         = 4
	InvalidClientErrorCode                          = 5
	VPFormatsNotSupportedErrorCode                  = 6
	InvalidPresentationDefinitionURIErrorCode       = 7
	InvalidPresentationDefinitionReferenceErrorCode = 8
	OtherAuthorizationResponseErrorCode             = 9
)

type errorResponse struct {
	Error string `json:"error,omitempty"`
}
