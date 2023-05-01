/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

// StringArray represents an array of strings.
// It wraps a standard Go array and provides gomobile-compatible methods that allow a caller to use this type
// in an array-like manner.
type StringArray struct {
	Strings []string
}

// NewStringArray returns a new string array. It acts as a gomobile-compatible wrapper around a Go string array.
func NewStringArray() *StringArray {
	return &StringArray{}
}

// Append appends the given string to the array.
// It returns a reference to the StringArray in order to allow a caller to chain together Append calls.
func (s *StringArray) Append(value string) *StringArray {
	s.Strings = append(s.Strings, value)

	return s
}

// Length returns the number of strings contained within this StringArray object.
func (s *StringArray) Length() int {
	return len(s.Strings)
}

// AtIndex returns the string at the given index.
// If the index passed in is out of bounds, then an empty string is returned.
func (s *StringArray) AtIndex(index int) string {
	maxIndex := len(s.Strings) - 1
	if index > maxIndex || index < 0 {
		return ""
	}

	return s.Strings[index]
}
