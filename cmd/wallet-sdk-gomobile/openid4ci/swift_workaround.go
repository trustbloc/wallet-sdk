/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci

// Constructors workaround to generate swift compatible objective-c functions.
type Constructors struct{}

// NewInteraction workaround to generate swift compatible objective-c functions.
func (c *Constructors) NewInteraction(initiateIssuanceURI string, config *ClientConfig) (*Interaction, error) {
	return NewInteraction(initiateIssuanceURI, config)
}
