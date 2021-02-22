// +build windows

package gui

import (
	"context"

	gwin "gabyx/githooks/apps/dialog/gui/windows"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
)

func ShowEntry(ctx context.Context, s *set.Entry) (res.Entry, error) {

	return gwin.ShowEntry(ctx, s)
}
