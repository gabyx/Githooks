//go:build windows

package gui

import (
	"context"

	gwin "github.com/gabyx/githooks/githooks/apps/dialog/gui/windows"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

func ShowNotification(ctx context.Context, n *set.Notification) error {
	return gwin.ShowNotification(ctx, n)
}
