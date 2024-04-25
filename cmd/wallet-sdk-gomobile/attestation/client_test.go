/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation //nolint:testpackage

import (
	_ "embed"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	arieskms "github.com/trustbloc/kms-go/spi/kms"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/pkg/attestation"
	"github.com/trustbloc/wallet-sdk/pkg/models"
)

//go:embed testdata/sample-jwt.jwt
var sampleJWT []byte

func TestNewClient(t *testing.T) {
	t.Run("test success", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		c, err := NewClient(NewCreateClientArgs("https://attestation.com", kms.GetCrypto()))
		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("test success with doc loader", func(t *testing.T) {
		kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
		require.NoError(t, err)

		c, err := NewClient(
			NewCreateClientArgs("https://attestation.com", kms.GetCrypto()).
				SetDocumentLoader(&documentLoaderMock{}),
		)
		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("No args", func(t *testing.T) {
		_, err := NewClient(nil)
		require.ErrorContains(t, err, "args object must be provided")
	})
}

func TestClient_GetAttestationVC(t *testing.T) {
	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	keyHandle, err := kms.Create(arieskms.ED25519)
	require.NoError(t, err)

	verificationMethod := &api.VerificationMethod{
		ID:   "did:test:foo#abcd",
		Type: "JsonWebKey2020",
		Key:  models.VerificationKey{JSONWebKey: keyHandle.JWK},
	}

	correctAPI := &serverAPIMock{
		attestationInitResp: &attestation.AttestWalletInitResponse{
			Challenge: "1224",
			SessionID: "7446",
		},
		attestationCompleteResp: &attestation.AttestWalletCompleteResponse{
			WalletAttestationVC: string(sampleJWT),
		},
	}

	correctRequest := `
		{
			"type": "urn:attestation:application:trustbloc",
			"application": {
				"type":    "wallet-cli",
				"name":    "wallet-cli",
				"version": "1.0"
			},
			"compliance": {
				"type": "fcra"				
			}
		}
	`

	t.Run("test success", func(t *testing.T) {
		args := NewCreateClientArgs("https://attestation.com", kms.GetCrypto())
		args.customAPI = correctAPI

		c, err := NewClient(args)
		require.NoError(t, err)
		require.NotNil(t, c)

		vc, err := c.GetAttestationVC(verificationMethod, correctRequest)
		require.NoError(t, err)
		require.NotNil(t, vc)
	})

	t.Run("invalid payload", func(t *testing.T) {
		args := NewCreateClientArgs("https://attestation.com", kms.GetCrypto())
		args.customAPI = correctAPI

		c, err := NewClient(args)
		require.NoError(t, err)
		require.NotNil(t, c)

		vc, err := c.GetAttestationVC(verificationMethod, "{")
		require.ErrorContains(t, err, "parse attestation vc payload")
		require.Nil(t, vc)
	})

	t.Run("nil verificationMethod", func(t *testing.T) {
		args := NewCreateClientArgs("https://attestation.com", kms.GetCrypto())

		c, err := NewClient(args)
		require.NoError(t, err)
		require.NotNil(t, c)

		_, err = c.GetAttestationVC(nil, correctRequest)
		require.ErrorContains(t, err, "INVALID_SDK_USAGE")
	})

	t.Run("nil verificationMethod", func(t *testing.T) {
		args := NewCreateClientArgs("https://attestation.com", kms.GetCrypto())
		args.customAPI = &serverAPIMock{
			attestationInitErr:     errors.New("init error"),
			attestationCompleteErr: errors.New("complete error"),
		}

		c, err := NewClient(args)
		require.NoError(t, err)
		require.NotNil(t, c)

		_, err = c.GetAttestationVC(verificationMethod, correctRequest)
		require.ErrorContains(t, err, "init error")
	})
}

func TestNewCreateClientArgs(t *testing.T) {
	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	args := NewCreateClientArgs("https://attestation.com", kms.GetCrypto())
	require.NotNil(t, args)

	args.SetDocumentLoader(&documentLoaderMock{})
	require.NotNil(t, args.documentLoader)

	args.DisableOpenTelemetry()
	require.True(t, args.disableOpenTelemetry)

	args.SetHTTPTimeoutNanoseconds(100)
	require.Equal(t, 100*time.Nanosecond, *args.httpTimeout)

	args.DisableHTTPClientTLSVerify()
	require.True(t, args.disableHTTPClientTLSVerification)

	args.AddHeader(&api.Header{})
	require.Len(t, args.additionalHeaders.GetAll(), 1)

	args.SetMetricsLogger(&metricsLoggerMock{})
	require.NotNil(t, args.metricsLogger)
}

func TestGetAttestationPayloadHash(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		payload := `
		{
			"type": "urn:attestation:application:trustbloc",
			"application": {
				"type":    "wallet-cli",
				"name":    "wallet-cli",
				"version": "1.0"
			},
			"compliance": {
				"type": "fcra"				
			}
		}
	`

		hash, err := GetAttestationPayloadHash(payload)
		require.NoError(t, err)
		require.NotEmpty(t, hash)
	})

	t.Run("Invalid payload", func(t *testing.T) {
		_, err := GetAttestationPayloadHash("{")
		require.ErrorContains(t, err, "Unexpected EOF reached")
	})
}

type documentLoaderMock struct {
	LoadResult *api.LDDocument
	LoadErr    error
}

func (d *documentLoaderMock) LoadDocument(string) (*api.LDDocument, error) {
	return d.LoadResult, d.LoadErr
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

type metricsLoggerMock struct{}

func (m *metricsLoggerMock) Log(_ *api.MetricsEvent) error {
	return nil
}
