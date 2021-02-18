// +build windows

package gui

import (
	"context"
	"path/filepath"
	"runtime"
	"syscall"
	"unicode/utf16"
	"unsafe"

	res "gabyx/githooks/apps/dialog/result"
	sets "gabyx/githooks/apps/dialog/settings"
	cm "gabyx/githooks/common"
	strs "gabyx/githooks/strings"

	"github.com/ulule/deepcopier"
)

var (
	getOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
	getSaveFileNameW = comdlg32.NewProc("GetSaveFileNameW")
)

const (
	ofnOverwritePrompt  = 0x00000002
	ofnNoChangeDir      = 0x00000008
	ofnAllowMultiSelect = 0x00000200
	ofnPathMustExist    = 0x00000800
	ofnFileMustExist    = 0x00001000
	ofnCreatePrompt     = 0x00002000
	ofnNoReadOnlyReturn = 0x00008000
	ofnExplorer         = 0x00080000
	ofnForceShowHidden  = 0x10000000
)

// OpenFileNameW https://msdn.microsoft.com/en-us/library/windows/desktop/ms646839.aspx
// nolint: structcheck
type openFileNameW struct {
	lStructSize       uint32
	hwndOwner         syscall.Handle
	hInstance         syscall.Handle
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          syscall.Handle
	lpTemplateName    *uint16
	pvReserved        unsafe.Pointer
	dwReserved        uint32
	flagsEx           uint32
}

func translateFileSave(s *sets.FileSave) (ofn openFileNameW, buf []uint16, err error) {

	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))
	ofn.lpstrTitle, err = syscall.UTF16PtrFromString(s.Title)
	cm.AssertNoErrorPanic(err, "Wrong UTF16 conversion")

	buf = make([]uint16, maxPath)
	ofn.lpstrFile = &buf[0]
	ofn.nMaxFile = uint32(len(buf))

	if strs.IsNotEmpty(s.Filename) {
		f, e := syscall.UTF16FromString(SanitizeFilename(s.Filename))
		cm.AssertNoErrorPanic(e, "Wrong UTF16 conversion")
		copy(buf, f)
	}

	ofn.lpstrInitialDir, err = syscall.UTF16PtrFromString(filepath.FromSlash(s.Root))
	cm.AssertNoErrorPanic(err, "Wrong UTF16 conversion")

	ext := filepath.Ext(s.Filename)
	if strs.IsNotEmpty(ext) {
		ofn.lpstrDefExt, err = syscall.UTF16PtrFromString(ext)
		cm.AssertNoErrorPanic(err, "Wrong UTF16 conversion")
	}

	ofn.flags = ofnNoChangeDir | ofnExplorer | ofnPathMustExist | ofnNoReadOnlyReturn

	if s.ConfirmOverwrite {
		ofn.flags |= ofnOverwritePrompt
	}

	if s.ConfirmCreate {
		ofn.flags |= ofnCreatePrompt
	}

	if s.ShowHidden {
		ofn.flags |= ofnForceShowHidden
	}

	if i := initFilters(s.FileFilters); len(i) != 0 {
		ofn.lpstrFilter = &i[0]
	}

	return ofn, buf, err
}

