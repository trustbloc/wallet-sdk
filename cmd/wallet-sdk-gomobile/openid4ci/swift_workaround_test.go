/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/openid4ci"
)

func TestConstructors_NewInteraction(t *testing.T) {
	issuerServerHandler := &mockIssuerServerHandler{t: t}
	server := httptest.NewServer(issuerServerHandler)

	defer server.Close()

	kms, err := localkms.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)

	config := getTestClientConfig(t, kms, nil)

	requestURI := createTestRequestURI(server.URL)

	constructors := openid4ci.Constructors{}

	interaction, err := constructors.NewInteraction(requestURI, config)
	require.NoError(t, err)
	require.NotNil(t, interaction)
}
