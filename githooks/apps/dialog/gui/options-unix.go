// +build !windows,!darwin

package gui

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	gunix "gabyx/githooks/apps/dialog/gui/unix"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
)

func getChoices(output string) (indices []uint) {
	out := strings.Split(strings.TrimSpace(output), "\x1e")

	indices = make([]uint, 0, len(out))

	for i := range out {
		idx, err := strconv.ParseUint(out[i], 10, 32)
		if err == nil {
			indices = append(indices, uint(idx))
		}
	}

	return
}

func ShowOptions(ctx context.Context, opts *set.Options) (r res.Options, err error) {

	if len(opts.Options) == 0 {
		err = cm.ErrorF("You need at list one option specified.")

		return
	}

	if opts.Style == set.OptionsStyleButtons && !opts.MultipleSelection {
		return showOptionsWithButtons(ctx, opts)
	}

	args := []string{
		"--list",
		"--hide-header",
		"--column=id",
		"--column=",
		"--hide-column=1",
		"--print-column=1"}

	// Zenity prints default title and text if not set.
	args = append(args, "--title", opts.Title)
	args = append(args, "--text", opts.Text, "--no-markup")

	if opts.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", opts.Width))
	}

	if opts.Height > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", opts.Height))
	}

	switch opts.WindowIcon {
	case set.ErrorIcon:
		args = append(args, "--window-icon=error")
	case set.WarningIcon:
		args = append(args, "--window-icon=warning")
	case set.InfoIcon:
		args = append(args, "--window-icon=info")
	case set.QuestionIcon:
		args = append(args, "--window-icon=question")
	}

	if strs.IsNotEmpty(opts.OkLabel) {
		args = append(args, "--ok-label", opts.OkLabel)
	}

	if strs.IsNotEmpty(opts.CancelLabel) {
		args = append(args, "--cancel-label", opts.CancelLabel)
	}

	if opts.ExtraButtons != nil {
		for i := range opts.ExtraButtons {

			if strs.IsEmpty(opts.ExtraButtons[i]) {
				return res.Options{}, cm.ErrorF("Empty label for extra button is not allowed")
			}

			args = append(args, "--extra-button", opts.ExtraButtons[i])
		}
	}

	if opts.NoWrap {
		args = append(args, "--no-wrap")
	}

	if opts.Ellipsize {
		args = append(args, "--ellipsize")
	}

	if opts.DefaultCancel {
		args = append(args, "--default-cancel")
	}

	// List options
	if opts.MultipleSelection {
		args = append(args, "--multiple")
		args = append(args, "--separator", "\x1e")
	}

	// Add choices with ids.
	for i := range opts.Options {
		args = append(args, fmt.Sprintf("%d", i), opts.Options[i])
	}

	out, err := gunix.RunZenity(ctx, args, "")
	if err == nil {
		return res.Options{
			General:   res.OkResult(),
			Selection: getChoices(string(out))}, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 1 {

			// Handle extra buttons.
			if len(out) > 0 {
				button := string(out[:len(out)-1])
				for i := range opts.ExtraButtons {
					if button == opts.ExtraButtons[i] {
						return res.Options{
							General: res.ExtraButtonResult(uint(i))}, nil
					}
				}
			}

			return res.Options{General: res.CancelResult()}, nil
		}
	}

	return res.Options{}, err
}
