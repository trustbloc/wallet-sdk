/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	"github.com/trustbloc/wallet-sdk/pkg/api"
)

func TestMobileActivityLoggerWrapper_Log(t *testing.T) {
	activityLogger := &wrapper.MobileActivityLoggerWrapper{MobileAPIActivityLogger: mem.NewActivityLogger()}

	activity := &api.Activity{
		ID:   uuid.New(),
		Type: api.LogTypeCredentialActivity,
		Time: time.Now(),
		Data: api.Data{
			Client:    "Client",
			Operation: "Operation",
			Status:    "Status",
			Params:    map[string]interface{}{"Param": "Value"},
		},
	}

	err := activityLogger.Log(activity)
	require.NoError(t, err)
}
