/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walleterror_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

func TestNewValidationError(t *testing.T) {
	err := walleterror.NewValidationError("AAA", 10, "TEST_ERROR", errors.New("error"))

	require.Equal(t, "AAA0-0010", err.Code)
	require.Equal(t, "TEST_ERROR", err.Category)
	require.Equal(t, "error", err.ParentError)
}

func TestNewExecutionError(t *testing.T) {
	err := walleterror.NewExecutionError("AAA", 10, "TEST_ERROR", errors.New("error"))

	require.Equal(t, "AAA1-0010", err.Code)
	require.Equal(t, "TEST_ERROR", err.Category)
	require.Empty(t, err.Message)
	require.Equal(t, "error", err.ParentError)
}

func TestNewExecutionErrorWithMessage(t *testing.T) {
	err := walleterror.NewExecutionErrorWithMessage("AAA", 10, "TEST_ERROR", "message",
		errors.New("error"))

	require.Equal(t, "AAA1-0010", err.Code)
	require.Equal(t, "TEST_ERROR", err.Category)
	require.Equal(t, "message", err.Message)
	require.Equal(t, "error", err.ParentError)
}

func TestNewExecutionErrorWithServerErrorCode(t *testing.T) {
	err := walleterror.NewExecutionError("AAA", 10, "TEST_ERROR", errors.New("error"),
		walleterror.WithServerErrorCode("server-code"),
		walleterror.WithServerErrorMessage("server-message"))

	require.Equal(t, "AAA1-0010", err.Code)
	require.Equal(t, "TEST_ERROR", err.Category)
	require.Empty(t, err.Message)
	require.Equal(t, "error", err.ParentError)
	require.Equal(t, "server-code", err.ServerCode)
	require.Equal(t, "server-message", err.ServerMessage)
}

func TestNewInvalidSDKUsageError(t *testing.T) {
	err := walleterror.NewInvalidSDKUsageError("AAA", errors.New("error"))

	require.Equal(t, "AAA3-0000", err.Code)
	require.Equal(t, "INVALID_SDK_USAGE", err.Category)
	require.Equal(t, "error", err.ParentError)
}

func TestNewSystemError(t *testing.T) {
	err := walleterror.NewSystemError("AAA", 10, "TEST_ERROR", errors.New("error"))

	require.Equal(t, "AAA2-0010", err.Code)
	require.Equal(t, "TEST_ERROR", err.Category)
	require.Equal(t, "error", err.ParentError)
	require.Equal(t, "TEST_ERROR(AAA2-0010):error", err.Error())
}
