//go:build !windows && !darwin

package gui

import (
	"context"

	gunix "github.com/gabyx/githooks/githooks/apps/dialog/gui/unix"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
)

// ShowFileSave shows a file-save dialog.
func ShowFileSave(ctx context.Context, s *set.FileSave) (r res.File, err error) {
	zenity, err := gunix.GetZenityExecutable()
	if err != nil {
		return
	}
	return ShowFileSaveZenity(ctx, zenity, s)
}

// ShowFileSelection shows a file-selection dialog.
func ShowFileSelection(ctx context.Context, s *set.FileSelection) (r res.File, err error) {
	zenity, err := gunix.GetZenityExecutable()
	if err != nil {
		return
	}
	return ShowFileSelectionZenity(ctx, zenity, s)
}
