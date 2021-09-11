//go:build !windows && !darwin

package gui

import (
	"context"

	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

// ShowMessage shows a message dialog.
func ShowMessage(ctx context.Context, msg *set.Message) (r res.Message, err error) {
	zenity, err := gunix.GetZenityExecutable()
	if err != nil {
		return
	}

	return ShowMessageZenity(ctx, zenity, msg)
}
