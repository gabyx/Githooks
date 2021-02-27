// +build darwin

package gui

import (
	"context"
	"os/exec"
	"strings"

	gmac "gabyx/githooks/apps/dialog/gui/darwin"
	res "gabyx/githooks/apps/dialog/result"
	sets "gabyx/githooks/apps/dialog/settings"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
)

// NewMessageFromEntry create a new message
// setting based on a entry setting.
func NewMessageFromEntry(e *sets.Entry) sets.Message {
	return sets.Message{
		General:       e.General,
		GeneralText:   e.GeneralText,
		DefaultButton: e.DefaultButton,
		Icon:          e.Icon,
		Style:         sets.InfoStyle,
	}
}

func translateEntry(entry *sets.Entry) (d gmac.EntryData, err error) {

	m := NewMessageFromEntry(entry)

	if strs.IsEmpty(m.CancelLabel) {
		m.CancelLabel = "Cancel"
	}
	if strs.IsEmpty(m.OkLabel) {
		m.OkLabel = "OK"
	}

	if entry.ExtraButtons != nil {
		err = cm.ErrorF("Extra buttons are not supported on macOS")

		return
	}

	m.Style = sets.QuestionStyle
	m.Icon = sets.InfoIcon

	md, err := translateMessage(&m)
	if err != nil {
		return
	}

	d = gmac.NewFromEntry(&md)
	d.Opts.HiddenAnswer = entry.HideDefaultEntry
	d.Opts.DefaultAnswer = entry.DefaultEntry

	return
}

func ShowEntry(ctx context.Context, s *sets.Entry) (res.Entry, error) {

	data, err := translateEntry(s)
	if err != nil {
		return res.Entry{}, err
	}

	out, err := gmac.RunOSAScript(ctx, "entry", data, "")
	if err == nil {

		// Any linebreak at the end will be trimmed away.
		s := strings.TrimSuffix(string(out), "\n")

		return res.Entry{
			General: res.OkResult(),
			Text:    s}, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 1 {
			return res.Entry{General: res.CancelResult()}, nil
		}
	}

	return res.Entry{}, err
}
