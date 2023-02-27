/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

// StringArray represents an array of strings.
// It wraps a standard Go array and provides gomobile-compatible methods that allow a caller to use this type
// in an array-like manner.
type StringArray struct {
	strings []string
}

// Length returns the number of strings contained within this StringArray object.
func (s *StringArray) Length() int {
	return len(s.strings)
}

// AtIndex returns the string at the given index.
// If the index passed in is out of bounds, then an empty string is returned.
func (s *StringArray) AtIndex(index int) string {
	maxIndex := len(s.strings) - 1
	if index > maxIndex || index < 0 {
		return ""
	}

	return s.strings[index]
}
