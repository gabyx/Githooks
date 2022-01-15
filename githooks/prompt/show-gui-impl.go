package prompt

import (
	"errors"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/gabyx/githooks/githooks/apps/dialog/gui"
	"github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

func showEntryDialog(
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
	opts.ForceTopMost = true // only for Windows this is crucial, such that it does not get ignored.

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
	defaultAnswer string,
	defaultOptionIdx int,
	options []string,
	longOptions []string) (string, error) {

	opts := settings.Options{}
	opts.Title = title
	opts.Text = question
	opts.Options = longOptions
	opts.Style = settings.OptionsStyleButtons
	opts.WindowIcon = settings.QuestionIcon
	opts.Width = 400
	opts.ForceTopMost = true // only for Windows this is crucial, such that it does not get ignored.

	if defaultOptionIdx >= 0 {
		opts.DefaultOptions = []uint{uint(defaultOptionIdx)}
	}

	res, err := gui.ShowOptions(nil, &opts) // nolint

	switch {
	case err != nil || res.IsCanceled():
		return defaultAnswer, err
	case res.IsOk():
		if res.Options == nil {
			// nothing chosen
			return defaultAnswer, err
		} else {
			return options[res.Options[0]], err
		}
	}

	cm.Panic("Wrong dialog result state")

	return defaultAnswer, err
}

func showPromptLoop(
	p *Context,
	defaultAnswer string,
	runPrompt func() (string, error),
	validator AnswerValidator,
	canCancel bool) (string, error) {

	var err error
	var ans string

	nPrompts := uint(0) // How many times we showed the prompt
	maxPrompts := p.maxTries

	for nPrompts < maxPrompts {

		ans, err = runPrompt()
		nPrompts++

		if errors.Is(err, ErrorCanceled) {
			if !canCancel {
				p.log.WarnF("User canceled. Remaining tries %v.", maxPrompts-nPrompts)

				continue
			} else {
				return defaultAnswer, err
			}

		} else if err != nil {
			p.log.WarnF("Prompt failed: %s", err.Error())

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

func showMessageGUI(
	p *Context,
	title string,
	message string,
	asError bool) error {

	opts := settings.Message{}
	opts.Title = title
	opts.Text = message
	opts.ForceTopMost = true // only for Windows this is crucial, such that it does not get ignored.

	if asError {
		opts.WindowIcon = settings.InfoIcon
		opts.Icon = settings.InfoIcon
	} else {
		opts.WindowIcon = settings.ErrorIcon
		opts.Icon = settings.ErrorIcon
	}
	opts.Width = 400

	_, err := gui.ShowMessage(nil, &opts) // nolint

	return err
}

func showOptionsGUI(
	p *Context,
	title string,
	question string,
	defaultAnswer string,
	defaultOptionIdx int,
	options []string,
	longOptions []string,
	validator AnswerValidator) (string, error) {

	return showPromptLoop(
		p,
		defaultAnswer,
		func() (string, error) {
			return showOptionsDialog(title, question, defaultAnswer, defaultOptionIdx, options, longOptions)
		},
		validator,
		false)
}

func showEntryGUI(
	p *Context,
	title string,
	text string,
	defaultAnswer string,
	validator AnswerValidator,
	canCancel bool) (string, error) {

	return showPromptLoop(
		p,
		defaultAnswer,
		func() (string, error) {
			return showEntryDialog(title, text, defaultAnswer)
		},
		validator,
		canCancel)
}
