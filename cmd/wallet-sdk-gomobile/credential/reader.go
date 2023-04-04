/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"

// A Reader is capable of reading VCs from some underlying storage mechanism.
type Reader interface {
	// Get retrieves a VC.
	Get(id string) (*verifiable.Credential, error)
	// GetAll retrieves all VCs.
	GetAll() (*verifiable.CredentialsArray, error)
}
