package gui

import (
	"context"
	res "gabyx/githooks/apps/dialog/result"
	sets "gabyx/githooks/apps/dialog/settings"
	strs "gabyx/githooks/strings"

	"github.com/lxn/walk"
	"github.com/lxn/win"

	. "github.com/lxn/walk/declarative"
)

type ListModel struct {
	walk.ListModelBase
	options []string
}

type OptionsApp struct {
	*walk.Dialog
	listBox  *walk.ListBox
	list     *ListModel
	multiple bool

	acceptPB *walk.PushButton
	cancelPB *walk.PushButton

	selection []uint
}

func defineList(app *OptionsApp, opts *sets.Options) Widget {
	app.multiple = opts.MultipleSelection

	return ListBox{
		AssignTo:                 &app.listBox,
		Model:                    app.list,
		OnSelectedIndexesChanged: app.listCurrentIndicesChanged,
		MultiSelection:           app.multiple,
	}

}

func defineButtons(app *OptionsApp, opts *sets.Options, r *res.Options) []Widget {

	ok := "Ok"
	if strs.IsNotEmpty(opts.OkLabel) {
		ok = opts.OkLabel
	}

	cancel := "Cancel"
	if strs.IsNotEmpty(opts.CancelLabel) {
		cancel = opts.CancelLabel
	}

	return []Widget{
		PushButton{
			AssignTo: &app.acceptPB,
			Text:     ok,
			OnClicked: func() {
				*r = res.Options{
					General:   res.OkResult(),
					Selection: app.selection}

				app.Accept()
			}},
		PushButton{
			AssignTo: &app.cancelPB,
			Text:     cancel,
			OnClicked: func() {
				*r = res.Options{General: res.CancelResult()}

				app.Cancel()
			}},
	}
}

func centerOnScreen(app *walk.Dialog) {
	wScreen := int(win.GetSystemMetrics(win.SM_CXSCREEN))
	hScreen := int(win.GetSystemMetrics(win.SM_CYSCREEN))

	rect := app.Bounds()
	rect.X = wScreen/2 - rect.Width/2  // nolint: gomnd
	rect.Y = hScreen/2 - rect.Height/2 // nolint: gomnd
	_ = app.SetBounds(rect)
}

func ShowOptions(ctx context.Context, opts *sets.Options) (r res.Options, err error) {

	app := &OptionsApp{list: &ListModel{options: opts.Options}}

	size := Size{Width: 240, Height: 400}    // nolint: gomnd
	minSize := Size{Width: 240, Height: 320} // nolint: gomnd

	if opts.Width != 0 {
		size.Width = int(opts.Width)
	}

	if opts.Height != 0 {
		size.Height = int(opts.Height)
	}

	defaultButton := &app.acceptPB
	cancelButton := &app.cancelPB
	if opts.DefaultCancel {
		defaultButton, cancelButton = cancelButton, defaultButton
	}

	// nolint: gomnd
	m := Dialog{
		AssignTo:      &app.Dialog,
		Title:         opts.Title,
		MinSize:       minSize,
		Size:          size,
		DefaultButton: defaultButton,
		CancelButton:  cancelButton,

		Layout: VBox{
			Spacing: 4,
			Margins: Margins{
				Left:   8,
				Top:    8,
				Bottom: 8,
				Right:  8}},

		Children: []Widget{
			TextLabel{
				RightToLeftReading: true,
				MinSize:            Size{Width: 10, Height: 0},
				Text:               opts.Text},
			defineList(app, opts),
			Composite{
				Layout:   HBox{SpacingZero: true, MarginsZero: true},
				Children: defineButtons(app, opts, &r),
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
			r = res.Options{General: res.CancelResult()}
		}
	})

	app.Form().Activating().Once(func() {
		centerOnScreen(app.Dialog)
		_ = app.BringToTop()
	})

	ret := app.Run()

	if ret == walk.DlgCmdCancel {
		r = res.Options{General: res.CancelResult()}

		return
	}

	return
}

func (app *OptionsApp) listCurrentIndicesChanged() {

	if app.multiple {
		s := app.listBox.SelectedIndexes()
		app.selection = make([]uint, 0, len(s))
		for _, idx := range s {
			app.selection = append(app.selection, uint(idx))
		}
	} else {
		app.selection = []uint{uint(app.listBox.CurrentIndex())}
	}
}

func (m *ListModel) ItemCount() int {
	return len(m.options)
}

func (m *ListModel) Value(index int) interface{} {
	return m.options[index]
}
