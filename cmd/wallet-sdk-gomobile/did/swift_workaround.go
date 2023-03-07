/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package did

import "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"

// Constructors workaround to generate swift compatible objective-c functions.
type Constructors struct{}

// NewCreatorWithKeyWriter workaround to generate swift compatible objective-c functions.
func (c *Constructors) NewCreatorWithKeyWriter(keyWriter api.KeyWriter) (*Creator, error) {
	return NewCreatorWithKeyWriter(keyWriter)
}

// NewCreatorWithKeyReader workaround to generate swift compatible objective-c functions.
func (c *Constructors) NewCreatorWithKeyReader(keyReader api.KeyReader) (*Creator, error) {
	return NewCreatorWithKeyReader(keyReader)
}

// NewResolver workaround to generate swift compatible objective-c functions.
func (c *Constructors) NewResolver(resolverServerURI string) (*Resolver, error) {
	return NewResolver(resolverServerURI)
}
