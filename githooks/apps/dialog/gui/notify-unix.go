//go:build !windows && !darwin

package gui

import (
	"context"

	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

// ShowNotification shows a system notifaction.
func ShowNotification(ctx context.Context, n *set.Notification) error {
	zenity, err := gunix.GetZenityExecutable()
	if err != nil {
		return err
	}
	return ShowNotificationZenity(ctx, zenity, n)
}
