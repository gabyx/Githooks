// +build mock

package prompt

import (
	"os"
	"strings"
)

const AssertOutputIsTerminal = false

// ShowPromptOptions mocks the real ShowPromptOptions by reading
// from the environment or if not defined calls the normal implementation.
// This is only for tests.
func (p *Context) ShowPromptOptions(text string,
	hintText string,
	shortOptions string,
	longOptions ...string) (answer string, err error) {

	if strings.Contains(text, "This repository wants you to trust all current") {
		answer, defined := os.LookupEnv("TRUST_ALL_HOOKS")
		if defined {
			return strings.ToLower(answer), nil
		}
	} else if strings.Contains(text, "Do you accept the changes") {
		answer, defined := os.LookupEnv("ACCEPT_CHANGES")
		if defined {
			return strings.ToLower(answer), nil
		}
	} else if strings.Contains(text, "There is a new Githooks update available") {
		answer, defined := os.LookupEnv("EXECUTE_UPDATE")
		if defined {
			return strings.ToLower(answer), nil
		}
	}

	return showPromptOptions(p, text, hintText, shortOptions, longOptions...)
}

// ShowPrompt mocks the real ShowPrompt by reading
// from the environment or if not defined calls the normal implementation.
// This is only for tests.
func (p *Context) ShowPrompt(
	text string,
	defaultAnswer string,
	validator AnswerValidator) (answer string, err error) {
	return showPrompt(p, text, defaultAnswer, validator)
}

// ShowPromptMulti shows multiple prompts to enter multiple answers and
// validates it with a validator. An empty answer exits the prompt.
func (p *Context) ShowPromptMulti(
	text string,
	exitAnswer string,
	validator AnswerValidator) (answers []string, err error) {
	return showPromptMulti(p, text, exitAnswer, validator)
}
