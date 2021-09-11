//go:build windows

package gui

import (
	"context"

	gwin "github.com/gabyx/githooks/githooks/apps/dialog/gui/windows"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

func ShowMessage(ctx context.Context, msg *set.Message) (res.Message, error) {
	return gwin.ShowMessage(ctx, msg)
}
