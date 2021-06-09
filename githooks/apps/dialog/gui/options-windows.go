// +build windows

package gui

import (
	"context"

	gwin "gabyx/githooks/apps/dialog/gui/windows"
	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
	sets "gabyx/githooks/apps/dialog/settings"
	cm "gabyx/githooks/common"
)

func ShowOptions(ctx context.Context, opts *set.Options) (r res.Options, err error) {
	if len(opts.Options) == 0 {
		err = cm.ErrorF("You need at least one option specified.")

		return
	}

	if opts.Style == sets.OptionsStyleButtons && !opts.MultipleSelection {
		return showOptionsWithButtons(ctx, opts)
	}

	return gwin.ShowOptions(ctx, opts)
}
