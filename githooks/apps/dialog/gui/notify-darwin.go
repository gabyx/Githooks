//go:build darwin

package gui

import (
	"context"
	"strings"

	gmac "github.com/gabyx/githooks/githooks/apps/dialog/gui/darwin"
	sets "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	strs "github.com/gabyx/githooks/githooks/strings"
)

func ShowNotification(ctx context.Context, s *sets.Notification) error {

	data := gmac.NotifyData{}

	if strs.IsEmpty(s.Title) {
		data.Opts.WithTitle = "Notification" // Always prints "ScriptEditor" otherwise...
	} else {
		data.Opts.WithTitle = s.Title
	}

	if i := strings.IndexByte(s.Text, '\n'); i >= 0 && i < len(s.Text) {
		data.Opts.Subtitle = s.Text[:i]
		data.Text = s.Text[i+1:]
	} else {
		data.Text = s.Text
	}

	_, err := gmac.RunOSAScript(ctx, "notify", data, "")

	return err
}
