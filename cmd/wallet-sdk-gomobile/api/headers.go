/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

// Header represents an HTTP header.
type Header struct {
	Name  string
	Value string
}

// NewHeader returns a new HTTP Header.
func NewHeader(name, value string) *Header {
	return &Header{
		Name:  name,
		Value: value,
	}
}

// Headers represents an array of HTTP headers that can be used in certain APIs.
type Headers struct {
	headers []Header
}

// NewHeaders returns a new Headers object.
func NewHeaders() *Headers {
	return &Headers{}
}

// Add adds the given HTTP Header to this list of headers.
func (h *Headers) Add(header *Header) {
	h.headers = append(h.headers, *header)
}

// GetAll returns all the headers set in this Headers object as an array of Header objects.
// This method is not compatible with gomobile and so will not be available in the generated bindings.
func (h *Headers) GetAll() []Header {
	return h.headers
}
