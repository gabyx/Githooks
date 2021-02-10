package prompt

import (
	cm "gabyx/githooks/common"
	"strings"

	"github.com/gen2brain/dlgs"
)

func showPromptDialog(
	text string,
	defaultAnswer string) (string, error) {
	ans, success, err := dlgs.Entry("Githooks needs your attention:", text, defaultAnswer)

	if err != nil || !success {
		return "", cm.CombineErrors(err, cm.Error("GUI prompt dialog failed."))
	}

	return ans, nil
}

func showPromptOptionDialog(
	question string,
	defaultAnswerIdx int,
	options []string,
	longOptions []string) (string, error) {

	if len(options) == 2 { // nolint: gomnd

		o1 := strings.ToLower(options[0])
		o2 := strings.ToLower(options[1])

		if (o1 == "y" && o2 == "n") ||
			(o1 == "n" && o2 == "y") {
			// This is a normal yes/no prompt

			defaultCancel := (defaultAnswerIdx == 0 && o1 == "n") ||
				(defaultAnswerIdx == 1 && o2 == "n")

			yes, err := dlgs.Question("Githooks needs your attention:", question, defaultCancel)
			if err != nil {
				return "",
					cm.CombineErrors(err, cm.Error("GUI option dialog failed."))
			}

			if yes {
				return "y", nil
			}

			return "n", nil
		}
	}

	// Make a list dialog
	selected, success, err := dlgs.List("Githooks needs your attention:", question, longOptions)
	if err != nil || !success {
		return "", cm.CombineErrors(err, cm.Error("GUI option dialog failed."))
	}

	// Get the chosen answer
	for i := range longOptions {
		if selected == longOptions[i] {
			return strings.ToLower(options[i]), nil
		}
	}

	// User has not chosen anything or canceled...
	return "", nil
}

func showPromptLoop(
	p *Context,
	defaultAnswer string,
	runPrompt func() (string, error),
	validator AnswerValidator) (string, error) {

	var err error // all errors

	nPrompts := uint(0) // How many times we showed the prompt
	maxPrompts := p.maxTries

	for nPrompts < maxPrompts {

		ans, e := runPrompt()
		nPrompts++

		if err != nil {
			err = cm.CombineErrors(err, e)

			break
		}

		if p.printAnswer {
			p.log.InfoF(" -> Received: '%s'", ans)
		}

		// Validate the answer if possible.
		if validator == nil {
			return ans, nil
		}

		e = validator(ans)
		if e == nil {
			return ans, nil
		}

		p.log.WarnF("Answer validation error: %s", e.Error())

		if nPrompts < maxPrompts {
			p.log.WarnF("Remaining tries %v.", maxPrompts-nPrompts)
		} else if p.panicIfMaxTries {
			p.log.PanicF("Could not validate answer in '%v' tries.", maxPrompts)
		}
	}

	p.log.WarnF("Could not get answer, taking default '%s'.", defaultAnswer)

	return defaultAnswer, err
}

func showPromptOptionsGUI(
	p *Context,
	question string,
	defaultAnswerIdx int,
	options []string,
	longOptions []string,
	validator AnswerValidator) (string, error) {

	defaultAnswer := ""
	if defaultAnswerIdx >= 0 {
		defaultAnswer = options[defaultAnswerIdx]
	}

	return showPromptLoop(
		p,
		defaultAnswer,
		func() (string, error) {
			return showPromptOptionDialog(question, defaultAnswerIdx, options, longOptions)
		},
		validator)
}

func showPromptGUI(
	p *Context,
	text string,
	defaultAnswer string,
	validator AnswerValidator) (string, error) {

	return showPromptLoop(
		p,
		defaultAnswer,
		func() (string, error) {
			return showPromptDialog(text, defaultAnswer)
		},
		validator)
}
