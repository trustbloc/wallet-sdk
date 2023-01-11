package walleterror

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/trustbloc/wallet-sdk/pkg/walleterror"
)

type Error struct {
	Code    string `json:"code"`
	Error   string `json:"error"`
	Details string `json:"details"`
}

func Translate(err error) error {
	var result *Error

	var walletError *walleterror.Error

	if errors.As(err, &walletError) {
		result = &Error{
			Code:    walletError.Code,
			Error:   walletError.ErrorName,
			Details: walletError.Details,
		}
	} else {
		result = &Error{
			Code:    "UKN2-000",
			Error:   "UNEXPECTED_ERROR",
			Details: err.Error(),
		}
	}

	formatted, fmtErr := json.Marshal(result)
	if fmtErr != nil {
		return fmt.Errorf("unexpected format error: %w", fmtErr)
	}

	return errors.New(string(formatted))
}

func Parse(message string) *Error {
	walletErr := &Error{}
	err := json.Unmarshal([]byte(message), walletErr)
	if err != nil {
		return &Error{
			Code:    "GNR2-000",
			Error:   "GENERAL_ERROR",
			Details: message,
		}
	}
	return walletErr
}
