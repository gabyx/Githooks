package prompt

import (
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"

	"strings"
)

func getDefaultAnswer(options []string) (string, int) {
	for idx, r := range options {
		if strings.ToLower(r) != r { // is it an upper case letter?
			return strings.ToLower(r), idx
		}
	}

	return "", -1
}

// CreateValidatorAnswerOptions creates a validator which validates against
// a list of options (case-insensitive).
func CreateValidatorAnswerOptions(options []string) AnswerValidator {
	return func(answer string) error {
		for i := range options {
			if strings.EqualFold(answer, options[i]) {
				return nil
			}
		}

		return NewValidationError("Answer '%s' not in '%q'.", answer, options)
	}
}

// ValidatorAnswerNotEmpty checks that answers are non-empty.
var ValidatorAnswerNotEmpty AnswerValidator = func(s string) error {
	if strs.IsEmpty(strings.TrimSpace(s)) {
		return NewValidationError("Answer must not be empty.")
	}

	return nil
}

// CreateValidatorIsDirectory creates a answer validator
// which checks existing paths.
func CreateValidatorIsDirectory(tildeRepl string) AnswerValidator {
	return func(s string) error {
		s = cm.ReplaceTildeWith(s, tildeRepl)
		if !cm.IsDirectory(s) {
			return NewValidationError("Answer must be an existing directory.")
		}

		return nil
	}
}
