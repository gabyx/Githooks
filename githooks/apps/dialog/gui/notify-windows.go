// +build windows

package gui

import (
	"context"

	gwin "gabyx/githooks/apps/dialog/gui/windows"
	set "gabyx/githooks/apps/dialog/settings"
)

func ShowNotification(ctx context.Context, s *set.Notification) error {
	return gwin.ShowNotification(ctx, s)
}
