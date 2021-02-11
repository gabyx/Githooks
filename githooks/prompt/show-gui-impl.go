package prompt

import (
	"gabyx/githooks/prompt/gui"
	strs "gabyx/githooks/strings"
)

func showPromptLoopGUI(
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

		if err != nil {
			break
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

func showPromptOptionsGUI(
	p *Context,
	title string,
	question string,
	defaultAnswerIdx int,
	options []string,
	longOptions []string,
	validator AnswerValidator) (string, error) {

	defaultAnswer := ""
	if defaultAnswerIdx >= 0 {
		defaultAnswer = options[defaultAnswerIdx]
	}

	return showPromptLoopGUI(
		p,
		defaultAnswer,
		func() (string, error) {
			return gui.ShowOptionsDialog(title, question, defaultAnswerIdx, options, longOptions)
		},
		validator)
}

func showPromptGUI(
	p *Context,
	title string,
	text string,
	defaultAnswer string,
	validator AnswerValidator) (string, error) {

	return showPromptLoopGUI(
		p,
		defaultAnswer,
		func() (string, error) {
			return gui.ShowPromptDialog(title, text, defaultAnswer)
		},
		validator)
}
