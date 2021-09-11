//go:build !windows

package gui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// ShowEntry shows an entry dialog with `zenity`.
func ShowEntryZenity(ctx context.Context, zenity string, e *set.Entry) (r res.Entry, err error) {

	args := []string{"--entry"}

	// Zenity prints default title and text if not set.
	args = append(args, "--title", e.Title)

	if e.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", e.Width))
	}

	if e.Height > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", e.Height))
	}

	switch e.WindowIcon {
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
	args = append(args, "--text", e.Text, "--no-markup")

	if strs.IsNotEmpty(e.OkLabel) {
		args = append(args, "--ok-label", e.OkLabel)
	}

	if strs.IsNotEmpty(e.CancelLabel) {
		args = append(args, "--cancel-label", e.CancelLabel)
	}

	if e.ExtraButtons != nil {
		var extraButtons []string
		extraButtons, err = addInvisiblePrefix(e.ExtraButtons)
		if err != nil {
			return
		}

		for i := range extraButtons {
			args = append(args, "--extra-button", extraButtons[i])
		}
	}

	if e.NoWrap {
		args = append(args, "--no-wrap")
	}

	if e.Ellipsize {
		args = append(args, "--ellipsize")
	}

	if e.DefaultCancel {
		args = append(args, "--default-cancel")
	}

	if strs.IsNotEmpty(e.DefaultEntry) {
		args = append(args, "--entry-text", e.DefaultEntry)
	}

	if e.HideDefaultEntry {
		args = append(args, "--hide-text")
	}

	out, err := gunix.RunZenity(ctx, zenity, args, "")
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
				return res.Entry{General: getResultButtons(string(out), len(e.ExtraButtons)+1)}, nil
			}

			return res.Entry{General: res.CancelResult()}, nil
		}
	}

	return res.Entry{}, err
}
