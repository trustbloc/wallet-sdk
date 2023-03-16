/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package credentialstatus provides a client for verifying Verifiable Credential revocation status using the
// Credential.Status field.
package credentialstatus

import (
	"fmt"
	"net/http"

	"github.com/hyperledger/aries-framework-go-ext/component/vc/status"
	"github.com/hyperledger/aries-framework-go-ext/component/vc/status/resolver"
	"github.com/hyperledger/aries-framework-go-ext/component/vc/status/validator"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
)

// Config holds parameters for initializing a Verifier.
type Config struct {
	HTTPClient *http.Client
}

// Verifier verifies Credential Status.
type Verifier struct {
	client statusClient
}

type statusClient interface {
	VerifyStatus(credential *verifiable.Credential) error
}

// NewVerifier creates a Credential Status Verifier.
func NewVerifier(config *Config) (*Verifier, error) {
	client := &status.Client{
		ValidatorGetter: validator.GetValidator,
		Resolver:        resolver.NewResolver(config.HTTPClient, ""),
	}

	return &Verifier{
		client: client,
	}, nil
}

// Verify checks the Credential Status, returning an error if the status field is invalid, the status is revoked, or if
// it isn't possible to verify the credential's status.
func (v *Verifier) Verify(vc *verifiable.Credential) error {
	err := v.client.VerifyStatus(vc)
	if err != nil {
		return fmt.Errorf("status verification failed: %w", err)
	}

	return nil
}
