//go:build windows

package gui

import (
	"context"

	gwin "github.com/gabyx/githooks/githooks/apps/dialog/gui/windows"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	set "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	sets "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	cm "github.com/gabyx/githooks/githooks/common"
)

func ShowOptions(ctx context.Context, opts *set.Options) (r res.Options, err error) {
	if len(opts.Options) == 0 {
		err = cm.ErrorF("You need at least one option specified.")

		return
	}

	if opts.Style == sets.OptionsStyleButtons && !opts.MultipleSelection {
		return showOptionsWithButtons(ctx, opts, nil)
	}

	return gwin.ShowOptions(ctx, opts)
}
