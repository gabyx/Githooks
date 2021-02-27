// +build windows

package gui

import (
	"context"
	res "gabyx/githooks/apps/dialog/result"
	sets "gabyx/githooks/apps/dialog/settings"
	strs "gabyx/githooks/strings"

	"github.com/lxn/walk"

	. "github.com/lxn/walk/declarative"
)

type ListModel struct {
	walk.ListModelBase
	options []string
}

type OptionsApp struct {
	*walk.Dialog

	icon     *walk.Icon
	listBox  *walk.ListBox
	list     *ListModel
	multiple bool

	acceptPB *walk.PushButton
	cancelPB *walk.PushButton
}

func defineList(app *OptionsApp, opts *sets.Options) Widget {
	app.multiple = opts.MultipleSelection

	return ListBox{
		AssignTo:       &app.listBox,
		Model:          app.list,
		MultiSelection: app.multiple,
	}
}

func defineListButtons(app *OptionsApp, opts *sets.Options, r *res.Options) []Widget {
	ok := "OK"
	if strs.IsNotEmpty(opts.OkLabel) {
		ok = opts.OkLabel
	}

	cancel := "Cancel"
	if strs.IsNotEmpty(opts.CancelLabel) {
		cancel = opts.CancelLabel
	}

	okCallback := func() {
		*r = res.Options{
			General: res.OkResult(),
			Options: app.getCurrentSelectedIndices()}
		app.Accept()
	}

	cancelCallback := func() {
		*r = res.Options{General: res.CancelResult()}
		app.Cancel()
	}

	extraButtonCallback := func(index uint) func() {
		return func() {
			*r = res.Options{
				General: res.ExtraButtonResult(index)}
			app.Accept()
		}
	}

	return defineOkCancelButtons(
		ok, cancel, opts.ExtraButtons,
		&app.acceptPB, &app.cancelPB,
		okCallback, cancelCallback, extraButtonCallback)
}

// nolint: gomnd
func defineListText(app *OptionsApp, opts *sets.Options, addTextIcon bool) (w []Widget) {

	switch opts.WindowIcon {
	case sets.InfoIcon:
		app.icon = walk.IconInformation()
	case sets.WarningIcon:
		app.icon = walk.IconWarning()
	case sets.ErrorIcon:
		app.icon = walk.IconError()
	case sets.QuestionIcon:
		app.icon = walk.IconQuestion()
	}

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

	w = append(w, TextLabel{
		RightToLeftReading: true,
		MinSize:            Size{Width: 10, Height: 0},
		Text:               opts.Text})

	return
}

// Shows an options dialog.
func ShowOptions(ctx context.Context, opts *sets.Options) (r res.Options, err error) {

	app := &OptionsApp{list: &ListModel{options: opts.Options}}

	minSize := Size{Width: 240, Height: 240}                        // nolint: gomnd
	size := walk.Size{Width: minSize.Width, Height: minSize.Height} // nolint: gomnd

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
				Children: defineListText(app, opts, false),
			},
			defineList(app, opts),
			Composite{
				Layout:   HBox{SpacingZero: true, MarginsZero: true},
				Children: defineListButtons(app, opts, &r),
			},
		},
	}

	if err = m.Create(nil); err != nil {
		return
	}

	// Select default items
	setDefaultSelection(app, opts)

	app.Disposing().Once(func() {
		// The dialog was closed -> Make it canceled if
		// not yet set (ok or cancel pressed).
		if r.IsUnset() {
			r = res.Options{General: res.CancelResult()}
		}
	})

	centerAndSetSize(app.Dialog, size)

	if ctx != nil {
		watchTimeout(ctx, app.Dialog)
	}

	ret := app.Run()

	if ret == walk.DlgCmdCancel || ret == walk.DlgCmdClose {
		r = res.Options{General: res.CancelResult()}

		return
	}

	if ctx != nil && ctx.Err() != nil {
		err = ctx.Err()
	}

	return
}

func setDefaultSelection(app *OptionsApp, opts *sets.Options) {
	dO := len(opts.DefaultOptions)

	indices := make([]int, 0, dO)
	for i := range opts.DefaultOptions {
		if i >= len(opts.Options) {
			continue
		}
		indices = append(indices, int(opts.DefaultOptions[i]))
	}

	if opts.MultipleSelection {
		app.listBox.SetSelectedIndexes(indices)
	} else if len(indices) > 0 {
		_ = app.listBox.SetCurrentIndex(indices[len(indices)-1])
	}
}

func (app *OptionsApp) getCurrentSelectedIndices() (indices []uint) {

	if app.multiple {
		s := app.listBox.SelectedIndexes()
		indices = make([]uint, 0, len(s))
		for _, idx := range s {
			indices = append(indices, uint(idx))
		}
	} else {
		indices = []uint{uint(app.listBox.CurrentIndex())}
	}

	return
}

func (m *ListModel) ItemCount() int {
	return len(m.options)
}

func (m *ListModel) Value(index int) interface{} {
	return m.options[index]
}
