//go:build windows

package gui

import (
	"context"
	"sync"
	"syscall"
	"unsafe"
)

const (
	whCallWndProcRet = 12     // WH_CALLWNDPROCRET
	wmInitDialog     = 0x0110 // WM_INITDIALOG
	wmSysCommand     = 0x0112 // WM_SYSCOMMAND
	scClose          = 0xf060 // SC_CLOSE
)

var (
	getCurrentThreadId = kernel32.NewProc("GetCurrentThreadId")

	sendMessage    = user32.NewProc("SendMessageW")
	getClassName   = user32.NewProc("GetClassNameW")
	callNextHookEx = user32.NewProc("CallNextHookEx")

	setWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	unhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
)

type cwPretStruct struct {
	Result  uintptr
	LParam  uintptr
	WParam  uintptr
	Message uint32
	Wnd     uintptr
}

// hookDialog hooks the dialog created with `initDialog`
// such that `ctx` can be canceled and the dialog gets closed.
func hookDialog(
	ctx context.Context,
	initDialog func(wnd uintptr)) (unhook context.CancelFunc, err error) {

	if ctx != nil && ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var mtx sync.Mutex
	var hook, wnd uintptr

	tid, _, _ := getCurrentThreadId.Call()

	hookFunc := func(code int, wparam uintptr, lparam *cwPretStruct) uintptr {

		if lparam.Message == wmInitDialog {

			name := [8]uint16{}
			getClassName.Call(lparam.Wnd, uintptr(unsafe.Pointer(&name)), uintptr(len(name))) //nolint: errcheck

			if syscall.UTF16ToString(name[:]) == "#32770" { // The class for a dialog box
				var close bool

				mtx.Lock()
				if ctx != nil && ctx.Err() != nil {
					close = true
				} else {
					wnd = lparam.Wnd
				}
				mtx.Unlock()

				if close {
					sendMessage.Call(lparam.Wnd, wmSysCommand, scClose, 0) //nolint: errcheck
				} else if initDialog != nil {
					initDialog(lparam.Wnd)
				}
			}
		}

		next, _, _ := callNextHookEx.Call(hook, uintptr(code), wparam, uintptr(unsafe.Pointer(lparam)))

		return next
	}

	hook, _, err = setWindowsHookEx.Call(whCallWndProcRet, syscall.NewCallback(hookFunc), 0, tid)

	if hook == 0 {
		return nil, err
	}

	if ctx == nil {
		return func() { unhookWindowsHookEx.Call(hook) }, nil //nolint: errcheck
	}

	wait := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			mtx.Lock()
			w := wnd
			mtx.Unlock()

			if w != 0 {
				// Send close to the window.
				sendMessage.Call(w, wmSysCommand, scClose, 0) //nolint: errcheck
			}
		case <-wait:
		}
	}()

	return func() {
		unhookWindowsHookEx.Call(hook) //nolint: errcheck
		close(wait)
	}, nil
}
