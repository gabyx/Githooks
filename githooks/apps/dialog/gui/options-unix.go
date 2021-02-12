//+ build linux

package gui

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	dcm "gabyx/githooks/apps/dialog/common"
	gunix "gabyx/githooks/apps/dialog/gui/unix"
	strs "gabyx/githooks/strings"
)

func getChoices(output string) (indices []uint) {
	out := strings.Split(strings.TrimSpace(output), "|")

	indices = make([]uint, 0, len(out))

	for i := range out {
		idx, err := strconv.ParseUint(out[i], 10, 32)
		if err == nil {
			indices = append(indices, uint(idx))
		}
	}

	return
}

func ShowOptions(ctx context.Context, s *OptionsSettings) ([]uint, error) {

	args := []string{
		"--list",
		"--hide-header",
		"--column=id",
		"--column=",
		"--hide-column=1",
		"--print-column=1"}

	if strs.IsNotEmpty(s.Title) {
		args = append(args, "--title", s.Title)
	}

	if strs.IsNotEmpty(s.Text) {
		args = append(args, "--text", s.Text, "--no-markup")
	}

	if s.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", s.Width))
	}
	if s.Height > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", s.Height))
	}

	switch s.WindowIcon {
	case ErrorIcon:
		args = append(args, "--window-icon=error")
	case WarningIcon:
		args = append(args, "--window-icon=warning")
	case InfoIcon:
		args = append(args, "--window-icon=info")
	case QuestionIcon:
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
		args = append(args, "--separator=|")
	}

	// Add choices with ids.
	for i := range s.Options {
		args = append(args, fmt.Sprintf("%d", i), s.Options[i])
	}

	out, err := gunix.RunZenity(ctx, args)
	if err == nil {
		return getChoices(string(out)), nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 1 {

			// Handle extra buttons.
			if len(out) > 0 {
				button := string(out[:len(out)-1])
				for i := range s.ExtraButtons {
					if button == s.ExtraButtons[i] {
						return nil, &dcm.ErrExtraButton{ButtonIndex: uint(i)}
					}
				}
			}

			return nil, dcm.ErrCancled
		}
	}

	return nil, err
}
