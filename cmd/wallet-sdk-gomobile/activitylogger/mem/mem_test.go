/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mem_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/activitylogger/mem"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/api"
	goapi "github.com/trustbloc/wallet-sdk/pkg/api"
)

func TestActivityLogger(t *testing.T) {
	activityLogger := mem.NewActivityLogger()

	activity1 := &api.Activity{
		GoAPIActivity: &goapi.Activity{ID: uuid.New()},
	}

	err := activityLogger.Log(activity1)
	require.NoError(t, err)

	activity2 := &api.Activity{
		GoAPIActivity: &goapi.Activity{ID: uuid.New()},
	}

	err = activityLogger.Log(activity2)
	require.NoError(t, err)

	numberOfActivitiesLogged := activityLogger.Length()
	require.Equal(t, 2, numberOfActivitiesLogged)

	retrievedActivity1 := activityLogger.AtIndex(0)
	require.Equal(t, activity1, retrievedActivity1)

	retrievedActivity2 := activityLogger.AtIndex(1)
	require.Equal(t, activity2, retrievedActivity2)

	nonExistentActivity := activityLogger.AtIndex(2)
	require.Nil(t, nonExistentActivity)

	nonExistentActivity = activityLogger.AtIndex(-1)
	require.Nil(t, nonExistentActivity)
}
