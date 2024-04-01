/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation_test

import (
	_ "embed"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/kms-go/doc/jose"
	"github.com/trustbloc/vc-go/jwt"

	"github.com/trustbloc/wallet-sdk/internal/testutil"
	"github.com/trustbloc/wallet-sdk/pkg/attestation"
)

const (
	mockKeyID = "did:test:foo#testId"
)

//go:embed testdata/sample-jwt.jwt
var sampleJWT []byte

func TestNewClient(t *testing.T) {
	lddl := testutil.DocumentLoader(t)

	client := attestation.NewClient(
		&attestation.ClientConfig{
			DocumentLoader: lddl,
			AttestationURL: "https://attestation.com",
		})

	require.NotNil(t, client)
}

func TestClient_GetAttestationVC(t *testing.T) {
	initReq := attestation.AttestWalletInitRequest{
		Payload: map[string]interface{}{
			"type": "urn:attestation:application:trustbloc",
			"application": map[string]interface{}{
				"type":    "wallet-cli",
				"name":    "wallet-cli",
				"version": "1.0",
			},
			"compliance": []interface{}{
				map[string]interface{}{
					"type": "fcra",
				},
			},
		},
	}

	t.Run("test success", func(t *testing.T) {
		lddl := testutil.DocumentLoader(t)

		client := attestation.NewClient(
			&attestation.ClientConfig{
				DocumentLoader: lddl,
				AttestationURL: "https://attestation.com",
				CustomAPI: &serverAPIMock{
					attestationInitResp: &attestation.AttestWalletInitResponse{
						Challenge: "1224",
						SessionID: "7446",
					},
					attestationCompleteResp: &attestation.AttestWalletCompleteResponse{
						WalletAttestationVC: string(sampleJWT),
					},
				},
			})

		_, err := client.GetAttestationVC(initReq, &jwtSignerMock{
			keyID: mockKeyID,
		})
		require.NoError(t, err)
	})

	t.Run("Sign failed", func(t *testing.T) {
		lddl := testutil.DocumentLoader(t)

		client := attestation.NewClient(
			&attestation.ClientConfig{
				DocumentLoader: lddl,
				AttestationURL: "https://attestation.com",
				CustomAPI: &serverAPIMock{
					attestationInitResp: &attestation.AttestWalletInitResponse{
						Challenge: "1224",
						SessionID: "7446",
					},
					attestationCompleteResp: &attestation.AttestWalletCompleteResponse{
						WalletAttestationVC: string(sampleJWT),
					},
				},
			})

		_, err := client.GetAttestationVC(initReq, &jwtSignerMock{
			keyID: mockKeyID,
			Err:   errors.New("sign error"),
		})
		require.ErrorContains(t, err, "JWT_SIGNING_FAILED")
	})

	t.Run("Invalid did", func(t *testing.T) {
		lddl := testutil.DocumentLoader(t)

		client := attestation.NewClient(
			&attestation.ClientConfig{
				DocumentLoader: lddl,
				AttestationURL: "https://attestation.com",
				CustomAPI: &serverAPIMock{
					attestationInitResp: &attestation.AttestWalletInitResponse{
						Challenge: "1224",
						SessionID: "7446",
					},
					attestationCompleteResp: &attestation.AttestWalletCompleteResponse{
						WalletAttestationVC: string(sampleJWT),
					},
				},
			})

		_, err := client.GetAttestationVC(initReq, &jwtSignerMock{
			keyID: "invalid",
		})

		require.ErrorContains(t, err, "KEY_ID_MISSING_DID_PART")
	})

	t.Run("test success", func(t *testing.T) {
		lddl := testutil.DocumentLoader(t)

		client := attestation.NewClient(
			&attestation.ClientConfig{
				DocumentLoader: lddl,
				AttestationURL: "https://attestation.com",
				CustomAPI: &serverAPIMock{
					attestationInitResp: &attestation.AttestWalletInitResponse{
						Challenge: "1224",
						SessionID: "7446",
					},
					attestationCompleteResp: &attestation.AttestWalletCompleteResponse{
						WalletAttestationVC: "inv",
					},
				},
			})

		_, err := client.GetAttestationVC(initReq, &jwtSignerMock{
			keyID: mockKeyID,
		})
		require.ErrorContains(t, err, "PARSE_ATTESTATION_VC_FAILED")
	})
}

type serverAPIMock struct {
	attestationInitResp *attestation.AttestWalletInitResponse
	attestationInitErr  error

	attestationCompleteResp *attestation.AttestWalletCompleteResponse
	attestationCompleteErr  error
}

func (a *serverAPIMock) AttestationInit(
	_ attestation.AttestWalletInitRequest,
	_ string,
) (*attestation.AttestWalletInitResponse, error) {
	if a.attestationInitErr != nil {
		return nil, a.attestationInitErr
	}

	return a.attestationInitResp, nil
}

func (a *serverAPIMock) AttestationComplete(
	_ attestation.AttestWalletCompleteRequest,
	_ string,
) (*attestation.AttestWalletCompleteResponse, error) {
	if a.attestationCompleteErr != nil {
		return nil, a.attestationCompleteErr
	}

	return a.attestationCompleteResp, nil
}

type jwtSignerMock struct {
	keyID string
	Err   error
}

func (s *jwtSignerMock) GetKeyID() string {
	return s.keyID
}

func (s *jwtSignerMock) SignJWT(_ jwt.SignParameters, _ []byte) ([]byte, error) {
	return []byte("test signature"), s.Err
}

func (s *jwtSignerMock) CreateJWTHeaders(_ jwt.SignParameters) (jose.Headers, error) {
	return jose.Headers{
		jose.HeaderKeyID:     "KeyID",
		jose.HeaderAlgorithm: "ES384",
	}, nil
}
