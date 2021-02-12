package common

import (
	"errors"
	strs "gabyx/githooks/strings"
)

var ErrCancled = errors.New("Cancled")

type ErrExtraButton struct {
	ButtonIndex uint
}

func (e *ErrExtraButton) Error() string {
	return strs.Fmt("Extra Button '%v'", e.ButtonIndex)
}
