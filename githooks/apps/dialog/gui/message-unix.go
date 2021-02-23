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

func ShowMessage(ctx context.Context, msg *set.Message) (r res.Message, err error) {

	msg.SetDefaultIcons()

	var args []string

	switch msg.Style {
	case set.QuestionStyle:
		args = append(args, "--question")
	case set.InfoStyle:
		args = append(args, "--info")
	case set.WarningStyle:
		args = append(args, "--warning")
	case set.ErrorStyle:
		args = append(args, "--error")
	}

	// Zenity prints default title and text if not set.
	args = append(args, "--title", msg.Title)

	if msg.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", msg.Width))
	}

	if msg.Height > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", msg.Height))
	}

	switch msg.WindowIcon {
	case set.ErrorIcon:
		args = append(args, "--window-icon=error")
	case set.WarningIcon:
		args = append(args, "--window-icon=warning")
	case set.InfoIcon:
		args = append(args, "--window-icon=info")
	case set.QuestionIcon:
		args = append(args, "--window-icon=question")
	}

	// Zenity prints default title and text if not set.
	args = append(args, "--text", msg.Text, "--no-markup")

	if strs.IsNotEmpty(msg.OkLabel) {
		args = append(args, "--ok-label", msg.OkLabel)
	}

	if msg.Style == set.QuestionStyle && strs.IsNotEmpty(msg.CancelLabel) {
		args = append(args, "--cancel-label", msg.CancelLabel)
	}

	if msg.ExtraButtons != nil {
		var extraButtons []string
		extraButtons, err = addInvisiblePrefix(msg.ExtraButtons)
		if err != nil {
			return
		}

		for i := range extraButtons {
			args = append(args, "--extra-button", extraButtons[i])
		}
	}

	if msg.NoWrap {
		args = append(args, "--no-wrap")
	}

	if msg.Ellipsize {
		args = append(args, "--ellipsize")
	}

	if msg.DefaultCancel {
		args = append(args, "--default-cancel")
	}

	switch msg.Icon {
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
				return res.Message{General: getResultButtons(string(out), len(msg.ExtraButtons)+1)}, nil
			}

			return res.Message{General: res.CancelResult()}, nil
		}
	}

	return res.Message{}, err
}
