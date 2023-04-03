/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4vp //nolint: testpackage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
)

func TestClientConfig_AddHeader(t *testing.T) {
	o := &Opts{}
	o.AddHeader(&api.Header{
		Name:  "testName",
		Value: "testValue",
	})

	headers := o.additionalHeaders.GetAll()
	require.Len(t, headers, 1)
	require.Equal(t, "testName", headers[0].Name)
	require.Equal(t, "testValue", headers[0].Value)
}
