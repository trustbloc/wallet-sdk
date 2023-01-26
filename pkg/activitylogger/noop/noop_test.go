/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package noop_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/activitylogger/noop"
)

func TestActivityLogger(t *testing.T) {
	activityLogger := noop.NewActivityLogger()

	err := activityLogger.Log(nil)
	require.NoError(t, err)
}
