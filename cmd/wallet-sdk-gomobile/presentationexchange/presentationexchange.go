/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package presentationexchange defines functionality for doing presentation exchange operations.
package presentationexchange

// An Exchange is used to perform matching operations.
type Exchange struct{}

// NewExchange returns a new Exchange object.
func NewExchange() *Exchange {
	return &Exchange{}
}

// Match does something (TODO: Implement).
func (p *Exchange) Match(presentationDefinition, vp []byte, matchOptions string) ([]byte, error) {
	return []byte("Example data"), nil
}
