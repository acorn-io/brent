package apierror

import (
	"fmt"

	"github.com/acorn-io/schemer/validation"
)

type APIError struct {
	Code      validation.ErrorCode
	Message   string
	Cause     error
	FieldName string
}

func NewAPIError(code validation.ErrorCode, message string) error {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// WrapAPIError will cause the API framework to log the underlying err before returning the APIError as a response.
// err WILL NOT be in the API response
func WrapAPIError(err error, code validation.ErrorCode, message string) error {
	return &APIError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

func (a *APIError) Error() string {
	if a.FieldName != "" {
		return fmt.Sprintf("%s=%s: %s", a.FieldName, a.Code, a.Message)
	}
	return fmt.Sprintf("%s: %s", a.Code, a.Message)
}
