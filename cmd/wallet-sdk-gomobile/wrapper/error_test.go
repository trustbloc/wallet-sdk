/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/walleterror"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/wrapper"
	goapiwalleterror "github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

func TestToMobileError(t *testing.T) {
	t.Run("Nil error", func(t *testing.T) {
		err := wrapper.ToMobileError(nil)
		require.NoError(t, err)
	})
	t.Run("goapiwalleterror.Error passed in", func(t *testing.T) {
		walletError := &goapiwalleterror.Error{
			Code:        "Code",
			Category:    "Category",
			Message:     "Message",
			ParentError: "Details",
		}

		err := wrapper.ToMobileError(walletError)
		require.Error(t, err)

		parsedErr := walleterror.Parse(err.Error())

		require.Equal(t, "Code", parsedErr.Code)
		require.Equal(t, "Category", parsedErr.Category)
		require.Equal(t, "Message", parsedErr.Message)
		require.Equal(t, "Details", parsedErr.Details)
	})
	t.Run("Non-goapiwalleterror.Error passed in", func(t *testing.T) {
		goErr := errors.New("regular Go error")

		err := wrapper.ToMobileError(goErr)
		require.Error(t, err)

		parsedErr := walleterror.Parse(err.Error())

		require.Equal(t, "UKN2-000", parsedErr.Code)
		require.Equal(t, "OTHER_ERROR", parsedErr.Category)
		require.Equal(t, "regular Go error", parsedErr.Details)
	})
	t.Run("Non-goapiwalleterror.Error wrapped with another Non-goapiwalleterror.Error passed in", func(t *testing.T) {
		goErr := fmt.Errorf("higher-level error: %w", errors.New("regular Go error"))

		err := wrapper.ToMobileError(goErr)
		require.Error(t, err)

		parsedErr := walleterror.Parse(err.Error())

		require.Equal(t, "UKN2-000", parsedErr.Code)
		require.Equal(t, "OTHER_ERROR", parsedErr.Category)
		require.Equal(t, "higher-level error: regular Go error", parsedErr.Details)
	})
	t.Run("goapiwalleterror.Error wrapped by one higher-level non-goapiwalleterror.Error is passed in",
		func(t *testing.T) {
			walletError := &goapiwalleterror.Error{
				Code:        "Code",
				Category:    "Category",
				ParentError: "Details",
			}

			wrappedWalletError := fmt.Errorf("higher-level error: %w", walletError)

			err := wrapper.ToMobileError(wrappedWalletError)
			require.Error(t, err)

			parsedErr := walleterror.Parse(err.Error())

			require.Equal(t, "Code", parsedErr.Code)
			require.Equal(t, "Category", parsedErr.Category)
			require.Equal(t, "higher-level error: Details", parsedErr.Details)
		})
	t.Run("goapiwalleterror.Error wrapped by two higher-level non-goapiwalleterror.Errors is passed in",
		func(t *testing.T) {
			walletError := &goapiwalleterror.Error{
				Code:        "Code",
				Category:    "Category",
				ParentError: "Details",
			}

			doubleWrappedWalletError := fmt.Errorf("even-higher-level error: %w",
				fmt.Errorf("higher-level error: %w", walletError))

			err := wrapper.ToMobileError(doubleWrappedWalletError)
			require.Error(t, err)

			parsedErr := walleterror.Parse(err.Error())

			require.Equal(t, "Code", parsedErr.Code)
			require.Equal(t, "Category", parsedErr.Category)
			require.Equal(t, "even-higher-level error: higher-level error: Details", parsedErr.Details)
		},
	)
	t.Run("goapiwalleterror.Error wrapped by another goapiwalleterror.Error", func(t *testing.T) {
		// Note: We shouldn't actually do this anywhere in our code. If this happens, then the highest-level
		// goapiwalleterror.Error is the one that will be detected and converted properly to the Gomobile error type,
		// while the lower one will get "squashed" into the Details field. This test just confirms that this is the
		// expected behaviour in such a scenario.
		lowerLevelWalletError := &goapiwalleterror.Error{
			Code:        "Lower-Level-Code",
			Category:    "Lower-Level-Category",
			ParentError: "Lower-Level-Details",
		}

		higherLevelWalletError := &goapiwalleterror.Error{
			Code:        "Higher-Level-Code",
			Category:    "Higher-Level-Category",
			ParentError: lowerLevelWalletError.Error(),
		}

		wrappedWalletError := fmt.Errorf("regular Go error: %w", higherLevelWalletError)

		err := wrapper.ToMobileError(wrappedWalletError)
		require.Error(t, err)

		parsedErr := walleterror.Parse(err.Error())

		require.Equal(t, "Higher-Level-Code", parsedErr.Code)
		require.Equal(t, "Higher-Level-Category", parsedErr.Category)
		require.Equal(t, "regular Go error: Lower-Level-Category(Lower-Level-Code):Lower-Level-Details",
			parsedErr.Details)
	})
}
