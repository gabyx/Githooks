package prompt

import (
	"errors"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"

	"gabyx/githooks/apps/dialog/gui"
	"gabyx/githooks/apps/dialog/settings"
)

func showPromptDialog(
	title string,
	text string,
	defaultAnswer string) (string, error) {

	opts := settings.Entry{}
	opts.Title = title
	opts.Text = text
	opts.DefaultEntry = defaultAnswer
	opts.WindowIcon = settings.InfoIcon
	opts.Icon = settings.InfoIcon
	opts.Width = 400

	res, err := gui.ShowEntry(nil, &opts) // nolint

	switch {
	case err != nil:
		return defaultAnswer, err
	case res.IsCanceled():
		return defaultAnswer, ErrorCanceled
	case res.IsOk():
		return res.Text, err
	}

	cm.Panic("Wrong dialog result state")

	return defaultAnswer, err
}

func showOptionsDialog(
	title string,
	question string,
	defaultOptionIdx int,
	options []string,
	longOptions []string) (string, error) {

	opts := settings.Options{}
	opts.Title = title
	opts.Text = question
	opts.Options = longOptions
	opts.DefaultOptions = []uint{uint(defaultOptionIdx)}
	opts.Style = settings.OptionsStyleButtons
	opts.WindowIcon = settings.QuestionIcon
	opts.Width = 400

	res, err := gui.ShowOptions(nil, &opts) // nolint

	switch {
	case err != nil || res.IsCanceled():
		return options[defaultOptionIdx], err
	case res.IsOk():
		return options[res.Selection[0]], err
	}

	cm.Panic("Wrong dialog result state")

	return options[defaultOptionIdx], err
}

func showPromptLoop(
	p *Context,
	defaultAnswer string,
	runPrompt func() (string, error),
	validator AnswerValidator,
	cancelResultsInRetry bool) (string, error) {

	var err error
	var ans string

	nPrompts := uint(0) // How many times we showed the prompt
	maxPrompts := p.maxTries

	for nPrompts < maxPrompts {

		ans, err = runPrompt()
		nPrompts++

		if errors.Is(err, ErrorCanceled) {
			if cancelResultsInRetry {
				p.log.WarnF("User canceled. Remaining tries %v.", maxPrompts-nPrompts)

				continue
			} else {
				break
			}
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
		validator,
		true)
}

func showEntryGUI(
	p *Context,
	title string,
	text string,
	defaultAnswer string,
	validator AnswerValidator,
	cancelResultsInRetry bool) (string, error) {

	return showPromptLoop(
		p,
		defaultAnswer,
		func() (string, error) {
			return showPromptDialog(title, text, defaultAnswer)
		},
		validator,
		cancelResultsInRetry)
}
