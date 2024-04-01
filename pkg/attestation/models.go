/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation

// AttestWalletInitRequest attestation init request.
type AttestWalletInitRequest struct {
	Payload map[string]interface{} `json:"payload"`
}

// AttestWalletInitResponse attestation init response.
type AttestWalletInitResponse struct {
	Challenge string `json:"challenge"`
	SessionID string `json:"session_id"`
}

// AttestWalletCompleteRequest attestation complete request.
type AttestWalletCompleteRequest struct {
	AssuranceLevel string `json:"assurance_level"`
	Proof          Proof  `json:"proof"`
	SessionID      string `json:"session_id"`
}

// Proof complete request proof.
type Proof struct {
	Jwt       string `json:"jwt,omitempty"`
	ProofType string `json:"proof_type"`
}

type jwtProofClaims struct {
	Issuer   string `json:"iss,omitempty"`
	Audience string `json:"aud,omitempty"`
	IssuedAt int64  `json:"iat,omitempty"`
	Nonce    string `json:"nonce,omitempty"`
	Exp      int64  `json:"exp,omitempty"`
}

// AttestWalletCompleteResponse attestation complete response.
type AttestWalletCompleteResponse struct {
	WalletAttestationVC string `json:"wallet_attestation_vc"`
}
