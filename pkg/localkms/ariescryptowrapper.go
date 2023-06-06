/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package localkms

import (
	"fmt"
	"strings"

	"github.com/hyperledger/aries-framework-go/spi/crypto"
	"github.com/hyperledger/aries-framework-go/spi/kms"
)

// AriesCryptoWrapper wraps aries crypto implementations to conform api.Crypto interface.
type AriesCryptoWrapper struct {
	cryptosKMS    kms.KeyManager
	wrappedCrypto crypto.Crypto
}

// NewAriesCryptoWrapper returns new instance of AriesCryptoWrapper.
func NewAriesCryptoWrapper(cryptosKMS kms.KeyManager, wrappedCrypto crypto.Crypto) *AriesCryptoWrapper {
	return &AriesCryptoWrapper{
		cryptosKMS:    cryptosKMS,
		wrappedCrypto: wrappedCrypto,
	}
}

// Sign gets key from kms using keyID and use it to sign data.
func (c *AriesCryptoWrapper) Sign(msg []byte, keyID string) ([]byte, error) {
	kidParts := strings.Split(keyID, "#")
	if len(kidParts) == 2 { //nolint: gomnd
		keyID = kidParts[1]
	}

	kh, err := c.cryptosKMS.Get(keyID)
	if err != nil {
		return nil, fmt.Errorf("invalid key id %q: %w", keyID, err)
	}

	return c.wrappedCrypto.Sign(msg, kh)
}

// Verify gets key from kms using keyID and use it to verify data.
func (c *AriesCryptoWrapper) Verify(signature, msg []byte, keyID string) error {
	kh, err := c.cryptosKMS.Get(keyID)
	if err != nil {
		return fmt.Errorf("invalid key id %q: %w", keyID, err)
	}

	return c.wrappedCrypto.Verify(signature, msg, kh)
}
