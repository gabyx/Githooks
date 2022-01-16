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

type EntryApp struct {
	*walk.Dialog

	icon     *walk.Icon
	lineEdit *walk.LineEdit

	acceptPB *walk.PushButton
	cancelPB *walk.PushButton
}

func defineEntryButtons(app *EntryApp, entry *sets.Entry, r *res.Entry) []Widget {
	ok := "OK"
	if strs.IsNotEmpty(entry.OkLabel) {
		ok = entry.OkLabel
	}

	cancel := "Cancel"
	if strs.IsNotEmpty(entry.CancelLabel) {
		cancel = entry.CancelLabel
	}

	okCallback := func() {
		*r = res.Entry{
			General: res.OkResult(),
			Text:    app.lineEdit.Text()}
		app.Accept()
	}

	cancelCallback := func() {
		*r = res.Entry{General: res.CancelResult()}
		app.Cancel()
	}

	extraButtonCallback := func(index uint) func() {
		return func() {
			*r = res.Entry{General: res.ExtraButtonResult(index)}
			app.Accept()
		}
	}

	return defineOkCancelButtons(
		ok, cancel, entry.ExtraButtons,
		&app.acceptPB, &app.cancelPB,
		okCallback, cancelCallback, extraButtonCallback)
}

// nolint: gomnd
func defineDefaultEntry(app *EntryApp, opts *sets.Entry, addTextIcon bool) (w []Widget) {

	app.icon = getIcon(opts.WindowIcon)

	if app.icon != nil && addTextIcon {
		bitmap, err := walk.NewBitmapFromIconForDPI(app.icon, walk.Size{Width: 48, Height: 48}, 96)
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
	if opts.NoWrap {
		minSize.Width = 0
	}

	w = append(w, TextLabel{
		RightToLeftReading: true,
		MinSize:            minSize,
		Text:               opts.Text})

	return
}

func defineEntryEdit(app *EntryApp, opts *sets.Entry) Widget {
	return LineEdit{AssignTo: &app.lineEdit, PasswordMode: opts.HideDefaultEntry}
}

// Shows an entry dialog.
// nolint: gomnd
func ShowEntry(ctx context.Context, entry *sets.Entry) (r res.Entry, err error) {

	app := &EntryApp{}

	minSize := Size{Width: 380, Height: 120}
	size := walk.Size{Width: 400, Height: 150}

	if entry.Width != 0 {
		size.Width = int(entry.Width)
	}

	if entry.Height != 0 {
		size.Height = int(entry.Height)
	}

	defaultButton := &app.acceptPB
	cancelButton := &app.cancelPB
	if entry.DefaultCancel {
		defaultButton, cancelButton = cancelButton, defaultButton
	}

	// nolint: gomnd
	m := Dialog{
		AssignTo:      &app.Dialog,
		Title:         entry.Title,
		MinSize:       minSize,
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
				Layout:   HBox{Spacing: 10, MarginsZero: true},
				Children: defineDefaultEntry(app, entry, false),
			},
			defineEntryEdit(app, entry),
			Composite{
				Layout:   HBox{SpacingZero: true, MarginsZero: true},
				Children: defineEntryButtons(app, entry, &r),
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
			r = res.Entry{General: res.CancelResult()}
		}
	})

	centerAndSetSize(app.Dialog, size)

	if entry.ForceTopMost {
		forceTopMost(app.Dialog)
	}

	if ctx != nil {
		watchTimeout(ctx, app.Dialog)
	}

	ret := app.Run()

	if ret == walk.DlgCmdCancel || ret == walk.DlgCmdClose {
		r = res.Entry{General: res.CancelResult()}

		return
	}

	if ctx != nil && ctx.Err() != nil {
		err = ctx.Err()
	}

	return
}
