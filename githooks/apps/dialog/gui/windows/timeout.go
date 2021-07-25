// +build windows

package gui

import (
	"context"

	"github.com/lxn/walk"
)

func watchTimeout(ctx context.Context, dlg *walk.Dialog) {
	wait := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			dlg.Close(walk.DlgCmdTimeout)
			close(wait)

			return
		case <-wait:
		}
	}()
}
