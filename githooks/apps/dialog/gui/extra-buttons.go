//go:build !windows

package gui

import (
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"

	"strings"
)

const idPrefix rune = '\u200B'

// addInvisiblePrefix is a workaround:
// Append invisible spaces before each extra button,
// to identify the index afterwards (no string parsing of the label!).
// First button has 2 invisible 'idPrefix', second has 3, etc. to possibly
// include the OK button.
func addInvisiblePrefix(extraButtons []string) (res []string, err error) {
	if extraButtons == nil {
		return nil, nil
	}

	res = make([]string, len(extraButtons))
	id := string(idPrefix)

	for i := range extraButtons {
		if strs.IsEmpty(extraButtons[i]) {
			return nil, cm.ErrorF("Empty label for extra button is not allowed")
		}

		id += string(idPrefix)

		if strs.IsEmpty(extraButtons[i]) {
			err = cm.ErrorF("Empty label for extra button is not allowed")

			return
		}

		res[i] = id + strings.TrimLeft(extraButtons[i], string(idPrefix))
	}

	return
}

// getResultButtons gets the pressed button,
// First button is the Ok button (if not excluded), 2,3,... are extra buttons.
// nolint: mnd
func getResultButtons(out string, maxButtons int) res.General {
	s := strings.TrimSpace(out)
	cm.DebugAssert(strs.IsNotEmpty(s))

	r := []rune(s)

	i := 0
	for ; i < maxButtons && i < len(r); i++ {
		if r[i] != idPrefix {
			break
		}
	}

	cm.DebugAssert(i >= 1)

	// One 'idPrefix' found -> its the ok Button.
	if i == 1 {
		return res.OkResult()
	}

	// otherwise its an extra button
	return res.ExtraButtonResult(uint(i - 2))
}
