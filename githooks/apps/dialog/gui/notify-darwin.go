//go:build darwin

package gui

import (
	"context"
	"strings"

	gmac "github.com/gabyx/githooks/githooks/apps/dialog/gui/darwin"
	sets "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	strs "github.com/gabyx/githooks/githooks/strings"
)

func ShowNotification(ctx context.Context, n *sets.Notification) error {

	data := gmac.NotifyData{}

	if strs.IsEmpty(n.Title) {
		data.Opts.WithTitle = "Notification" // Always prints "ScriptEditor" otherwise...
	} else {
		data.Opts.WithTitle = n.Title
	}

	if i := strings.IndexByte(n.Text, '\n'); i >= 0 && i < len(n.Text) {
		data.Opts.Subtitle = n.Text[:i]
		data.Text = n.Text[i+1:]
	} else {
		data.Text = n.Text
	}

	_, err := gmac.RunOSAScript(ctx, "notify", data, "")

	return err
}
