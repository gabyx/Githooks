//go:build windows

package gui

import (
	"context"

	gwin "github.com/gabyx/githooks/githooks/apps/dialog/gui/windows"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

func ShowEntry(ctx context.Context, s *set.Entry) (res.Entry, error) {

	return gwin.ShowEntry(ctx, s)
}
