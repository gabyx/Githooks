// +build !windows,!darwin

package gui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	gunix "gabyx/githooks/apps/dialog/gui/unix"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
)

func ShowEntry(ctx context.Context, s *set.Entry) (res.Entry, error) {

	args := []string{"--entry"}

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

	if strs.IsNotEmpty(s.EntryText) {
		args = append(args, "--entry-text", s.EntryText)
	}

	if s.HideEntryText {
		args = append(args, "--hide-text")
	}

	out, err := gunix.RunZenity(ctx, args, "")
	if err == nil {

		// Any linebreak at the end will be trimmed away.
		s := strings.TrimSuffix(string(out), "\n")

		return res.Entry{
			General: res.OkResult(),
			Text:    s}, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 1 {

			// Handle extra buttons.
			if len(out) > 0 {
				button := string(out[:len(out)-1])
				for i := range s.ExtraButtons {
					if button == s.ExtraButtons[i] {
						return res.Entry{
							General: res.ExtraButtonResult(uint(i))}, nil
					}
				}
			}

			return res.Entry{General: res.CancelResult()}, nil
		}
	}

	return res.Entry{}, err
}
