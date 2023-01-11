package walleterror

import "fmt"

type Error struct {
	Code      string
	ErrorName string
	Details   string
}

func New(code, error string, parentError error) *Error {
	return &Error{
		Code:      code,
		ErrorName: error,
		Details:   parentError.Error(),
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s(%s):%s", e.ErrorName, e.Code, e.Details)
}
