// +build windows

package gui

import (
	"context"

	gwin "gabyx/githooks/apps/dialog/gui/windows"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
)

func ShowOptions(ctx context.Context, opts *set.Options) (res.Options, error) {
	return gwin.ShowOptions(ctx, opts)
}
