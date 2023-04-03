/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper_test

import (
	"errors"
	"testing"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/walleterror"

	"github.com/stretchr/testify/require"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapiwalleterror "github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

func TestToMobileError(t *testing.T) {
	t.Run("Nil error", func(t *testing.T) {
		err := wrapper.ToMobileError(nil)
		require.NoError(t, err)
	})
	t.Run("Wallet error passed in", func(t *testing.T) {
		walletError := &goapiwalleterror.Error{
			Code:        "Code",
			Scenario:    "Category",
			ParentError: "Details",
		}

		err := wrapper.ToMobileError(walletError)
		require.Error(t, err)

		parsedErr := walleterror.Parse(err.Error())

		require.Equal(t, "Code", parsedErr.Code)
		require.Equal(t, "Category", parsedErr.Category)
		require.Equal(t, "Details", parsedErr.Details)
	})
	t.Run("goapiwalleterror.Error passed in", func(t *testing.T) {
		walletError := &goapiwalleterror.Error{
			Code:        "Code",
			Scenario:    "Category",
			ParentError: "Details",
		}

		err := wrapper.ToMobileError(walletError)
		require.Error(t, err)

		parsedErr := walleterror.Parse(err.Error())

		require.Equal(t, "Code", parsedErr.Code)
		require.Equal(t, "Category", parsedErr.Category)
		require.Equal(t, "Details", parsedErr.Details)
	})
	t.Run("Non-goapiwalleterror.Error passed in", func(t *testing.T) {
		goErr := errors.New("regular Go error")

		err := wrapper.ToMobileError(goErr)
		require.Error(t, err)

		parsedErr := walleterror.Parse(err.Error())

		require.Equal(t, "UKN2-000", parsedErr.Code)
		require.Equal(t, "UNEXPECTED_ERROR", parsedErr.Category)
		require.Equal(t, "regular Go error", parsedErr.Details)
	})
}
