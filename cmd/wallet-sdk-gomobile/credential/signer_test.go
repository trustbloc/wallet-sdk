/*
Copyright Avast Software. All Rights Reserved.
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credential_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/did-go/doc/did"
	afgotime "github.com/trustbloc/did-go/doc/util/time"
	afgoverifiable "github.com/trustbloc/vc-go/verifiable"

	. "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/verifiable"
)

const (
	credID   = "foo-cred"
	mockDID  = "did:test:foo"
	mockVMID = "#key-1"
	mockKID  = mockDID + mockVMID
)

func TestSigner_Issue(t *testing.T) {
	expectErr := errors.New("expected error")

	mockCredential, err := afgoverifiable.CreateCredential(afgoverifiable.CredentialContents{
		ID:      credID,
		Types:   []string{afgoverifiable.VCType},
		Context: []string{afgoverifiable.V1ContextURI},
		Subject: []afgoverifiable.Subject{{
			ID: "foo",
		}},
		Issuer: &afgoverifiable.Issuer{
			ID: mockDID,
		},
		Issued: afgotime.NewTime(time.Now()),
	}, nil)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		t.Run("given raw credential", func(t *testing.T) {
			s := NewSigner(
				&mockResolver{ResolveVal: mockDocResolution(t)},
				&mockCrypto{SignVal: []byte("foo")},
			)

			issuedCred, err := s.Issue(verifiable.NewCredential(mockCredential), mockKID)
			require.NoError(t, err)
			require.NotNil(t, issuedCred)
		})
	})

	t.Run("failure", func(t *testing.T) {
		t.Run("signing credential", func(t *testing.T) {
			s := NewSigner(
				&mockResolver{ResolveVal: mockDocResolution(t)},
				&mockCrypto{SignErr: expectErr},
			)

			_, err := s.Issue(verifiable.NewCredential(mockCredential), "did:test:foo#key-1")
			require.Error(t, err)
			require.Contains(t, err.Error(), "signing credential")
			require.ErrorIs(t, err, expectErr)
		})
	})
}

func mockDocResolution(t *testing.T) []byte {
	t.Helper()

	docRes := &did.DocResolution{
		DIDDocument: makeDoc(mockVM(t)),
	}

	docBytes, err := docRes.JSONBytes()
	require.NoError(t, err)

	return docBytes
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
		ID:      mockDID,
		Context: did.ContextV1,
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
	ResolveVal []byte
	ResolveErr error
}

func (m *mockResolver) Resolve(string) ([]byte, error) {
	return m.ResolveVal, m.ResolveErr
}

type mockCrypto struct {
	SignVal   []byte
	SignErr   error
	VerifyErr error
}

func (m *mockCrypto) Sign([]byte, string) ([]byte, error) {
	return m.SignVal, m.SignErr
}

func (m *mockCrypto) Verify(_, _ []byte, _ string) error {
	return m.VerifyErr
}
