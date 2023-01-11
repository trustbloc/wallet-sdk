/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialquery

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	module                               = "CRQ"
	CredentialReaderNotSetError          = "CREDENTIAL_READER_NOT_SET"     //nolint:gosec //false positive
	CredentialReaderReadFailedError      = "CREDENTIAL_READER_READ_FAILED" //nolint:gosec //false positive
	CreateVPFailedError                  = "CREATE_VP_FAILED"
	NoCredentialSatisfyRequirementsError = "NO_CREDENTIAL_SATISFY_REQUIREMENTS" //nolint:gosec //false positive
)

// Constants' names and reasons are obvious so they do not require additional comments.
// nolint:golint,nolintlint
const (
	CredentialReaderNotSetCode = iota
	CredentialReaderReadFailedCode
	CreateVPFailedCode
	NoCredentialSatisfyRequirementsCode
)
