package prompt

import (
	"errors"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

func formatTitle(p *Context) string {
	return p.log.GetInfoFormatter(false)("Githooks - Git Hook Manager")
}

func formatTitleQuestion(p *Context) string {
	return p.log.GetPromptFormatter(false)("Githooks - Git Hook Manager")
}

// showOptions shows a prompt to the user with `text`
// with the options `shortOptions` and optional long options `longOptions`.
func showOptions(
	p *Context,
	text string,
	hintText string,
	shortOptions string,
	longOptions ...string) (string, error) {

	var err error
	options := strings.Split(shortOptions, "/")
	validator := CreateValidatorAnswerOptions(options)

	defaultAnswer, defaultOptionIdx := getDefaultAnswer(options)

	if p.useGUI {
		// Use the GUI dialog.
		ans, e := showOptionsGUI(
			p,
			formatTitleQuestion(p),
			text,
			defaultAnswer,
			defaultOptionIdx,
			options, longOptions,
			validator)

		if e == nil {
			return strings.ToLower(ans), nil
		}

		err = cm.CombineErrors(err, e)
		p.log.Info(p.promptFmt("%s %s [%s]: ", text, hintText, shortOptions))

	} else {
		// Use the terminal (if possible...)
		emptyCausesDefault := strs.IsNotEmpty(defaultAnswer)
		question := p.promptFmt("%s %s [%s]: ", text, hintText, shortOptions)

		ans, isPromptDisplayed, e :=
			showPromptOptionsTerminal(
				p,
				question,
				defaultAnswer,
				emptyCausesDefault,
				validator)

		if e == nil {
			return ans, nil // is already in lower case
		}
		err = cm.CombineErrors(err, e)

		if !isPromptDisplayed {
			// Show the prompt in the log output
			p.log.Info(question)
		}
	}

	return defaultAnswer, err
}

func showEntryLoopTerminal(
	p *Context,
	text string,
	defaultAnswer string,
	emptyCausesDefault bool,
	validator AnswerValidator) (string, bool, error) {

	var err error // all errors

	nPrompts := uint(0) // How many times we showed the prompt
	maxPrompts := p.maxTries

	switch {
	case p.termIn == nil:
		err = cm.ErrorF("No terminal input available to read prompt answer.")
		return defaultAnswer, false, err // nolint: nlreturn
	case p.termOut == nil:
		err = cm.ErrorF("No terminal output available to show prompt.")
		return defaultAnswer, false, err // nolint: nlreturn
	}

	// Write to terminal output.
	writeOut := func(s string) {
		_, e := p.termOut.Write([]byte(s))
		err = cm.CombineErrors(err, e)
	}

	for nPrompts < maxPrompts {

		writeOut(text)
		nPrompts++

		success := p.termInScanner.Scan()

		if !success {
			writeOut("\n")
			err = cm.CombineErrors(err, cm.ErrorF("Could not read from terminal."))

			break
		}

		ans := p.termInScanner.Text()

		if p.printAnswer {
			writeOut(strs.Fmt(" -> Received: '%s'\n", ans))
		}

		// Fallback to default answer.
		if strs.IsEmpty(ans) && emptyCausesDefault {
			ans = defaultAnswer
		}

		// Trim everything.
		ans = strings.ToLower(strings.TrimSpace(ans))

		// Validate the answer if possible.
		if validator == nil {
			return ans, true, nil
		}

		valErr := validator(ans)
		if valErr == nil {
			return ans, true, nil
		}

		warning := p.errorFmt("Answer validation error: %s", valErr.Error())
		writeOut(warning + "\n")

		if nPrompts < maxPrompts {
			warning := p.errorFmt("Remaining tries %v.", maxPrompts-nPrompts)
			writeOut(warning + "\n")
		} else {
			msg := strs.Fmt("Could not validate answer in '%v' tries.", maxPrompts)
			if p.panicIfMaxTries {
				p.log.PanicF(msg)
			} else {
				return defaultAnswer, nPrompts != 0, valErr
			}
		}
	}

	warning := p.errorFmt("Could not get answer, taking default '%s'.", defaultAnswer)
	writeOut(warning + "\n")

	return defaultAnswer, nPrompts != 0, err
}

func showPromptOptionsTerminal(
	p *Context,
	question string,
	defaultAnswer string,
	emptyCausesDefault bool,
	validator AnswerValidator) (string, bool, error) {

	return showEntryLoopTerminal(
		p,
		question,
		defaultAnswer,
		emptyCausesDefault,
		validator)
}

// showEntry shows a prompt to the user with `text`.
func showEntry(
	p *Context,
	text string,
	defaultAnswer string,
	validator func(string) error,
	canCancel bool) (ans string, err error) {

	if p.useGUI {
		ans, err = showEntryGUI(p, formatTitle(p), text, defaultAnswer, validator, canCancel)
		if err == nil {
			return
		}

		if canCancel && errors.Is(err, ErrorCanceled) {
			return defaultAnswer, err
		}

	} else {

		if strs.IsNotEmpty(defaultAnswer) {
			text = p.promptFmt("%s [%s]: ", text, defaultAnswer)
		} else {
			text = p.promptFmt("%s : ", text)
		}

		var isPromptDisplayed bool
		ans, isPromptDisplayed, err =
			showEntryTerminal(p, text, defaultAnswer, validator)

		if err == nil {
			return
		}

		if !isPromptDisplayed {
			// Show the prompt in the log output
			p.log.Info(text)
		}
	}

	return defaultAnswer, err
}

func showEntryTerminal(
	p *Context,
	text string,
	defaultAnswer string,
	validator AnswerValidator) (string, bool, error) {
	return showEntryLoopTerminal(p, text, defaultAnswer, true, validator)
}

func showEntryMulti(
	p *Context,
	text string,
	exitAnswer string,
	validator AnswerValidator) (answers []string, err error) {

	// Wrap the validator into another one which
	// reacts on `exitAnswer` for non-GUI prompts.
	var val AnswerValidator = validator
	exitReceived := false

	if !p.useGUI {

		if strs.IsEmpty(exitAnswer) {
			text += " [empty cancels]"
		} else {
			text += strs.Fmt(" ['%s' cancels]", exitAnswer)
		}

		val = func(s string) error {
			if s == exitAnswer {
				exitReceived = true

				return nil
			}

			return validator(s)
		}
	}

	var ans string

	for {
		ans, err = showEntry(p, text, "", val, true)

		if err != nil {

			if _, ok := err.(ValidationError); ok {
				continue
			} else if errors.Is(err, ErrorCanceled) {
				err = nil

				break
			}

			break

		} else if exitReceived {
			break
		}

		// Add the entry.
		answers = append(answers, ans)
	}

	return
}
