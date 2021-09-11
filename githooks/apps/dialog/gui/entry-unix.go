//go:build !windows && !darwin

package gui

import (
	"context"

	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

// ShowEntry shows an entry dialog.
func ShowEntry(ctx context.Context, e *set.Entry) (r res.Entry, err error) {
	zenity, err := gunix.GetZenityExecutable()
	if err != nil {
		return
	}
	return ShowEntryZenity(ctx, zenity, e)
}
