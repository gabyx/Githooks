package common

import (
	"errors"
	"fmt"
	"strings"

	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/hashicorp/go-multierror"
)

// GithooksFailure is a normal hook failure.
type GithooksFailure struct {
	error string
}

func (e *GithooksFailure) Error() string {
	return e.error
}

// Error makes an error message.
func Error(lines ...string) error {
	return errors.New(strings.Join(lines, "\n"))
}

// ErrorF makes an error message.
func ErrorF(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

func formatErrors(errors []error) string {
	var b strings.Builder

	l := len(errors)
	for i := range errors {
		b.WriteString(strs.Fmt("✗  %s", strings.ReplaceAll(errors[i].Error(), "\n", "\n  ")))
		if i+1 < l {
			b.WriteRune('\n')
		}
	}

	return b.String()
}

// FormatError formats an error.
func FormatError(err error) string {
	if e, ok := err.(*multierror.Error); ok {
		e.ErrorFormat = formatErrors

		return e.Error()
	}

	return err.Error()
}

// CombineErrors combines multiple errors into one.
func CombineErrors(err error, errs ...error) error {
	e := multierror.Append(err, errs...)
	e.ErrorFormat = formatErrors

	return e.ErrorOrNil()
}

// Panic panics with an `error`.
func Panic(lines ...string) {
	panic(Error(lines...))
}

// PanicF panics with an `error`.
func PanicF(format string, args ...interface{}) {
	panic(ErrorF(format, args...))
}

// AssertOrPanic Assert a condition is `true`, otherwise panic.
func AssertOrPanic(condition bool, lines ...string) {
	if !condition {
		Panic(lines...)
	}
}

// AssertOrPanicF Assert a condition is `true`, otherwise panic.
func AssertOrPanicF(condition bool, format string, args ...interface{}) {
	if !condition {
		PanicF(format, args...)
	}
}

// PanicIf Assert a condition is `true`, otherwise panic.
func PanicIf(condition bool, lines ...string) {
	if condition {
		Panic(lines...)
	}
}

// PanicIfF Assert a condition is `true`, otherwise panic.
func PanicIfF(condition bool, format string, args ...interface{}) {
	if condition {
		PanicF(format, args...)
	}
}

// AssertNoErrorPanic Assert no error, otherwise panic.
func AssertNoErrorPanic(err error, lines ...string) {
	if err != nil {
		PanicIf(true,
			append(lines, " -> errors:\n"+FormatError(err))...)
	}
}

// AssertNoErrorPanicF Assert no error, otherwise panic.
func AssertNoErrorPanicF(err error, format string, args ...interface{}) {
	if err != nil {
		PanicIfF(true, format+" -> errors:\n"+FormatError(err), args...)
	}
}
