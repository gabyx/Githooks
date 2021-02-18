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

	"github.com/ulule/deepcopier"
)

func translateEntry(entry *sets.Entry) (d gmac.MsgData, err error) {

	m := sets.Message{}
	err = deepcopier.Copy(entry).To(&m)
	cm.AssertNoErrorPanic(err, "Struct copy failed")

	if strs.IsEmpty(m.CancelLabel) {
		m.CancelLabel = "Cancel"
	}
	if strs.IsEmpty(m.OkLabel) {
		m.OkLabel = "Ok"
	}

	if entry.ExtraButtons != nil {
		err = cm.ErrorF("Extra buttons are not supported on macOS")

		return
	}

	m.Style = sets.QuestionStyle

	d, err = translateMessage(&m)
	if err != nil {
		return
	}

	// Entry fields
	d.Opts.HiddenAnswer = entry.HideEntryText
	d.Opts.DefaultAnswer = entry.EntryText

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
