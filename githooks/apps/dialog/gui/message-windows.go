// +build windows

package gui

import (
	"context"

	gwin "gabyx/githooks/apps/dialog/gui/windows"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
)

func ShowMessage(ctx context.Context, msg *set.Message) (res.Message, error) {
	return gwin.ShowMessage(ctx, msg)
}
