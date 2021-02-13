// +build !windows,!darwin

package gui

import (
	"context"
	"fmt"
	"os/exec"

	gunix "gabyx/githooks/apps/dialog/gui/unix"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
	strs "gabyx/githooks/strings"
)

func ShowMessage(ctx context.Context, s *set.Message) (res.Message, error) {

	var args []string

	switch s.Style {
	case set.QuestionStyle:
		args = append(args, "--question")
	case set.InfoStyle:
		args = append(args, "--info")
	case set.WarningStyle:
		args = append(args, "--warning")
	case set.ErrorStyle:
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
	case set.ErrorIcon:
		args = append(args, "--window-icon=error")
	case set.WarningIcon:
		args = append(args, "--window-icon=warning")
	case set.InfoIcon:
		args = append(args, "--window-icon=info")
	case set.QuestionIcon:
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
	case set.ErrorIcon:
		args = append(args, "--icon-name=dialog-error")
	case set.WarningIcon:
		args = append(args, "--icon-name=dialog-warning")
	case set.InfoIcon:
		args = append(args, "--icon-name=dialog-information")
	case set.QuestionIcon:
		args = append(args, "--icon-name=dialog-question")
	}

	out, err := gunix.RunZenity(ctx, args, "")
	if err == nil {
		return res.Message{General: res.OkResult()}, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 1 {

			// Handle extra buttons.
			if len(out) > 0 {
				button := string(out[:len(out)-1])
				for i := range s.ExtraButtons {
					if button == s.ExtraButtons[i] {
						return res.Message{
							General: res.ExtraButtonResult(uint(i))}, nil
					}
				}
			}

			return res.Message{General: res.CancelResult()}, nil
		}
	}

	return res.Message{}, err
}
