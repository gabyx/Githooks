//go:build windows

package gui

import (
	"context"

	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	sets "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/lxn/walk"

	. "github.com/lxn/walk/declarative"
)

type MessageApp struct {
	*walk.Dialog

	icon *walk.Icon

	acceptPB *walk.PushButton
	cancelPB *walk.PushButton
}

func defineMessageButtons(app *MessageApp, msg *sets.Message, r *res.Message) []Widget {
	ok := "OK"
	if msg.Style == sets.QuestionStyle {
		ok = "Yes"
	}
	if strs.IsNotEmpty(msg.OkLabel) {
		ok = msg.OkLabel
	}

	okCallback := func() {
		*r = res.Message{
			General: res.OkResult()}
		app.Accept()
	}

	cancel := ""
	var cancelCallback func()
	if msg.Style == sets.QuestionStyle {
		cancel = "No"

		if strs.IsNotEmpty(msg.CancelLabel) {
			cancel = msg.CancelLabel
		}

		cancelCallback = func() {
			*r = res.Message{General: res.CancelResult()}
			app.Cancel()
		}
	}

	extraButtonCallback := func(index uint) func() {
		return func() {
			*r = res.Message{General: res.ExtraButtonResult(index)}
			app.Accept()
		}
	}

	return defineOkCancelButtons(
		ok, cancel, msg.ExtraButtons,
		&app.acceptPB, &app.cancelPB,
		okCallback, cancelCallback, extraButtonCallback)
}

// nolint: gomnd
func defineMessageText(app *MessageApp, msg *sets.Message, addTextIcon bool) (w []Widget) {

	app.icon = getIcon(msg.WindowIcon)

	icon := getIcon(msg.Icon)
	if icon != nil && addTextIcon {

		bitmap, err := walk.NewBitmapFromIconForDPI(icon, walk.Size{Width: 48, Height: 48}, 96)
		if err == nil {
			w = append(w, ImageView{
				Background: TransparentBrush{},
				Image:      bitmap,
				Mode:       ImageViewModeCenter,
				MinSize:    Size{Width: 48, Height: 48},
				MaxSize:    Size{Width: 48, Height: 48},
			})
		}
	}

	minSize := Size{Width: 10, Height: 0}
	if msg.NoWrap {
		minSize.Width = 0
	}

	w = append(w, TextLabel{
		RightToLeftReading: true,
		MinSize:            minSize,
		Text:               msg.Text})

	return
}

// Shows an Message dialog.
// nolint: gomnd
func ShowMessage(ctx context.Context, msg *sets.Message) (r res.Message, err error) {

	app := &MessageApp{}

	msg.SetDefaultIcons()

	minSize := Size{Width: 380, Height: 120}
	size := walk.Size{Width: 400, Height: 150}

	if msg.Width != 0 {
		size.Width = int(msg.Width)
	}

	if msg.Height != 0 {
		size.Height = int(msg.Height)
	}

	defaultButton := &app.acceptPB
	cancelButton := &app.cancelPB
	if msg.DefaultCancel {
		defaultButton, cancelButton = cancelButton, defaultButton
	}

	// nolint: gomnd
	m := Dialog{
		AssignTo:      &app.Dialog,
		Title:         msg.Title,
		MinSize:       minSize,
		Size:          Size{Width: size.Width, Height: size.Height},
		DefaultButton: defaultButton,
		CancelButton:  cancelButton,

		Layout: VBox{
			Spacing: 8,
			Margins: Margins{
				Left:   12,
				Top:    12,
				Bottom: 12,
				Right:  12}},

		Children: []Widget{
			Composite{
				Layout:   HBox{Spacing: 15, MarginsZero: true},
				Children: defineMessageText(app, msg, true),
			},
			Composite{
				Layout:   HBox{SpacingZero: true, MarginsZero: true},
				Children: defineMessageButtons(app, msg, &r),
			},
		},
	}

	if err = m.Create(nil); err != nil {
		return
	}

	app.Disposing().Once(func() {
		// The dialog was closed -> Make it canceled if
		// not yet set (ok or cancel pressed).
		if r.IsUnset() {
			r = res.Message{General: res.CancelResult()}
		}
	})

	if app.icon != nil {
		_ = app.SetIcon(app.icon)
	}

	centerAndSetSize(app.Dialog, size)
	if msg.ForceTopMost {
		forceTopMost(app.Dialog)
	}

	if ctx != nil {
		watchTimeout(ctx, app.Dialog)
	}

	ret := app.Run()

	if ret == walk.DlgCmdCancel || ret == walk.DlgCmdClose {
		r = res.Message{General: res.CancelResult()}

		return
	}

	if ctx != nil && ctx.Err() != nil {
		err = ctx.Err()
	}

	return
}
