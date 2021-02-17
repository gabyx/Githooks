// +build windows

package gui

import (
	cm "gabyx/githooks/common"
	"os"
	"syscall"
	"unsafe"
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
