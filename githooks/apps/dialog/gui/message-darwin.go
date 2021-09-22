//go:build darwin

package gui

import (
	"context"
	"os/exec"

	gmac "github.com/gabyx/githooks/githooks/apps/dialog/gui/darwin"
	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	sets "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

const (
	infoIcon     = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertNoteIcon.icns"
	warningIcon  = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertCautionIcon.icns"
	warningIcon2 = "/System/Library/UserNotifications/Bundles/com.apple.notificationcenter.askpermissions.bundle/Contents/Resources/AlertCautionIcon.icns" // nolint
	errorIcon    = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/AlertStopIcon.icns"
	questionIcon = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/GenericQuestionMarkIcon.icns"
)

func getDefaultInfoIcon() string {
	if cm.IsFile(infoIcon) {
		return infoIcon
	}

	return "info"
}

func getDefaultWarningIcon() string {
	if cm.IsFile(warningIcon) {
		return warningIcon
	} else if cm.IsFile(warningIcon2) {
		return warningIcon2
	}

	return "caution"
}

func getDefaultErrorIcon() string {
	if cm.IsFile(errorIcon) {
		return errorIcon
	}

	return "stop"
}

func getDefaultQuestionIcon() string {
	if cm.IsFile(questionIcon) {
		return questionIcon
	}

	return "info"
}

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
		d.Opts.WithIcon = getDefaultInfoIcon()
	case sets.ErrorStyle:
		d.Opts.WithIcon = getDefaultErrorIcon()
	case sets.WarningStyle:
		d.Opts.WithIcon = getDefaultWarningIcon()
	case sets.QuestionStyle:
		d.Opts.WithIcon = getDefaultQuestionIcon()
	}

	// Overwrite icon
	switch msg.Icon {
	default:
		fallthrough
	case sets.InfoIcon:
		d.Opts.WithIcon = getDefaultInfoIcon()
	case sets.ErrorIcon:
		d.Opts.WithIcon = getDefaultErrorIcon()
	case sets.WarningIcon:
		d.Opts.WithIcon = getDefaultWarningIcon()
	case sets.QuestionIcon:
		d.Opts.WithIcon = getDefaultQuestionIcon()
	}

	if len(msg.ExtraButtons) > 1 {
		return d, cm.ErrorF("Only one additional button is allowed on macOS.")
	}

	extraButtons, err := addInvisiblePrefix(msg.ExtraButtons)
	if err != nil {
		return
	}

	if msg.Style == sets.QuestionStyle {
		if strs.IsEmpty(msg.CancelLabel) {
			msg.CancelLabel = "No"
		}
		if strs.IsEmpty(msg.OkLabel) {
			msg.OkLabel = "Yes"
		}

		msg.OkLabel = string(idPrefix) + msg.OkLabel

		d.Opts.Buttons = append(extraButtons, msg.CancelLabel, msg.OkLabel) // nolint: gocritic
		d.Opts.CancelButton = len(extraButtons) + 1

	} else {
		if strs.IsEmpty(msg.OkLabel) {
			msg.OkLabel = "OK"
		}

		msg.OkLabel = string(idPrefix) + msg.OkLabel

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

func ShowMessage(ctx context.Context, s *sets.Message) (res.Message, error) {

	data, err := translateMessage(s)
	if err != nil {
		return res.Message{}, err
	}

	out, err := gmac.RunOSAScript(ctx, "message", data, "")
	if err == nil {
		return res.Message{
			General: getResultButtons(string(out),
				len(s.ExtraButtons)+1)}, nil
	}

	if err, ok := err.(*exec.ExitError); ok {
		if err.ExitCode() == gmac.ExitCodeCancel {
			return res.Message{General: res.CancelResult()}, nil
		}
	}

	return res.Message{}, err
}
