/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/walleterror"
	goapiwalleterror "github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// ToMobileError translates Go API errors to gomobile API errors.
func ToMobileError(err error) error {
	if err == nil {
		return nil
	}

	var result *walleterror.Error

	var walletError *goapiwalleterror.Error

	if errors.As(err, &walletError) {
		result = &walleterror.Error{
			Code:     walletError.Code,
			Category: walletError.Scenario,
			Details:  walletError.ParentError,
		}
	} else {
		result = &walleterror.Error{
			Code:     "UKN2-000",
			Category: "UNEXPECTED_ERROR",
			Details:  err.Error(),
		}
	}

	formatted, fmtErr := json.Marshal(result)
	if fmtErr != nil {
		return fmt.Errorf("failed to marshal error: %w", fmtErr)
	}

	return errors.New(string(formatted))
}
