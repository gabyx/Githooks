package prompt

import (
	"errors"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

func formatTitle(_ *Context) string {
	return cm.FormatInfoMessage("Githooks - Git Hook Manager")
}

func formatTitleQuestion(_ *Context) string {
	return cm.FormatPromptMessage("Githooks - Git Hook Manager")
}

// showMessage shows a message to the user with `text`.
func showMessage(p *Context, text string, asError bool) error {
	var err error

	m := "%s [any key to continue]: "

	if p.useGUI {
		// Use the GUI dialog.
		e := showMessageGUI(
			p,
			formatTitle(p),
			text, asError)

		if e == nil {
			return nil
		}

		err = cm.CombineErrors(err, e)

		if asError {
			p.log.InfoF(m, text)
		} else {
			p.log.ErrorF(m, text)
		}
	} else {
		// Use the terminal (if possible...)
		var message string
		if asError {
			message = cm.FormatErrorMessage(m, text)
		} else {
			message = cm.FormatInformationMessage(m, text)
		}

		isPromptDisplayed, e := showMessageTerminal(
			p,
			message,
			asError)

		if e == nil {
			return nil // is already in lower case
		}
		err = cm.CombineErrors(err, e)

		if !isPromptDisplayed {
			// Show the prompt in the log output
			p.log.Info(message)
		}
	}

	return err
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
	m := "%s %s [%s]: "

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
		p.log.Info(cm.FormatPromptMessage(m, text, hintText, shortOptions))
	} else {
		// Use the terminal (if possible...)
		emptyCausesDefault := strs.IsNotEmpty(defaultAnswer)
		question := cm.FormatPromptMessage(m, text, hintText, shortOptions)

		ans, isPromptDisplayed, e :=
			showOptionsTerminal(
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
	answerToLowerCase bool,
	validator AnswerValidator) (string, bool, error) {
	var err error // all errors

	nPrompts := uint(0) // How many times we showed the prompt
	maxPrompts := p.maxTries

	switch {
	case p.termIn == nil:
		err = cm.ErrorF("No terminal input available to read prompt answer.")
		return defaultAnswer, false, err //nolint:nlreturn
	case p.termOut == nil:
		err = cm.ErrorF("No terminal output available to show prompt.")
		return defaultAnswer, false, err //nolint:nlreturn
	}

	// Write to terminal output.
	writeOut := func(s string) {
		_, e := p.termOut.Write([]byte(s))
		err = cm.CombineErrors(err, e)
	}

	writeErr := func(s string) {
		_, e := p.termErr.Write([]byte(s))
		err = cm.CombineErrors(err, e)
	}

	for nPrompts < maxPrompts {
		_, e := p.termPrompt.Write([]byte(text))
		err = cm.CombineErrors(err, e)
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
		ans = strings.TrimSpace(ans)
		if answerToLowerCase {
			ans = strings.ToLower(ans)
		}

		// Validate the answer if possible.
		if validator == nil {
			return ans, true, nil
		}

		valErr := validator(ans)
		if valErr == nil {
			return ans, true, nil
		}

		writeErr(strs.Fmt("Answer validation error: %s\n", valErr.Error()))

		if nPrompts < maxPrompts {
			writeErr(strs.Fmt("Remaining tries %v.\n", maxPrompts-nPrompts))
		} else {
			msg := strs.Fmt("Could not validate answer in '%v' tries.", maxPrompts)
			if p.panicIfMaxTries {
				p.log.PanicF(msg)
			} else {
				return defaultAnswer, nPrompts != 0, valErr
			}
		}
	}

	writeErr(strs.Fmt("Could not get answer, taking default '%s'.\n", defaultAnswer))

	return defaultAnswer, nPrompts != 0, err
}

func showMessageTerminal(
	p *Context,
	message string,
	_ bool) (bool, error) {
	_, isPromptDisplayed, err := showEntryLoopTerminal(
		p,
		message,
		"",
		true,
		false,
		nil)

	return isPromptDisplayed, err
}

func showOptionsTerminal(
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
		true,
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

		p.log.Info(text)
	} else {
		if strs.IsNotEmpty(defaultAnswer) {
			text = cm.FormatPromptMessage("%s [%s]: ", text, defaultAnswer)
		} else {
			text = cm.FormatPromptMessage("%s : ", text)
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
	return showEntryLoopTerminal(p, text, defaultAnswer, true, false, validator)
}

func showEntryMulti(
	p *Context,
	text string,
	exitAnswer string,
	validator AnswerValidator) (answers []string, err error) {
	// Wrap the validator into another one which
	// reacts on `exitAnswer` for non-GUI prompts.
	var val = validator
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
			var validationError ValidationError
			if errors.As(err, &validationError) {
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

	return answers, err
}
