/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential

import (
	"fmt"
	"net/http"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	pkgapi "github.com/trustbloc/wallet-sdk/pkg/api"
	"github.com/trustbloc/wallet-sdk/pkg/credentialstatus"
)

// StatusVerifier verifies credential status.
type StatusVerifier struct {
	verifier *credentialstatus.Verifier
}

// StatusVerifierOptionalArgs contains optional parameters for initializing a credential StatusVerifier.
type StatusVerifierOptionalArgs struct {
	didResolver api.DIDResolver
}

// NewStatusVerifierOptionalArgs returns a StatusVerifierOptionalArgs object.
func NewStatusVerifierOptionalArgs() *StatusVerifierOptionalArgs {
	return &StatusVerifierOptionalArgs{}
}

// SetDIDResolver sets the DID resolver to use.
// If no DID resolver is explicitly set, then the initialized StatusVerifier will not support
// DID-URL resolution of Status Credentials, but will still support HTTP resolution.
func (o *StatusVerifierOptionalArgs) SetDIDResolver(didResolver api.DIDResolver) {
	o.didResolver = didResolver
}

// NewStatusVerifier creates a credential status verifier.
func NewStatusVerifier(optionalArgs *StatusVerifierOptionalArgs) (*StatusVerifier, error) {
	var useDIDResolver pkgapi.DIDResolver

	if optionalArgs.didResolver == nil {
		useDIDResolver = &unsupportedResolver{}
	} else {
		useDIDResolver = &wrapper.VDRResolverWrapper{DIDResolver: optionalArgs.didResolver}
	}

	v, err := credentialstatus.NewVerifier(&credentialstatus.Config{
		HTTPClient:  http.DefaultClient,
		DIDResolver: useDIDResolver,
	})
	if err != nil {
		return nil, err
	}

	return &StatusVerifier{
		verifier: v,
	}, nil
}

// Verify verifies credential status.
func (s *StatusVerifier) Verify(vc *api.VerifiableCredential) error {
	return s.verifier.Verify(vc.VC)
}

type unsupportedResolver struct{}

func (u *unsupportedResolver) Resolve(string) (*did.DocResolution, error) {
	return nil, fmt.Errorf("did resolution not enabled for this VC status verifier")
}
