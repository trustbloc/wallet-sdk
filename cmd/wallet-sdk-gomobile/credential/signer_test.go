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

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/util"
	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	"github.com/stretchr/testify/require"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	. "github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/credential"
)

const (
	credID   = "foo-cred"
	mockDID  = "did:test:foo"
	mockVMID = "#key-1"
	mockKID  = mockDID + mockVMID
)

func TestSigner_Issue(t *testing.T) {
	expectErr := errors.New("expected error")

	mockCredential := &verifiable.Credential{
		ID:      credID,
		Types:   []string{verifiable.VCType},
		Context: []string{verifiable.ContextURI},
		Subject: verifiable.Subject{
			ID: "foo",
		},
		Issuer: verifiable.Issuer{
			ID: mockDID,
		},
		Issued: util.NewTime(time.Now()),
	}

	t.Run("success", func(t *testing.T) {
		t.Run("given raw credential", func(t *testing.T) {
			s, err := NewSigner(
				&mockReader{},
				&mockResolver{ResolveVal: mockDocResolution(t)},
				&mockCrypto{SignVal: []byte("foo")},
			)
			require.NoError(t, err)

			issuedCred, err := s.Issue(api.NewVerifiableCredential(mockCredential), "", mockKID)
			require.NoError(t, err)
			require.NotNil(t, issuedCred)
		})

		t.Run("given credential ID", func(t *testing.T) {
			s, err := NewSigner(
				&mockReader{getVal: api.NewVerifiableCredential(mockCredential)},
				&mockResolver{ResolveVal: mockDocResolution(t)},
				&mockCrypto{SignVal: []byte("foo")},
			)
			require.NoError(t, err)

			issuedCred, err := s.Issue(nil, credID, mockKID)
			require.NoError(t, err)
			require.NotNil(t, issuedCred)
		})
	})

	t.Run("failure", func(t *testing.T) {
		t.Run("signing credential", func(t *testing.T) {
			s, err := NewSigner(
				&mockReader{},
				&mockResolver{ResolveVal: mockDocResolution(t)},
				&mockCrypto{SignErr: expectErr},
			)
			require.NoError(t, err)

			_, err = s.Issue(api.NewVerifiableCredential(mockCredential), "", "")
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

type mockReader struct {
	getVal    *api.VerifiableCredential
	getAllVal *api.VerifiableCredentialsArray
	getErr    error
}

func (m *mockReader) Get(string) (*api.VerifiableCredential, error) {
	return m.getVal, m.getErr
}

func (m *mockReader) GetAll() (*api.VerifiableCredentialsArray, error) {
	return m.getAllVal, m.getErr
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
