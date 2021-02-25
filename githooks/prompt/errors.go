package prompt

import (
	"errors"
	cm "gabyx/githooks/common"
)

// ValidationError represents a validation error.
type ValidationError struct {
	error
}

func NewValidationError(format string, args ...interface{}) ValidationError {
	return ValidationError{cm.ErrorF(format, args...)}
}

var ErrorCanceled = errors.New("Cancled")
