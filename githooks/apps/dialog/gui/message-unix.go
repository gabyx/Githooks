//+ build linux

package gui

import (
	"context"
	"fmt"
	"os/exec"

	dcm "gabyx/githooks/apps/dialog/common"
	gunix "gabyx/githooks/apps/dialog/gui/unix"
	strs "gabyx/githooks/strings"
)

func ShowMessage(ctx context.Context, s *MessageSettings) (bool, error) {

	var args []string

	switch s.Style {
	case QuestionStyle:
		args = append(args, "--question")
	case InfoStyle:
		args = append(args, "--info")
	case WarningStyle:
		args = append(args, "--warning")
	case ErrorStyle:
		args = append(args, "--error")
	}

	if strs.IsNotEmpty(s.Title) {
		args = append(args, "--title", s.Title)
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

	if strs.IsNotEmpty(s.Text) {
		args = append(args, "--text", s.Text, "--no-markup")
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

	switch s.Icon {
	case ErrorIcon:
		args = append(args, "--icon-name=dialog-error")
	case WarningIcon:
		args = append(args, "--icon-name=dialog-warning")
	case InfoIcon:
		args = append(args, "--icon-name=dialog-information")
	case QuestionIcon:
		args = append(args, "--icon-name=dialog-question")
	}

	out, err := gunix.RunZenity(ctx, args)
	if err == nil {
		return true, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 1 {

			// Handle extra buttons.
			if len(out) > 0 {
				button := string(out[:len(out)-1])
				for i := range s.ExtraButtons {
					if button == s.ExtraButtons[i] {
						return false, &dcm.ErrExtraButton{ButtonIndex: uint(i)}
					}
				}
			}

			return false, dcm.ErrCancled
		}
	}

	return false, err
}
