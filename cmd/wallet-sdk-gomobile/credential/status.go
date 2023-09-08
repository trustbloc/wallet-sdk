/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/trustbloc/did-go/doc/did"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/credentialstatus"
)

// StatusVerifier verifies credential status.
type StatusVerifier struct {
	verifier *credentialstatus.Verifier
}

// NewStatusVerifier creates a credential status verifier.
// This StatusVerifier only supports HTTP resolution.
// To create a credential status verifier that also supports DID-URL resolution of Status Credentials,
// use NewStatusVerifierWithDIDResolver instead.
func NewStatusVerifier(opts *StatusVerifierOpts) (*StatusVerifier, error) {
	if opts == nil {
		opts = NewStatusVerifierOpts()
	}

	return newStatusVerifier(&unsupportedResolver{}, opts.httpTimeout)
}

// NewStatusVerifierWithDIDResolver creates a credential status verifier with a DID resolver.
func NewStatusVerifierWithDIDResolver(didResolver api.DIDResolver, opts *StatusVerifierOpts,
) (*StatusVerifier, error) {
	if didResolver == nil {
		return nil, errors.New("DID resolver must be provided. " +
			"If support for DID-URL resolution of status credentials is not needed, then use NewStatusVerifier instead")
	}

	if opts == nil {
		opts = NewStatusVerifierOpts()
	}

	return newStatusVerifier(&wrapper.VDRResolverWrapper{DIDResolver: didResolver}, opts.httpTimeout)
}

func newStatusVerifier(didResolver goapi.DIDResolver, httpTimeout *time.Duration) (*StatusVerifier, error) {
	httpClient := &http.Client{}

	if httpTimeout != nil {
		httpClient.Timeout = *httpTimeout
	} else {
		httpClient.Timeout = goapi.DefaultHTTPTimeout
	}

	v, err := credentialstatus.NewVerifier(&credentialstatus.Config{
		HTTPClient:  httpClient,
		DIDResolver: didResolver,
	})
	if err != nil {
		return nil, err
	}

	return &StatusVerifier{
		verifier: v,
	}, nil
}

// Verify verifies credential status.
func (s *StatusVerifier) Verify(vc *verifiable.Credential) error {
	return s.verifier.Verify(vc.VC)
}

type unsupportedResolver struct{}

func (u *unsupportedResolver) Resolve(string) (*did.DocResolution, error) {
	return nil, fmt.Errorf("DID resolution not enabled for this VC status verifier. " +
		"Use NewStatusVerifierWithDIDResolver to enable support")
}
