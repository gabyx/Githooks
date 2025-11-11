package prompt

import (
	"errors"

	cm "github.com/gabyx/githooks/githooks/common"
)

// ValidationError represents a validation error.
type ValidationError struct {
	error
}

// NewValidationError is the error for failed user input validation.
func NewValidationError(format string, args ...any) ValidationError {
	return ValidationError{cm.ErrorF(format, args...)}
}

// ErrorCanceled is the error for a canceled dialog.
var ErrorCanceled = errors.New("Cancled")
