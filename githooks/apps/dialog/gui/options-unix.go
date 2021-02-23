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

func ShowOptions(ctx context.Context, s *set.Options) (res.Options, error) {

	if len(opts.Options) == 0 {
		err = cm.ErrorF("You need at list one option specified.")

		return
	}

	if s.Style == set.OptionsStyleButtons && !s.MultipleSelection {
		return showOptionsWithButtons(ctx, s)
	}

	args := []string{
		"--list",
		"--hide-header",
		"--column=id",
		"--column=",
		"--hide-column=1",
		"--print-column=1"}

	// Zenity prints default title and text if not set.
	args = append(args, "--title", s.Title)
	args = append(args, "--text", s.Text, "--no-markup")

	if s.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", s.Width))
	}

	if s.Height > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", s.Height))
	}

	switch s.WindowIcon {
	case set.ErrorIcon:
		args = append(args, "--window-icon=error")
	case set.WarningIcon:
		args = append(args, "--window-icon=warning")
	case set.InfoIcon:
		args = append(args, "--window-icon=info")
	case set.QuestionIcon:
		args = append(args, "--window-icon=question")
	}

	if strs.IsNotEmpty(s.OkLabel) {
		args = append(args, "--ok-label", s.OkLabel)
	}

	if strs.IsNotEmpty(s.CancelLabel) {
		args = append(args, "--cancel-label", s.CancelLabel)
	}

	if s.ExtraButtons != nil {
		for i := range s.ExtraButtons {

			if strs.IsEmpty(s.ExtraButtons[i]) {
				return res.Options{}, cm.ErrorF("Empty label for extra button is not allowed")
			}

			args = append(args, "--extra-button", s.ExtraButtons[i])
		}
	}

	if s.NoWrap {
		args = append(args, "--no-wrap")
	}

	if s.Ellipsize {
		args = append(args, "--ellipsize")
	}

	if s.DefaultCancel {
		args = append(args, "--default-cancel")
	}

	// List options
	if s.MultipleSelection {
		args = append(args, "--multiple")
		args = append(args, "--separator", "\x1e")
	}

	// Add choices with ids.
	for i := range s.Options {
		args = append(args, fmt.Sprintf("%d", i), s.Options[i])
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
				for i := range s.ExtraButtons {
					if button == s.ExtraButtons[i] {
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
