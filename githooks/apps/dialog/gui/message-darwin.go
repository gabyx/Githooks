// +build darwin

package gui

import (
	"context"
	"os/exec"

	gmac "gabyx/githooks/apps/dialog/gui/darwin"
	res "gabyx/githooks/apps/dialog/result"
	sets "gabyx/githooks/apps/dialog/settings"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
)

const idPrefix rune = '\u200B'

func translateMessage(msg *sets.Message) (interface{}, error) {
	d := gmac.MsgData{}

	d.Text = msg.Text
	d.Operation = "displayDialog"
	d.Opts.WithTitle = msg.Title

	switch msg.Style {
	default:
		fallthrough
	case sets.InfoStyle:
		d.WithIcon = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertNoteIcon.icns"
	case sets.ErrorStyle:
		d.WithIcon = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns"
	case sets.WarningStyle:
		d.WithIcon = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertCautionIcon.icns"
	case sets.QuestionStyle:
		d.WithIcon = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/GenericQuestionMarkIcon.icns"
	}

	switch msg.Icon {
	case sets.ErrorIcon:
		d.WithIcon = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns"
	case sets.WarningIcon:
		d.WithIcon = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertCautionIcon.icns"
	case sets.InfoIcon:
		d.WithIcon = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertNoteIcon.icns"
	case sets.QuestionIcon:
		d.WithIcon = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/GenericQuestionMarkIcon.icns"
	}

	if len(msg.ExtraButtons) > 1 {
		return d, cm.ErrorF("Only one additional button is allowed on macOS.")
	}

	// Workaround: Append invisible spaces before each extra button,
	// to identify the index afterwards (no string parsing of the label!).
	for i := range msg.ExtraButtons {
		id := ""
		for j := 0; j < i+1; j++ {
			id += string(idPrefix)
		}
		msg.ExtraButtons[i] = id + msg.ExtraButtons[i]
	}

	if msg.Style == sets.QuestionStyle {
		if strs.IsEmpty(msg.CancelLabel) {
			msg.CancelLabel = "No"
		}
		if strs.IsEmpty(msg.OkLabel) {
			msg.OkLabel = "Yes"
		}

		d.Opts.Buttons = append(msg.ExtraButtons, msg.CancelLabel, msg.OkLabel) // nolint: gocritic
		d.Opts.CancelButton = len(msg.ExtraButtons) + 1

	} else {
		if strs.IsEmpty(msg.OkLabel) {
			msg.OkLabel = "Ok"
		}

		d.Opts.Buttons = append(msg.ExtraButtons, msg.OkLabel) // nolint: gocritic
		d.Opts.DefaultButton = len(msg.ExtraButtons) + 1
	}

	if msg.DefaultCancel && d.Opts.CancelButton != 0 {
		d.Opts.DefaultButton = d.Opts.CancelButton
	}

	return d, nil
}

func getResult(out string, maxExtraButtons int) res.Message {

	r := []rune(out)
	i := 0
	for ; i < maxExtraButtons && i < len(r); i++ {
		if r[i] != idPrefix {
			break
		}
	}

	// No 'idPrefix' found -> its the ok Button.
	if i == 0 {
		return res.Message{General: res.OkResult()}
	}

	// otherwise its an extra button
	return res.Message{General: res.ExtraButtonResult(uint(i))}
}

func ShowMessage(ctx context.Context, s *sets.Message) (res.Message, error) {

	data, err := translateMessage(s)
	if err != nil {
		return res.Message{}, err
	}

	out, err := gmac.RunOSAScript(ctx, "msg", data, "")
	if err == nil {
		return getResult(string(out), len(s.ExtraButtons)), nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 1 {
			return res.Message{General: res.CancelResult()}, nil
		}
	}

	return res.Message{}, err
}
