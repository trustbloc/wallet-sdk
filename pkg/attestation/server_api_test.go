/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package attestation //nolint: testpackage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/internal/mock"
	noopmetricslogger "github.com/trustbloc/wallet-sdk/pkg/metricslogger/noop"
)

func TestAttestationInitAPI(t *testing.T) {
	sampleRequest := AttestWalletInitRequest{
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

	sampleResp := `{"challenge": "challenge", "session_id":"session_id"}`

	t.Run("Success", func(t *testing.T) {
		resultReq := AttestWalletInitRequest{}

		client := &mock.HTTPClientMock{
			StatusCode:          200,
			ExpectedEndpoint:    "https://testurl/init",
			Response:            sampleResp,
			SentBodyUnMarshaled: &resultReq,
		}

		api := &serverAPI{httpClient: client, metricsLogger: noopmetricslogger.NewMetricsLogger()}

		resp, err := api.AttestationInit(sampleRequest, "https://testurl")
		require.NoError(t, err)
		require.Equal(t, "challenge", resp.Challenge)
		require.Equal(t, "session_id", resp.SessionID)
	})

	t.Run("Request failed", func(t *testing.T) {
		client := &mock.HTTPClientMock{
			StatusCode: 500,
		}

		api := &serverAPI{httpClient: client, metricsLogger: noopmetricslogger.NewMetricsLogger()}

		_, err := api.AttestationInit(sampleRequest, "https://testurl")
		require.ErrorContains(t, err, "expected status code 200 but got status code 500")
	})

	t.Run("Marshal failed", func(t *testing.T) {
		client := &mock.HTTPClientMock{
			StatusCode: 200,
		}

		api := &serverAPI{httpClient: client, metricsLogger: noopmetricslogger.NewMetricsLogger()}

		_, err := api.AttestationInit(AttestWalletInitRequest{
			Payload: map[string]interface{}{
				"sss": make(chan string),
			},
		},
			"https://testurl",
		)
		require.ErrorContains(t, err, "json: unsupported type")
	})
}

func TestAttestationCompletedAPI(t *testing.T) {
	sampleRequest := AttestWalletCompleteRequest{
		AssuranceLevel: "low",
		Proof: Proof{
			Jwt:       "jwt",
			ProofType: "ptype",
		},
		SessionID: "123",
	}

	sampleResp := `{"wallet_attestation_vc": "wallet_attestation_vc"}`

	t.Run("Success", func(t *testing.T) {
		resultReq := AttestWalletCompleteRequest{}

		client := &mock.HTTPClientMock{
			StatusCode:          200,
			ExpectedEndpoint:    "https://testurl/complete",
			Response:            sampleResp,
			SentBodyUnMarshaled: &resultReq,
		}

		api := &serverAPI{httpClient: client, metricsLogger: noopmetricslogger.NewMetricsLogger()}

		resp, err := api.AttestationComplete(sampleRequest, "https://testurl")
		require.NoError(t, err)
		require.Equal(t, "wallet_attestation_vc", resp.WalletAttestationVC)
	})

	t.Run("Request failed", func(t *testing.T) {
		client := &mock.HTTPClientMock{
			StatusCode: 500,
		}

		api := &serverAPI{httpClient: client, metricsLogger: noopmetricslogger.NewMetricsLogger()}

		_, err := api.AttestationComplete(sampleRequest, "https://testurl")
		require.ErrorContains(t, err, "expected status code 200 but got status code 500")
	})
}
