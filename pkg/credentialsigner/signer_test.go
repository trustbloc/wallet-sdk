/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialsigner_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/did-go/doc/did"
	afgotime "github.com/trustbloc/did-go/doc/util/time"
	"github.com/trustbloc/vc-go/verifiable"

	. "github.com/trustbloc/wallet-sdk/pkg/credentialsigner"
)

const (
	mockDID  = "did:test:foo"
	mockVMID = "#key-1"
	mockKID  = mockDID + mockVMID
)

func TestSigner_Issue(t *testing.T) {
	expectErr := errors.New("expected error")

	mockCredential, cErr := verifiable.CreateCredential(verifiable.CredentialContents{
		ID:      "foo",
		Types:   []string{verifiable.VCType},
		Context: []string{verifiable.V1ContextURI},
		Subject: []verifiable.Subject{{
			ID: "foo",
		}},
		Issuer: &verifiable.Issuer{
			ID: "did:foo:bar",
		},
		Issued: afgotime.NewTime(time.Now()),
	}, nil)

	require.NoError(t, cErr)

	t.Run("success", func(t *testing.T) {
		signer := New(&mockResolver{
			doc: mockDoc(t),
		}, &mockCrypto{})

		jwtVC, err := signer.Issue(mockCredential, &ProofOptions{
			KeyID:       mockKID,
			ProofFormat: ExternalJWTProofFormat,
		})
		require.NoError(t, err)
		require.NotEmpty(t, jwtVC)
	})

	t.Run("success did", func(t *testing.T) {
		signer := New(&mockResolver{
			doc: mockDoc(t),
		}, &mockCrypto{})

		jwtVC, err := signer.Issue(mockCredential, &ProofOptions{
			KeyID:       mockDID,
			ProofFormat: ExternalJWTProofFormat,
		})
		require.NoError(t, err)
		require.NotEmpty(t, jwtVC)
	})

	t.Run("invalid kid format", func(t *testing.T) {
		signer := New(&mockResolver{
			doc: mockDoc(t),
		}, &mockCrypto{})

		_, err := signer.Issue(mockCredential, &ProofOptions{
			KeyID:       "1#1#1",
			ProofFormat: ExternalJWTProofFormat,
		})
		require.ErrorContains(t, err, "invalid verification method format")
	})

	t.Run("no credential provided for signing", func(t *testing.T) {
		signer := New(&mockResolver{
			doc: mockDoc(t),
		}, &mockCrypto{})

		jwtVC, err := signer.Issue(nil, &ProofOptions{
			KeyID:       mockKID,
			ProofFormat: ExternalJWTProofFormat,
		})
		require.Error(t, err)
		require.Empty(t, jwtVC)
		require.Contains(t, err.Error(), "no credential provided")
	})

	t.Run("json-ld currently not implemented", func(t *testing.T) {
		signer := New(&mockResolver{
			doc: mockDoc(t),
		}, &mockCrypto{})

		jwtVC, err := signer.Issue(mockCredential, &ProofOptions{
			KeyID:       mockKID,
			ProofFormat: EmbeddedLDProofFormat,
		})
		require.Error(t, err)
		require.Empty(t, jwtVC)
		require.Contains(t, err.Error(), "JSON-LD proof format not currently supported")
	})

	t.Run("proof format not recognized", func(t *testing.T) {
		signer := New(&mockResolver{
			doc: mockDoc(t),
		}, &mockCrypto{})

		jwtVC, err := signer.Issue(mockCredential, &ProofOptions{
			KeyID:       mockKID,
			ProofFormat: "foo bar baz",
		})
		require.Error(t, err)
		require.Empty(t, jwtVC)
		require.Contains(t, err.Error(), "proof format not recognized")
	})

	t.Run("fail to resolve signing DID", func(t *testing.T) {
		signer := New(&mockResolver{
			err: expectErr,
		}, &mockCrypto{})

		jwtVC, err := signer.Issue(mockCredential, &ProofOptions{
			KeyID:       mockKID,
			ProofFormat: ExternalJWTProofFormat,
		})
		require.Error(t, err)
		require.Empty(t, jwtVC)
		require.ErrorIs(t, err, expectErr)
		require.Contains(t, err.Error(), "resolving verification method")
	})

	t.Run("jwt signer doesn't recognize VM type", func(t *testing.T) {
		vm := mockVM(t)
		vm.Type = "unknown verification method type"

		signer := New(&mockResolver{
			doc: makeDoc(vm),
		}, &mockCrypto{})

		jwtVC, err := signer.Issue(mockCredential, &ProofOptions{
			KeyID:       mockKID,
			ProofFormat: ExternalJWTProofFormat,
		})
		require.Error(t, err)
		require.Empty(t, jwtVC)
		require.Contains(t, err.Error(), "initializing jwt signer")
	})

	t.Run("fail to generate VC JWT claims", func(t *testing.T) {
		badCredential, cErr := verifiable.CreateCredential(verifiable.CredentialContents{
			ID:      "foo",
			Types:   []string{verifiable.VCType},
			Context: []string{verifiable.V1ContextURI},
			Subject: []verifiable.Subject{},
			Issuer: &verifiable.Issuer{
				ID: "did:foo:bar",
			},
			Issued: afgotime.NewTime(time.Now()),
		}, nil)
		require.NoError(t, cErr)

		signer := New(&mockResolver{
			doc: mockDoc(t),
		}, &mockCrypto{})

		jwtVC, err := signer.Issue(badCredential, &ProofOptions{
			KeyID:       mockKID,
			ProofFormat: ExternalJWTProofFormat,
		})
		require.Error(t, err)
		require.Empty(t, jwtVC)
		require.Contains(t, err.Error(), "failed to create JWT VC")
	})

	t.Run("signing error", func(t *testing.T) {
		signer := New(&mockResolver{
			doc: mockDoc(t),
		}, &mockCrypto{
			Err: expectErr,
		})

		jwtVC, err := signer.Issue(mockCredential, &ProofOptions{
			KeyID:       mockKID,
			ProofFormat: ExternalJWTProofFormat,
		})
		require.Error(t, err)
		require.Empty(t, jwtVC)
		require.ErrorIs(t, err, expectErr)
		require.Contains(t, err.Error(), "failed to create JWT VC")
	})
}

func mockDoc(t *testing.T) *did.Doc {
	t.Helper()

	return makeDoc(mockVM(t))
}

func mockVM(t *testing.T) *did.VerificationMethod {
	t.Helper()

	pkb, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	return &did.VerificationMethod{
		ID:         mockVMID,
		Controller: mockDID,
		Type:       "Ed25519VerificationKey2018",
		Value:      pkb,
	}
}

func makeDoc(vm *did.VerificationMethod) *did.Doc {
	return &did.Doc{
		ID: mockDID,
		AssertionMethod: []did.Verification{
			{
				VerificationMethod: *vm,
				Relationship:       did.AssertionMethod,
			},
		},
		VerificationMethod: []did.VerificationMethod{
			*vm,
		},
	}
}

type mockResolver struct {
	doc *did.Doc
	err error
}

func (m *mockResolver) Resolve(string) (*did.DocResolution, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &did.DocResolution{
		DIDDocument: m.doc,
	}, nil
}

type mockCrypto struct {
	Signature []byte
	Err       error
}

func (c *mockCrypto) Sign([]byte, string) ([]byte, error) {
	return c.Signature, c.Err
}

func (c *mockCrypto) Verify(_, _ []byte, _ string) error {
	return nil
}
