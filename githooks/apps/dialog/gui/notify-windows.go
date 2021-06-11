// +build windows

package gui

import (
	"context"

	gwin "github.com/gabyx/githooks/githooks/apps/dialog/gui/windows"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

func ShowNotification(ctx context.Context, s *set.Notification) error {
	return gwin.ShowNotification(ctx, s)
}
