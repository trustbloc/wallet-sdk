/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms

import (
	"fmt"
	"strings"

	"github.com/trustbloc/kms-go/wrapper/api"
)

// AriesCryptoWrapper wraps aries crypto implementations to conform api.Crypto interface.
type AriesCryptoWrapper struct {
	cryptoSuite api.Suite
}

// NewAriesCryptoWrapper returns new instance of AriesCryptoWrapper.
func NewAriesCryptoWrapper(cryptoSuite api.Suite) *AriesCryptoWrapper {
	return &AriesCryptoWrapper{
		cryptoSuite: cryptoSuite,
	}
}

// Sign gets key from kms using keyID and use it to sign data.
func (c *AriesCryptoWrapper) Sign(msg []byte, keyID string) ([]byte, error) {
	kidParts := strings.Split(keyID, "#")
	if len(kidParts) == 2 { //nolint: mnd
		keyID = kidParts[1]
	}

	fks, err := c.cryptoSuite.FixedKeySigner(keyID)
	if err != nil {
		return nil, fmt.Errorf("invalid key id %q: %w", keyID, err)
	}

	return fks.Sign(msg)
}
