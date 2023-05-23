/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"fmt"

	"github.com/hyperledger/aries-framework-go/component/kmscrypto/doc/jose/jwk"
)

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
