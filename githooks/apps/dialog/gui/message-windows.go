// +build windows

package gui

import (
	"context"

	res "gabyx/githooks/apps/dialog/result"
	set "gabyx/githooks/apps/dialog/settings"
)

func ShowMessage(ctx context.Context, s *set.Message) (res.Message, error) {

	return res.Message{}, nil
}
