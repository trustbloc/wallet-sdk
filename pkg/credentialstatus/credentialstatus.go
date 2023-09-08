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

	diddoc "github.com/trustbloc/did-go/doc/did"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	"github.com/trustbloc/vc-go/status"
	"github.com/trustbloc/vc-go/status/resolver"
	"github.com/trustbloc/vc-go/status/validator"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/wallet-sdk/pkg/api"
)

// Config holds parameters for initializing a Verifier.
type Config struct {
	HTTPClient  *http.Client
	DIDResolver api.DIDResolver
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
		Resolver: resolver.NewResolver(
			config.HTTPClient,
			&wrapResolver{resolver: config.DIDResolver},
			"",
		),
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

type wrapResolver struct {
	resolver api.DIDResolver
}

func (w *wrapResolver) Resolve(did string, _ ...vdrapi.DIDMethodOption) (*diddoc.DocResolution, error) {
	return w.resolver.Resolve(did)
}

func (w *wrapResolver) Create(string, *diddoc.Doc, ...vdrapi.DIDMethodOption) (*diddoc.DocResolution, error) {
	return nil, fmt.Errorf("create operation is not supported")
}

func (w *wrapResolver) Update(*diddoc.Doc, ...vdrapi.DIDMethodOption) error {
	return fmt.Errorf("update operation is not supported")
}

func (w *wrapResolver) Deactivate(string, ...vdrapi.DIDMethodOption) error {
	return fmt.Errorf("deactivate operation is not supported")
}

func (w *wrapResolver) Close() error {
	return nil
}
