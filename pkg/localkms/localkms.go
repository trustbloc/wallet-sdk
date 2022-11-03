/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package localkms contains a KMS implementation that uses local storage.
// It is not intended for production use and may not be secure.
package localkms

import (
	"errors"
	"fmt"

	arieskms "github.com/hyperledger/aries-framework-go/pkg/kms"
	arieslocalkms "github.com/hyperledger/aries-framework-go/pkg/kms/localkms"
)

// LocalKMS is a KMS implementation that uses local storage.
// It is not intended for production use and may not be secure.
type LocalKMS struct {
	ariesLocalKMS *arieslocalkms.LocalKMS
}

// NewLocalKMS returns a new Local KMS.
func NewLocalKMS() (*LocalKMS, error) {
	ariesLocalKMS, err := arieslocalkms.New("ThisIs://Unused", &provider{})
	if err != nil {
		return nil, err
	}

	return &LocalKMS{ariesLocalKMS: ariesLocalKMS}, nil
}

// Create creates a keyset of the given keyType and then writes it to storage.
// The keyID and raw public key bytes of the newly generated keyset are returned.
// Currently, this method only supports creating ED25519 keys.
func (k *LocalKMS) Create(keyType arieskms.KeyType) (string, []byte, error) {
	// The CreateAndExportPubKeyBytes method we use from the Aries Local KMS implementation returns raw bytes for
	// some key types, and marshalled key bytes for others. If we want to support other key types in the future
	// then we will need to ensure that either this method either only returns raw bytes for consistency (by converting
	// to raw bytes as needed) or that the KeyWriter interface documentation is updated to make the expected key format
	// clear.
	if keyType != arieskms.ED25519Type {
		return "", nil, fmt.Errorf("key type %s not supported", keyType)
	}

	return k.ariesLocalKMS.CreateAndExportPubKeyBytes(keyType)
}

// GetKey returns the public key associated with the given keyID as raw bytes.
func (k *LocalKMS) GetKey(string) ([]byte, error) {
	return nil, errors.New("not implemented")
}
