/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"fmt"

	"github.com/trustbloc/kms-go/doc/jose/jwk"
)

// JSONWebKey holds a public key with associated metadata, in JWK format.
type JSONWebKey struct {
	JWK *jwk.JWK `json:"jwk,omitempty"`
}

// Serialize returns a JSON representation of this JSONWebKey.
func (k *JSONWebKey) Serialize() (string, error) {
	if k.JWK == nil {
		return "", fmt.Errorf("json web key has no data to serialize")
	}

	keyBytes, err := k.JWK.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("serializing json web key: %w", err)
	}

	return string(keyBytes), nil
}

// ID returns the Key ID for this JSONWebKey.
func (k *JSONWebKey) ID() string {
	if k.JWK == nil {
		return ""
	}

	return k.JWK.KeyID
}

// ParseJSONWebKey parses a JSONWebKey from a JWK in JSON format.
func ParseJSONWebKey(data string) (*JSONWebKey, error) {
	key := &jwk.JWK{}

	err := key.UnmarshalJSON([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("parsing json web key: %w", err)
	}

	return &JSONWebKey{
		JWK: key,
	}, nil
}

// JSONWebKeySet represents a JWK Set object.
type JSONWebKeySet struct {
	JWKs []JSONWebKey
}

// NewJSONWebKeySet returns a new JSON Web Key Set.
// It acts as a gomobile-compatible wrapper around a Go array of JSONWebKey objects.
func NewJSONWebKeySet() *JSONWebKeySet {
	return &JSONWebKeySet{}
}

// Append appends the given JSONWebKey to the array.
// It returns a reference to the JSONWebKey in order to allow a caller to chain together Append calls.
func (j *JSONWebKeySet) Append(jsonWebKey *JSONWebKey) *JSONWebKeySet {
	j.JWKs = append(j.JWKs, *jsonWebKey)

	return j
}

// Length returns the number of JSONWebKeys contained within this JSONWebKeySet object.
func (j *JSONWebKeySet) Length() int {
	return len(j.JWKs)
}

// AtIndex returns the JSONWebKey at the given index.
// If the index passed in is out of bounds, then nil is returned.
func (j *JSONWebKeySet) AtIndex(index int) *JSONWebKey {
	maxIndex := len(j.JWKs) - 1
	if index > maxIndex || index < 0 {
		return nil
	}

	return &j.JWKs[index]
}
