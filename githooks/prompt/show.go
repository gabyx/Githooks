//go:build !mock

package prompt

// AssertOutputIsTerminal asserts that in some input/output redirection cases, the output
// is really a terminal.
const AssertOutputIsTerminal = true

// EnableGUI in general enables the GUI dialogs
// For testing this is disabled.
var EnableGUI = true

// ShowMessage shows a message to the user with `text`.
func (p *Context) ShowMessage(text string, asError bool) (err error) {
	return showMessage(p, text, asError)
}

// ShowOptions shows a prompt to the user with `text`
// with the options `shortOptions` and optional long options `longOptions`.
func (p *Context) ShowOptions(text string,
	hintText string,
	shortOptions string,
	longOptions ...string) (answer string, err error) {
	return showOptions(p, text, hintText, shortOptions, longOptions...)
}

// ShowEntry shows a prompt to enter an answer and
// validates it with a validator.
func (p *Context) ShowEntry(
	text string,
	defaultAnswer string,
	validator AnswerValidator) (answer string, err error) {
	return showEntry(p, text, defaultAnswer, validator, false)
}

// ShowEntryMulti shows multiple prompts to enter multiple answers and
// validates it with a validator. An empty answer exits the prompt.
func (p *Context) ShowEntryMulti(
	text string,
	exitAnswer string,
	validator AnswerValidator) (answers []string, err error) {
	return showEntryMulti(p, text, exitAnswer, validator)
}
