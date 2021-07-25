// +build !windows,!darwin

package gui

import (
	"context"
	"fmt"

	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// ShowNotification shows a system notifaction.
func ShowNotification(ctx context.Context, s *set.Notification) error {

	args := []string{"--notification"}

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

	_, err := gunix.RunZenity(ctx, args, "")

	return err
}
