//go:build !windows

package gui

import (
	"context"
	"strconv"

	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// ShowNotification shows a system notification with `zenity`.
func ShowNotificationZenity(ctx context.Context, zenity string, n *set.Notification) error {
	args := []string{"--notification"}

	if strs.IsNotEmpty(n.Title) {
		args = append(args, "--title", n.Title)
	}

	if n.Width > 0 {
		args = append(args, "--width", strconv.FormatUint(uint64(n.Width), 10))
	}

	if n.Height > 0 {
		args = append(args, "--height", strconv.FormatUint(uint64(n.Height), 10))
	}

	switch n.WindowIcon {
	case set.ErrorIcon:
		args = append(args, "--window-icon=error")
	case set.WarningIcon:
		args = append(args, "--window-icon=warning")
	case set.InfoIcon:
		args = append(args, "--window-icon=info")
	case set.QuestionIcon:
		args = append(args, "--window-icon=question")
	}

	if strs.IsNotEmpty(n.Text) {
		args = append(args, "--text", n.Text, "--no-markup")
	}

	_, err := gunix.RunZenity(ctx, zenity, args, "")

	return err
}
