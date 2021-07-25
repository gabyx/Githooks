// +build windows

package gui

import (
	"syscall"
	"unsafe"

	cm "github.com/gabyx/githooks/githooks/common"
)

type comObject struct{}

func (o *comObject) Call(trap uintptr, a ...uintptr) (r1, r2 uintptr, lastErr error) {
	self := uintptr(unsafe.Pointer(o))
	nargs := uintptr(len(a))
	// nolint: gomnd
	switch nargs {
	case 0:
		return syscall.Syscall(trap, nargs+1, self, 0, 0)
	case 1:
		return syscall.Syscall(trap, nargs+1, self, a[0], 0)
	case 2:
		return syscall.Syscall(trap, nargs+1, self, a[0], a[1])
	default:
		cm.Panic("COM call with too many arguments.")
	}

	return
}

type unknownVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
}
