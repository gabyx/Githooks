// +build windows

package gui

import (
	"context"
	res "gabyx/githooks/apps/dialog/result"
	sets "gabyx/githooks/apps/dialog/settings"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	messageBox       = user32.NewProc("MessageBoxW")
	enumChildWindows = user32.NewProc("EnumChildWindows")
	getDlgCtrlID     = user32.NewProc("GetDlgCtrlID")
	setWindowText    = user32.NewProc("SetWindowTextW")
)

func ShowMessage(ctx context.Context, msg *sets.Message) (r res.Message, err error) {

	msg.SetDefaultIcons()

	if len(msg.ExtraButtons) > 1 {
		err = cm.ErrorF("Only one additional button is allowed on Windows.")

		return
	}

	var flags uintptr

	switch {
	case msg.Style == sets.QuestionStyle && msg.ExtraButtons != nil:
		flags |= 0x3 // MB_YESNOCANCEL
	case msg.Style == sets.QuestionStyle:
		flags |= 0x4 // MB_YESNO
	case msg.ExtraButtons != nil:
		flags |= 0x1 // MB_OKCANCEL
	default:
		flags |= 0 // MB_OK
	}

	switch msg.Icon {
	case sets.ErrorIcon:
		flags |= 0x10 // MB_ICONERROR
	case sets.QuestionIcon:
		flags |= 0x20 // MB_ICONQUESTION
	case sets.WarningIcon:
		flags |= 0x30 // MB_ICONWARNING
	case sets.InfoIcon:
		flags |= 0x40 // MB_ICONINFORMATION
	}

	if msg.Style == sets.QuestionStyle && msg.DefaultCancel {
		if msg.ExtraButtons != nil {
			flags |= 0x100 // MB_DEFBUTTON2
		} else {
			flags |= 0x200 // MB_DEFBUTTON3
		}
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ctx != nil ||
		strs.IsNotEmpty(msg.OkLabel) ||
		strs.IsNotEmpty(msg.CancelLabel) ||
		msg.ExtraButtons != nil {

		unhook, e := hookMessageLabels(ctx, msg.Style, msg.OkLabel, msg.CancelLabel, msg.ExtraButtons)
		if err = e; err != nil {
			return
		}

		defer unhook()
	}

	activate()

	text, err := syscall.UTF16PtrFromString(msg.Text)
	cm.AssertNoErrorPanic(err, "Conversion string to UTF16 failed")

	title, err := syscall.UTF16PtrFromString(msg.Title)
	cm.AssertNoErrorPanic(err, "Conversion string to UTF16 failed")

	success, _, err := messageBox.Call(
		0,
		uintptr(unsafe.Pointer(text)),
		uintptr(unsafe.Pointer(title)), flags)

	if ctx != nil && ctx.Err() != nil {
		err = ctx.Err()

		return
	}

	if success == 0 {
		return
	}

	if success == 7 || (success == 2 && msg.Style != sets.QuestionStyle) { // IDNO
		return res.Message{General: res.ExtraButtonResult(0)}, nil
	}

	if success == 1 || success == 6 { // IDOK, IDYES
		return res.Message{General: res.OkResult()}, nil
	}

	return res.Message{General: res.CancelResult()}, nil
}

func hookMessageLabels(
	ctx context.Context,
	style sets.MessageStyle,
	okLabel string,
	cancelLabel string,
	extraButtons []string) (unhook context.CancelFunc, err error) {

	setButtonNames := func(wnd, lparam uintptr) uintptr {

		name := [8]uint16{}

		getClassName.Call(wnd, uintptr(unsafe.Pointer(&name)), uintptr(len(name))) // nolint: errcheck

		if syscall.UTF16ToString(name[:]) == "Button" {

			ctl, _, _ := getDlgCtrlID.Call(wnd)

			var text string

			// nolint: gomnd
			switch ctl {
			case 1, 6: // IDOK, IDYES
				text = okLabel
			case 2: // IDCANCEL
				cm.AssertOrPanic(extraButtons != nil)
				text = extraButtons[0]
			case 7: // IDNO
				text = cancelLabel
			}

			if strs.IsNotEmpty(text) {
				ptr, _ := syscall.UTF16PtrFromString(text)
				setWindowText.Call(wnd, uintptr(unsafe.Pointer(ptr))) // nolint: errcheck
			}

		}

		return 1
	}

	return hookDialog(ctx,
		func(wnd uintptr) {
			enumChildWindows.Call(wnd, syscall.NewCallback(setButtonNames), 0) // nolint: errcheck
		})
}
