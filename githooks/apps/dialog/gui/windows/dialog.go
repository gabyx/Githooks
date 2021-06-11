// +build windows

package gui

import (
	"os"
	"syscall"
	"unsafe"

	sets "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	cm "github.com/gabyx/githooks/githooks/common"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

var (
	getConsoleWindow = kernel32.NewProc("GetConsoleWindow")

	enumWindows              = user32.NewProc("EnumWindows")
	getWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	commDlgExtendedError     = comdlg32.NewProc("CommDlgExtendedError")

	setForegroundWindow = user32.NewProc("SetForegroundWindow")
)

// activate activates the current window in this thread.
func activate() {
	var hwnd uintptr

	//nolint: errcheck
	enumWindows.Call(
		syscall.NewCallback(func(wnd, lparam uintptr) uintptr {
			var pid uintptr
			getWindowThreadProcessId.Call(wnd, uintptr(unsafe.Pointer(&pid)))
			if int(pid) == os.Getpid() {
				hwnd = wnd

				return 0
			}

			return 1
		}), 0)

	if hwnd == 0 {
		hwnd, _, _ = getConsoleWindow.Call()
	}

	if hwnd != 0 {
		setForegroundWindow.Call(hwnd) //nolint: errcheck
	}
}

// getDialogError gets the common dialog error which happened.
func getDialogError() error {
	s, _, _ := commDlgExtendedError.Call()

	if s == 0 {
		return nil
	} else {
		return cm.ErrorF("Common Dialog Error: '%x'", s)
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

func addExtraButtons(extraButtons []string, extraCallback func(idx uint) func()) (w []Widget) {
	for i, button := range extraButtons {
		var e *walk.PushButton
		w = append(w, PushButton{
			AssignTo:  &e,
			Text:      button,
			OnClicked: extraCallback(uint(i))})
	}

	return
}

func defineOkCancelButtons(
	ok string,
	cancel string,
	extraButtons []string,
	acceptPB **walk.PushButton,
	cancelPB **walk.PushButton,
	okCallback func(),
	cancelCallback func(),
	extraCallback func(uint) func()) (w []Widget) {

	w = []Widget{
		HSpacer{},
		PushButton{
			AssignTo:  acceptPB,
			Text:      ok,
			OnClicked: okCallback},
	}

	if cancelCallback != nil {
		w = append(w, PushButton{
			AssignTo:  cancelPB,
			Text:      cancel,
			OnClicked: cancelCallback})
	}

	if extraButtons != nil {
		w = append(w, addExtraButtons(extraButtons, extraCallback)...)
	}

	return
}

func getIcon(icon sets.DialogIcon) *walk.Icon {
	switch icon {
	case sets.InfoIcon:
		return walk.IconInformation()
	case sets.WarningIcon:
		return walk.IconWarning()
	case sets.ErrorIcon:
		return walk.IconError()
	case sets.QuestionIcon:
		return walk.IconQuestion()
	}

	return nil
}

func centerAndSetSize(dlg *walk.Dialog, size walk.Size) {
	_ = dlg.BringToTop()
	dlg.Layout()
	_ = dlg.SetSize(size)
	centerOnScreen(dlg)

	dlg.Form().Activating().Once(func() {
		_ = dlg.SetSize(size)
		centerOnScreen(dlg)
	})
}
