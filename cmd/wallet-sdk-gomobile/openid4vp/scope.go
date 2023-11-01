/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp

// Scope represents an array of scope strings.
// Since arrays and slices are not compatible with gomobile, this type acts as a wrapper around a Go array of strings.
type Scope struct {
	scope []string
}

// NewScope creates Scope object from array of scopes.
func NewScope(scope []string) *Scope {
	return &Scope{scope: scope}
}

// Length returns the number scopes.
func (s *Scope) Length() int {
	return len(s.scope)
}

// AtIndex returns scope by index.
func (s *Scope) AtIndex(index int) string {
	return s.scope[index]
}
