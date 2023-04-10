/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

import goapiopenid4vp "github.com/trustbloc/wallet-sdk/pkg/openid4vp"

// VerifierDisplayData represents display information for a verifier.
type VerifierDisplayData struct {
	displayData *goapiopenid4vp.VerifierDisplayData
}

// DID returns the verifier's DID.
func (v *VerifierDisplayData) DID() string {
	return v.displayData.DID
}

// Name returns the verifier's name.
func (v *VerifierDisplayData) Name() string {
	return v.displayData.Name
}

// Purpose returns the verifier's purpose.
func (v *VerifierDisplayData) Purpose() string {
	return v.displayData.Purpose
}

// LogoURI returns the verifier's logo URI.
func (v *VerifierDisplayData) LogoURI() string {
	return v.displayData.LogoURI
}
