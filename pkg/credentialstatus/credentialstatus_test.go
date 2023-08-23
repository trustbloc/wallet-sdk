/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package credentialstatus //nolint:testpackage // access internal fields

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/vc-go/verifiable"
)

func TestNewVerifier(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		v, err := NewVerifier(&Config{
			HTTPClient: http.DefaultClient,
		})
		require.NoError(t, err)
		require.NotNil(t, v)
	})
}

func TestVerifier_Verify(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		v := &Verifier{
			client: &mockStatusClient{},
		}

		err := v.Verify(&verifiable.Credential{})
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		expectErr := errors.New("expected error")

		v := &Verifier{
			client: &mockStatusClient{
				verifyErr: expectErr,
			},
		}

		err := v.Verify(&verifiable.Credential{})
		require.Error(t, err)
		require.ErrorIs(t, err, expectErr)
	})
}

type mockStatusClient struct {
	verifyErr error
}

func (s *mockStatusClient) VerifyStatus(*verifiable.Credential) error {
	return s.verifyErr
}
