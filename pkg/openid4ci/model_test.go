/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package openid4ci_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/openid4ci"
)

func TestCredentialResponse_SerializeToCredentialsBytes(t *testing.T) {
	t.Run("Success jwt", func(t *testing.T) {
		credRes := &openid4ci.CredentialResponse{
			Credential: "test.jwt.sign",
		}
		res, err := credRes.SerializeToCredentialsBytes()
		require.NoError(t, err)
		require.Equal(t, "test.jwt.sign", string(res))
	})

	t.Run("Success json ld", func(t *testing.T) {
		credRes := &openid4ci.CredentialResponse{
			Credential: map[string]interface{}{
				"fld1": "val1",
				"fld2": "val2",
				"fld3": "val3",
			},
		}
		res, err := credRes.SerializeToCredentialsBytes()
		require.NoError(t, err)
		require.JSONEq(t, "{\"fld1\":\"val1\",\"fld2\":\"val2\",\"fld3\":\"val3\"}", string(res))
	})

	t.Run("Unsupported type", func(t *testing.T) {
		credRes := &openid4ci.CredentialResponse{
			Credential: make(chan int),
		}
		_, err := credRes.SerializeToCredentialsBytes()
		require.Error(t, err)
	})
}
