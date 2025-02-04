/*
Copyright Gen Digital Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package wrapper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/otel"
	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/walleterror"
	goapiwalleterror "github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

// ToMobileError translates Go API errors to gomobile API errors.
func ToMobileError(err error) error {
	return ToMobileErrorWithTrace(err, nil)
}

// ToMobileErrorWithTrace translates Go API errors to gomobile API errors.
func ToMobileErrorWithTrace(err error, trace *otel.Trace) error {
	if err == nil {
		return nil
	}

	gomobileError := convertToGomobileError(err, trace)

	marshalledGomobileError, err := json.Marshal(gomobileError)
	if err != nil {
		return fmt.Errorf("failed to marshal error: %w", err)
	}

	return errors.New(string(marshalledGomobileError))
}

func convertToGomobileError(err error, trace *otel.Trace) *walleterror.Error {
	traceID := ""
	if trace != nil {
		traceID = trace.TraceID()
	}

	errorToExamine := err

	for errorToExamine != nil {
		//nolint:errorlint // The linter wants us to use errors.As here, but we need to know the precise spot
		// in the error chain where the higher-level non-goapiwalleterror.Errors end.
		walletError, ok := errorToExamine.(*goapiwalleterror.Error)
		if ok {
			// If the highest-level error message it itself a goapiwalleterror.Error, then higherLevelErrorMessage
			// will be a blank string, so walletError.ParentError will simply be passed through.
			higherLevelErrorMessage := strings.ReplaceAll(err.Error(), walletError.Error(), "")

			mergedErrorMessage := higherLevelErrorMessage + walletError.ParentError

			return &walleterror.Error{
				Code:          walletError.Code,
				Category:      walletError.Category,
				Message:       walletError.Message,
				Details:       mergedErrorMessage,
				TraceID:       traceID,
				ServerCode:    walletError.ServerCode,
				ServerMessage: walletError.ServerMessage,
			}
		}

		errorToExamine = errors.Unwrap(errorToExamine)
	}

	// There's no wallet Error in the chain, and so there's no error code available. Let's create a new
	// gomobile wallet error using a generic code.
	return &walleterror.Error{
		Code:     "UKN2-000",
		Category: "OTHER_ERROR",
		Details:  err.Error(),
		TraceID:  traceID,
	}
}