func translateFileSelection(s *sets.FileSelection) (ofn openFileNameW, buf []uint16, err error) {

	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))
	ofn.lpstrTitle, err = syscall.UTF16PtrFromString(s.Title)
	cm.AssertNoErrorPanic(err, "Wrong UTF16 conversion")

	addSelectedFilenamesChars := 0
	if s.MultipleSelection {
		addSelectedFilenamesChars = 1024 * 256 //nolint: gomnd
	}

	buf = make([]uint16, maxPath+addSelectedFilenamesChars)
	ofn.lpstrFile = &buf[0]
	ofn.nMaxFile = uint32(len(buf))

	if strs.IsNotEmpty(s.Filename) {
		f, e := syscall.UTF16FromString(SanitizeFilename(s.Filename))
		cm.AssertNoErrorPanic(e, "Wrong UTF16 conversion")
		copy(buf, f)
	}

	ofn.lpstrInitialDir, err = syscall.UTF16PtrFromString(filepath.FromSlash(s.Root))
	cm.AssertNoErrorPanic(err, "Wrong UTF16 conversion")

	ext := filepath.Ext(s.Filename)
	if strs.IsNotEmpty(ext) {
		ofn.lpstrDefExt, err = syscall.UTF16PtrFromString(ext)
		cm.AssertNoErrorPanic(err, "Wrong UTF16 conversion")
	}

	ofn.flags = ofnNoChangeDir | ofnExplorer | ofnFileMustExist

	if s.MultipleSelection {
		ofn.flags |= ofnAllowMultiSelect
	}

	if s.ShowHidden {
		ofn.flags |= ofnForceShowHidden
	}

	if i := initFilters(s.FileFilters); len(i) != 0 {
		ofn.lpstrFilter = &i[0]
	}

	return ofn, buf, err
}

// ShowFileSave displays the file save dialog.
func ShowFileSave(ctx context.Context, s *sets.FileSave) (res.File, error) {

	if s.OnlyDirectories {
		ss := sets.FileSelection{}

		err := deepcopier.Copy(s).To(&ss)
		cm.AssertNoErrorPanic(err, "Struct copy failed")

		return pickFolders(ctx, &ss)
	}

	ofn, buf, err := translateFileSave(s)
	if err != nil {
		return res.File{}, err
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Make the context destroying the window by hooking it.
	if ctx != nil {
		unhook, err := hookDialog(ctx, nil)
		if err != nil {
			return res.File{}, err
		}
		defer unhook()
	}

	activate()

	r1, _, _ := getSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))

	if ctx != nil && ctx.Err() != nil {
		return res.File{}, ctx.Err()
	}

	if r1 == 0 {
		err = getDialogError()

		if err != nil {
			return res.File{}, err
		}

		return res.File{General: res.CancelResult()}, nil
	}

	return res.File{
		General: res.OkResult(),
		Paths:   []string{syscall.UTF16ToString(buf)}}, nil

}

func splitSelection(buf []uint16) []string {
	var i int
	var nul bool
	var split []string

	for j, p := range buf {
		if p == 0 {
			if nul {
				break
			}
			if i < j {
				split = append(split, string(utf16.Decode(buf[i:j])))
			}
			i = j + 1
			nul = true
		} else {
			nul = false
		}
	}

	len := len(split)
	if len == 0 {
		return split
	}

	if len--; len > 0 {
		base := split[0]
		for i := 0; i < len; i++ {
			split[i] = filepath.Join(base, string(split[i+1]))
		}
		split = split[:len]
	}

	return split
}

// ShowFileSelection displays a file selection dialog.
func ShowFileSelection(ctx context.Context, s *sets.FileSelection) (res.File, error) {

	if s.OnlyDirectories {
		return pickFolders(ctx, s)
	}

	ofn, buf, err := translateFileSelection(s)
	if err != nil {
		return res.File{}, err
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Make the context destroying the window by hooking it.
	if ctx != nil {
		unhook, err := hookDialog(ctx, nil)
		if err != nil {
			return res.File{}, err
		}
		defer unhook()
	}

	activate()

	r1, _, _ := getOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))

	if ctx != nil && ctx.Err() != nil {
		return res.File{}, ctx.Err()
	}

	if r1 == 0 {
		err = getDialogError()

		if err != nil {
			return res.File{}, err
		}

		return res.File{General: res.CancelResult()}, nil
	}

	// Split the selection
	splitSelection(buf)

	return res.File{
		General: res.OkResult(),
		Paths:   splitSelection(buf)}, nil
}

func initFilters(filters []sets.FileFilter) []uint16 {
	var res string
	for i := range filters {
		res += filters[i].Name
		res += "\x00"

		for j := range filters[i].Patterns {
			res += filters[i].Patterns[j]
			res += ";"
		}
		res += "\x00"
	}

	if strs.IsNotEmpty(res) {
		res += "\x00"
	}

	return utf16.Encode([]rune(res))
}
