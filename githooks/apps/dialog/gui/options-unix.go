//go:build !windows && !darwin

package gui

import (
	"context"

	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

// ShowOptions shows a option dialog.
func ShowOptions(ctx context.Context, opts *set.Options) (r res.Options, err error) {
	zenity, err := gunix.GetZenityExecutable()
	if err != nil {
		return
	}

	return ShowOptionsZenity(ctx, zenity, opts)
}
