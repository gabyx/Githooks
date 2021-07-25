// +build !windows,!darwin

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

// ShowEntry shows an entry dialog.
func ShowEntry(ctx context.Context, entry *set.Entry) (r res.Entry, err error) {

	args := []string{"--entry"}

	// Zenity prints default title and text if not set.
	args = append(args, "--title", entry.Title)

	if entry.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", entry.Width))
	}

	if entry.Height > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", entry.Height))
	}

	switch entry.WindowIcon {
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
	args = append(args, "--text", entry.Text, "--no-markup")

	if strs.IsNotEmpty(entry.OkLabel) {
		args = append(args, "--ok-label", entry.OkLabel)
	}

	if strs.IsNotEmpty(entry.CancelLabel) {
		args = append(args, "--cancel-label", entry.CancelLabel)
	}

	if entry.ExtraButtons != nil {
		var extraButtons []string
		extraButtons, err = addInvisiblePrefix(entry.ExtraButtons)
		if err != nil {
			return
		}

		for i := range extraButtons {
			args = append(args, "--extra-button", extraButtons[i])
		}
	}

	if entry.NoWrap {
		args = append(args, "--no-wrap")
	}

	if entry.Ellipsize {
		args = append(args, "--ellipsize")
	}

	if entry.DefaultCancel {
		args = append(args, "--default-cancel")
	}

	if strs.IsNotEmpty(entry.DefaultEntry) {
		args = append(args, "--entry-text", entry.DefaultEntry)
	}

	if entry.HideDefaultEntry {
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
				return res.Entry{General: getResultButtons(string(out), len(entry.ExtraButtons)+1)}, nil
			}

			return res.Entry{General: res.CancelResult()}, nil
		}
	}

	return res.Entry{}, err
}
