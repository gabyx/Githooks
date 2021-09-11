//go:build windows

package gui

import (
	"context"
	"path/filepath"

	gwin "github.com/gabyx/githooks/githooks/apps/dialog/gui/windows"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	cm "github.com/gabyx/githooks/githooks/common"
)

func ShowFileSave(ctx context.Context, s *set.FileSave) (r res.File, err error) {

	s.Root, err = filepath.Abs(filepath.FromSlash(s.Root))
	cm.AssertNoErrorPanic(err, "Could not get absolute path from '%s'", s.Root)

	r, err = gwin.ShowFileSave(ctx, s)

	for i := range r.Paths {
		r.Paths[i] = filepath.ToSlash(r.Paths[i])
	}

	return
}

func ShowFileSelection(ctx context.Context, s *set.FileSelection) (r res.File, err error) {

	s.Root, err = filepath.Abs(filepath.FromSlash(s.Root))
	cm.AssertNoErrorPanic(err, "Could not get absolute path from '%s'", s.Root)

	r, err = gwin.ShowFileSelection(ctx, s)

	for i := range r.Paths {
		r.Paths[i] = filepath.ToSlash(r.Paths[i])
	}

	return
}
