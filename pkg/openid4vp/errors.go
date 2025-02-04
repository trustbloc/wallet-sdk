/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	ErrorModule                                 = "OVP"
	InvalidAuthorizationRequestError            = "INVALID_AUTHORIZATION_REQUEST"
	RequestObjectFetchFailedError               = "REQUEST_OBJECT_FETCH_FAILED"
	CreateAuthorizedResponseFailedError         = "CREATE_AUTHORIZED_RESPONSE_FAILED"
	InvalidScopeError                           = "INVALID_SCOPE"
	InvalidRequestError                         = "INVALID_REQUEST"
	InvalidClientError                          = "INVALID_CLIENT"
	VPFormatsNotSupportedError                  = "VP_FORMATS_NOT_SUPPORTED" //nolint:gosec,lll //false positive, can't shorten
	InvalidPresentationDefinitionURIError       = "INVALID_PRESENTATION_DEFINITION_URI"
	InvalidPresentationDefinitionReferenceError = "INVALID_PRESENTATION_DEFINITION_REFERENCE"
	OtherAuthorizationResponseError             = "OTHER_AUTHORIZATION_RESPONSE_ERROR"
	MSEntraBadOrMissingFieldsError              = "MS_ENTRA_BAD_OR_MISSING_FIELDS"
	MSEntraNotFoundError                        = "MS_ENTRA_NOT_FOUND"
	MSEntraTokenError                           = "MS_ENTRA_TOKEN_ERROR"     //nolint:gosec,lll //false positive, can't shorten
	MSEntraTransientError                       = "MS_ENTRA_TRANSIENT_ERROR" //nolint:gosec,lll //false positive, can't shorten
)

// Constants' names and reasons are obvious, so they do not require additional comments.
// nolint:golint,nolintlint
const (
	InvalidAuthorizationRequestErrorCode            = 0
	RequestObjectFetchFailedCode                    = 1
	CreateAuthorizedResponseFailedCode              = 2
	InvalidScopeErrorCode                           = 3
	InvalidRequestErrorCode                         = 4
	InvalidClientErrorCode                          = 5
	VPFormatsNotSupportedErrorCode                  = 6
	InvalidPresentationDefinitionURIErrorCode       = 7
	InvalidPresentationDefinitionReferenceErrorCode = 8
	OtherAuthorizationResponseErrorCode             = 9
	MSEntraBadOrMissingFieldsErrorCode              = 10
	MSEntraNotFoundErrorCode                        = 11
	MSEntraTokenErrorCode                           = 12
	MSEntraTransientErrorCode                       = 13
)

type errorResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type msEntraErrorResponse struct {
	Error errorInfo `json:"error,omitempty"`
}

type errorInfo struct {
	InnerError innerError `json:"innerError,omitempty"`
}

type innerError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
