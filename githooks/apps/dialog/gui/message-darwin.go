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

const idPrefix rune = '\u200B'

func translateMessage(msg *sets.Message) (d gmac.MsgData, err error) {

	msg.SetDefaultIcons()

	d = gmac.MsgData{}

	d.Operation = "displayDialog"
	d.Text = msg.Text
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

	// Overwrite icon
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
	extraButtons := make([]string, len(msg.ExtraButtons))
	id := string(idPrefix)
	for i := range msg.ExtraButtons {
		id += string(idPrefix)

		if strs.IsEmpty(msg.ExtraButtons[i]) {
			err = cm.ErrorF("Empty label for extra button is not allowed")

			return
		}

		extraButtons[i] = id + msg.ExtraButtons[i]
	}

	if msg.Style == sets.QuestionStyle {
		if strs.IsEmpty(msg.CancelLabel) {
			msg.CancelLabel = "No"
		}
		if strs.IsEmpty(msg.OkLabel) {
			msg.OkLabel = string(idPrefix) + "Yes"
		}

		d.Opts.Buttons = append(extraButtons, msg.CancelLabel, msg.OkLabel) // nolint: gocritic
		d.Opts.CancelButton = len(extraButtons) + 1

	} else {
		if strs.IsEmpty(msg.OkLabel) {
			msg.OkLabel = string(idPrefix) + "Ok"
		}

		d.Opts.Buttons = append(extraButtons, msg.OkLabel) // nolint: gocritic
		d.Opts.DefaultButton = len(extraButtons) + 1
	}

	if msg.DefaultCancel && d.Opts.CancelButton != 0 {
		d.Opts.DefaultButton = d.Opts.CancelButton
	} else {
		d.Opts.DefaultButton = len(d.Opts.Buttons)
	}

	return
}

func getResult(out string, maxButtons int) res.Message {
	s := strings.TrimSpace(out)
	cm.DebugAssert(strs.IsNotEmpty(s))

	r := []rune(s)

	i := 0
	for ; i < maxButtons && i < len(r); i++ {
		if r[i] != idPrefix {
			break
		}
	}

	cm.DebugAssert(i > 0)

	// 1 'idPrefix' found -> its the ok Button.
	if i <= 1 {
		return res.Message{General: res.OkResult()}
	}

	// otherwise its an extra button
	return res.Message{General: res.ExtraButtonResult(uint(i - 1))}
}

func ShowMessage(ctx context.Context, s *sets.Message) (res.Message, error) {

	data, err := translateMessage(s)
	if err != nil {
		return res.Message{}, err
	}

	out, err := gmac.RunOSAScript(ctx, "message", data, "")
	if err == nil {
		return getResult(string(out), len(s.ExtraButtons)+1), nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == 1 {
			return res.Message{General: res.CancelResult()}, nil
		}
	}

	return res.Message{}, err
}
