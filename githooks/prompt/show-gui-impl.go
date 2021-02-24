package prompt

import (
	"errors"
	strs "gabyx/githooks/strings"
)

func showPromptDialog(
	title string,
	text string,
	defaultAnswer string) (string, error) {

	return "", nil
}

func showOptionsDialog(
	title string,
	question string,
	defaultOptionIdx int,
	options []string,
	longOptions []string) (string, error) {
	return "", nil
}

func showPromptLoop(
	p *Context,
	defaultAnswer string,
	runPrompt func() (string, error),
	validator AnswerValidator) (string, error) {

	var err error
	var ans string

	nPrompts := uint(0) // How many times we showed the prompt
	maxPrompts := p.maxTries

	for nPrompts < maxPrompts {

		ans, err = runPrompt()
		nPrompts++

		if errors.Is(err, CancledError) {
			p.log.WarnF("User canceled. Remaining tries %v.", maxPrompts-nPrompts)

			continue
		} else if err != nil {
			break // Any other error is fatal.
		}

		if p.printAnswer {
			p.log.InfoF(" -> Received: '%s'", ans)
		}

		// Validate the answer if possible.
		if validator == nil {
			return ans, nil
		}

		err = validator(ans)
		if err == nil {
			return ans, nil
		}

		p.log.WarnF("Answer validation error: %s", err.Error())

		if nPrompts < maxPrompts {
			p.log.WarnF("Remaining tries %v.", maxPrompts-nPrompts)
		} else {
			msg := strs.Fmt("Could not validate answer in '%v' tries.", maxPrompts)
			if p.panicIfMaxTries {
				p.log.PanicF(msg)
			} else {
				return defaultAnswer, err
			}
		}
	}

	p.log.WarnF("Could not get answer, taking default '%s'.", defaultAnswer)

	return defaultAnswer, err
}

func showOptionsGUI(
	p *Context,
	title string,
	question string,
	defaultOptionIdx int,
	options []string,
	longOptions []string,
	validator AnswerValidator) (string, error) {

	defaultAnswer := ""
	if defaultOptionIdx >= 0 {
		defaultAnswer = options[defaultOptionIdx]
	}

	return showPromptLoop(
		p,
		defaultAnswer,
		func() (string, error) {
			return showOptionsDialog(title, question, defaultOptionIdx, options, longOptions)
		},
		validator)
}

func showEntryGUI(
	p *Context,
	title string,
	text string,
	defaultAnswer string,
	validator AnswerValidator) (string, error) {

	return showPromptLoop(
		p,
		defaultAnswer,
		func() (string, error) {
			return showPromptDialog(title, text, defaultAnswer)
		},
		validator)
}
