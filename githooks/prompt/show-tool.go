package prompt

import (
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
	"os/exec"
	"strconv"
	"strings"
)

func runDialogToolOptions(
	tool ToolContext,
	title string,
	text string,
	defaultOptionIdx int,
	options []string) (ans string, err error) {

	cm.PanicIfF(!tool.IsSetup(), "Tool context is not setup.")
	cm.PanicIfF(defaultOptionIdx >= len(options), "Wrong index.")

	var opts []string
	for _, o := range options {
		opts = append(opts, "--option", o)
	}

	args := append([]string{
		"options",
		"--title", title,
		"--text", text,
		"--default-option", strs.Fmt("%v", defaultOptionIdx)},
		opts...)

	ans, err = cm.GetOutputFromExecutableTrimmed(tool.execCtx, tool.tool, nil, args...)

	if e, ok := err.(*exec.ExitError); ok {
		if e.ExitCode() == 1 {
			err = CancledError

			return
		}
	}

	ans = strings.ToLower(ans)

	// Get the chosen option idx.
	idx, e := strconv.ParseInt(ans, 10, 32)
	if e != nil || int(idx) >= len(options) {
		err = cm.ErrorF("Dialog tool returned wrong index '%v' (< '%v')",
			ans, len(options))

		return
	}

	return options[idx], nil
}

func runDialogToolEntry(
	tool ToolContext,
	title string,
	text string,
	defaultAnswer string) (ans string, err error) {

	cm.PanicIfF(!tool.IsSetup(), "Tool context is not setup.")

	args := []string{
		"entry",
		"--title", title,
		"--text", text,
		"--default-entry", defaultAnswer}

	ans, err = cm.GetOutputFromExecutableTrimmed(tool.execCtx, tool.tool, nil, args...)

	if e, ok := err.(*exec.ExitError); ok {
		if e.ExitCode() == 1 {
			err = CancledError
		}
	}

	return
}

func showOptionsTool(
	p *Context,
	title string,
	text string,
	defaultOptionIdx int,
	options []string,
	validator AnswerValidator) (string, error) {

	defaultAnswer := ""
	if defaultOptionIdx >= 0 {
		defaultAnswer = options[defaultOptionIdx]
	}

	return showPromptLoop(
		p,
		defaultAnswer,
		func() (string, error) {
			return runDialogToolOptions(p.tool, title, text, defaultOptionIdx, options)
		},
		validator)
}

func showEntryTool(
	p *Context,
	title string,
	text string,
	defaultAnswer string,
	validator AnswerValidator) (string, error) {

	return showPromptLoop(
		p,
		defaultAnswer,
		func() (string, error) {
			return runDialogToolEntry(p.tool, title, text, defaultAnswer)
		},
		validator)
}
