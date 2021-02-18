package gui

import (
	"context"
	sets "gabyx/githooks/apps/dialog/settings"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	shellNotifyIcon = shell32.NewProc("Shell_NotifyIconW")
	wtsSendMessage  = wtsapi32.NewProc("WTSSendMessageW")
)

// ShowNotification shows a notifaction in the task bar.
func ShowNotification(ctx context.Context, s *sets.Notification) error {

	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}

	var data notifyIconData
	data.StructSize = uint32(unsafe.Sizeof(data))
	data.ID = 0x378eb49c    // Random
	data.Flags = 0x00000010 // NIF_INFO
	data.State = 0x00000001 // NIS_HIDDEN

	info, err := syscall.UTF16FromString(s.Text)
	cm.AssertNoErrorPanic(err, "Conversion string to UTF16 failed")

	copy(data.Info[:len(data.Info)-1], info)

	title, err := syscall.UTF16FromString(s.Title)
	cm.AssertNoErrorPanic(err, "Conversion string to UTF16 failed")

	copy(data.InfoTitle[:len(data.InfoTitle)-1], title)

	switch s.WindowIcon {
	case sets.InfoIcon:
		data.InfoFlags |= 0x1 // NIIF_INFO
	case sets.WarningIcon:
		data.InfoFlags |= 0x2 // NIIF_WARNING
	case sets.ErrorIcon:
		data.InfoFlags |= 0x3 // NIIF_ERROR
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	success, _, err := shellNotifyIcon.Call(
		0, // NIM_ADD
		uintptr(unsafe.Pointer(&data)))

	if success == 0 {
		if errno, ok := err.(syscall.Errno); ok && errno == 0 {
			return wtsMessage(s)
		}

		return err
	}

	// nolint: errcheck,gomnd
	shellNotifyIcon.Call(
		2, // NIM_DELETE
		uintptr(unsafe.Pointer(&data)))

	return nil
}

func wtsMessage(s *sets.Notification) error {
	var flags uintptr

	switch s.WindowIcon {
	case sets.ErrorIcon:
		flags |= 0x10 // MB_ICONERROR
	case sets.QuestionIcon:
		flags |= 0x20 // MB_ICONQUESTION
	case sets.WarningIcon:
		flags |= 0x30 // MB_ICONWARNING
	case sets.InfoIcon:
		flags |= 0x40 // MB_ICONINFORMATION
	}

	title := s.Title
	if strs.IsEmpty(title) {
		title = "Notification"
	}

	timeout := 10

	ptext, err := syscall.UTF16FromString(s.Text)
	cm.AssertNoErrorPanic(err, "Conversion string to UTF16 failed")

	ptitle, err := syscall.UTF16FromString(title)
	cm.AssertNoErrorPanic(err, "Conversion string to UTF16 failed")

	var res uint32
	// nolint: gomnd
	success, _, err := wtsSendMessage.Call(
		0,          // WTS_CURRENT_SERVER_HANDLE
		0xffffffff, // WTS_CURRENT_SESSION
		uintptr(unsafe.Pointer(&ptitle[0])), uintptr(2*len(ptitle)),
		uintptr(unsafe.Pointer(&ptext[0])), uintptr(2*len(ptext)),
		flags, uintptr(timeout), uintptr(unsafe.Pointer(&res)), 0)

	if success == 0 {
		return err
	}

	return nil
}

type notifyIconData struct {
	StructSize      uint32
	Owner           uintptr
	ID              uint32
	Flags           uint32
	CallbackMessage uint32
	Icon            uintptr
	Tip             [128]uint16
	State           uint32
	StateMask       uint32
	Info            [256]uint16
	Version         uint32
	InfoTitle       [64]uint16
	InfoFlags       uint32
	// GuidItem     [16]byte
	// BalloonIcon  uintptr
}
