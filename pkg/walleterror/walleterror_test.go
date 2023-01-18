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
	require.Equal(t, "TEST_ERROR", err.Scenario)
	require.Equal(t, "error", err.ParentError)
}

func TestNewExecutionError(t *testing.T) {
	err := walleterror.NewExecutionError("AAA", 10, "TEST_ERROR", errors.New("error"))

	require.Equal(t, "AAA1-0010", err.Code)
	require.Equal(t, "TEST_ERROR", err.Scenario)
	require.Equal(t, "error", err.ParentError)
}

func TestNewSystemError(t *testing.T) {
	err := walleterror.NewSystemError("AAA", 10, "TEST_ERROR", errors.New("error"))

	require.Equal(t, "AAA2-0010", err.Code)
	require.Equal(t, "TEST_ERROR", err.Scenario)
	require.Equal(t, "error", err.ParentError)
	require.Equal(t, "TEST_ERROR(AAA2-0010):error", err.Error())
}
