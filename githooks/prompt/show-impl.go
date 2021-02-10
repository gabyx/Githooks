package prompt

import (
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
	"os"
	"strings"
)

// showPromptOptions shows a prompt to the user with `text`
// with the options `shortOptions` and optional long options `longOptions`.
func showPromptOptions(
	p *Context,
	text string,
	hintText string,
	shortOptions string,
	longOptions ...string) (answer string, err error) {

	options := strings.Split(shortOptions, "/")
	validator := CreateValidatorAnswerOptions(options)

	defaultAnswer, defaultAnswerIdx := getDefaultAnswer(options)

	if p.tool != nil {
		args := append([]string{text, hintText, shortOptions}, longOptions...)
		ans, e := cm.GetOutputFromExecutableTrimmed(p.execCtx, p.tool, cm.UseOnlyStdin(os.Stdin), args...)
		ans = strings.ToLower(ans)

		if e == nil {
			// Validate the answer if possible.
			if validator == nil {
				return ans, nil

			}

			e = validator(ans)
			if e == nil {
				return ans, nil
			}

			return defaultAnswer,
				cm.CombineErrors(e, cm.ErrorF("Answer validation error."))
		}

		err = cm.CombineErrors(e, cm.ErrorF("Could not execute dialog script '%q'", p.tool))
		// else: Runnning fallback ...
	}

	if p.useGUI {

		// Use the GUI dialog.
		answer, e := showPromptOptionsGUI(
			p,
			text,
			defaultAnswerIdx,
			options, longOptions,
			validator)

		if e == nil {
			return answer, nil
		}

		err = cm.CombineErrors(err, e)
		p.log.Info(p.promptFmt("%s %s [%s]: ", text, hintText, shortOptions))

	} else {
		// Use the terminal (if possible...)
		emptyCausesDefault := strs.IsNotEmpty(defaultAnswer)
		question := p.promptFmt("%s %s [%s]: ", text, hintText, shortOptions)

		answer, isPromptDisplayed, e :=
			showPromptOptionsTerminal(
				p,
				question,
				defaultAnswer,
				options,
				emptyCausesDefault,
				validator)

		if e == nil {
			return answer, nil
		}
		err = cm.CombineErrors(err, e)

		if !isPromptDisplayed {
			// Show the prompt in the log output
			p.log.Info(question)
		}
	}

	return defaultAnswer, err
}

func showPromptLoopTerminal(
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
	writeOut := func(s string) error {
		_, e := p.termOut.Write([]byte(s))
		return e // nolint: nlreturn
	}

	for nPrompts < maxPrompts {

		err = writeOut(text)
		nPrompts++

		success := p.termInScanner.Scan()

		if !success {
			err = cm.CombineErrors(err,
				writeOut("\n"),
				cm.ErrorF("Could not read from terminal."))

			break
		}

		ans := p.termInScanner.Text()

		if p.printAnswer {
			_ = writeOut(strs.Fmt(" -> Received: '%s'\n", ans))
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

		e := validator(ans)
		if e == nil {
			return ans, true, nil
		}

		warning := p.errorFmt("Answer validation error: %s", e.Error())
		err = cm.CombineErrors(err, writeOut(warning+"\n"))

		if nPrompts < maxPrompts {
			warning := p.errorFmt("Remaining tries %v.", maxPrompts-nPrompts)
			err = cm.CombineErrors(err, writeOut(warning+"\n"))
		} else if p.panicIfMaxTries {
			p.log.PanicF("Could not validate answer in '%v' tries.", maxPrompts)
		}
	}

	warning := p.errorFmt("Could not get answer, taking default '%s'.", defaultAnswer)
	err = cm.CombineErrors(err, writeOut(warning+"\n"))

	return defaultAnswer, nPrompts != 0, err
}

func showPromptOptionsTerminal(
	p *Context,
	question string,
	defaultAnswer string,
	options []string,
	emptyCausesDefault bool,
	validator AnswerValidator) (string, bool, error) {

	return showPromptLoopTerminal(
		p,
		question,
		defaultAnswer,
		emptyCausesDefault,
		validator)
}

// showPrompt shows a prompt to the user with `text`.
func showPrompt(
	p *Context,
	text string,
	defaultAnswer string,
	validator func(string) error) (answer string, err error) {

	cm.PanicIf(p.tool != nil, "Not yet implemented.")

	if p.useGUI {
		ans, e := showPromptGUI(p, text, defaultAnswer, validator)
		if err == nil {
			return ans, nil
		}

		err = cm.CombineErrors(err, e)

	} else {
		if strs.IsNotEmpty(defaultAnswer) {
			text = p.promptFmt("%s [%s]: ", text, defaultAnswer)
		} else {
			text = p.promptFmt("%s : ", text)
		}

		answer, isPromptDisplayed, e :=
			showPromptTerminal(p, text, defaultAnswer, validator)

		if e == nil {
			return answer, nil
		}
		err = cm.CombineErrors(err, e)

		if !isPromptDisplayed {
			// Show the prompt in the log output
			p.log.Info(text)
		}
	}

	return defaultAnswer, err
}

func showPromptTerminal(
	p *Context,
	text string,
	defaultAnswer string,
	validator AnswerValidator) (string, bool, error) {
	return showPromptLoopTerminal(p, text, defaultAnswer, true, validator)
}

func showPromptMulti(
	p *Context,
	text string,
	exitAnswer string,
	validator AnswerValidator) (answers []string, err error) {

	cm.PanicIf(p.tool != nil, "Not yet implemented.")

	// Wrap the validator into another one which
	// reacts on `exitAnswer`.
	exitReceived := false
	var val AnswerValidator = func(s string) error {
		if s == exitAnswer {
			exitReceived = true

			return nil
		}

		return validator(s)
	}

	for {
		ans, err := showPrompt(p, text, "", val)

		if exitReceived {
			break
		} else if err != nil {
			continue
		}

		// Add the entry.
		answers = append(answers, ans)
	}

	return
}
