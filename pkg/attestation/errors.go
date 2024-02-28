/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package attestation provides APIs for wallets to receive attestation credential.
package attestation

// Constants' names and reasons are obvious, so they do not require additional comments.
// nolint:golint,nolintlint
const (
	ErrorModule                   = "ATT"
	JWTSigningFailedError         = "JWT_SIGNING_FAILED"
	KeyIDMissingDIDPartError      = "KEY_ID_MISSING_DID_PART"
	ParseAttestationVCFailedError = "PARSE_ATTESTATION_VC_FAILED"
)

// Constants' names and reasons are obvious, so they do not require additional comments.
// nolint:golint,nolintlint
const (
	JWTSigningFailedCode         = 1
	KeyIDMissingDIDPartCode      = 2
	ParseAttestationVCFailedCode = 3
)
