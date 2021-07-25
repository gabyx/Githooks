// +build windows

package gui

import (
	"context"
	"reflect"
	"runtime"
	"syscall"
	"unsafe"

	res "github.com/gabyx/githooks/githooks/apps/dialog/result"
	sets "github.com/gabyx/githooks/githooks/apps/dialog/settings"
	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

var (
	shBrowseForFolder           = shell32.NewProc("SHBrowseForFolderW")
	shGetPathFromIDListEx       = shell32.NewProc("SHGetPathFromIDListEx")
	shCreateItemFromParsingName = shell32.NewProc("SHCreateItemFromParsingName")

	coInitializeEx   = ole32.NewProc("CoInitializeEx")
	coUninitialize   = ole32.NewProc("CoUninitialize")
	coCreateInstance = ole32.NewProc("CoCreateInstance")
	coTaskMemFree    = ole32.NewProc("CoTaskMemFree")
)

type BrowseInfo struct {
	Owner        uintptr
	Root         uintptr
	DisplayName  *uint16
	Title        *uint16
	Flags        uint32
	CallbackFunc uintptr
	LParam       uintptr
	Image        int32
}

func pickFolders(ctx context.Context, s *sets.FileSelection) (r res.File, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hr, _, _ := coInitializeEx.Call(0, 0x6) // COINIT_APARTMENTTHREADED|COINIT_DISABLE_OLE1DDE

	if hr != 0x80010106 { // nolint: gomnd // RPC_E_CHANGED_MODE
		if int32(hr) < 0 {
			err = cm.ErrorF("Failed call 'coInitializeEx': error '%v'", syscall.Errno(hr))

			return
		}

		defer coUninitialize.Call() //nolint: errcheck
	}

	var dialog *iFileOpenDialog
	hr, _, _ = coCreateInstance.Call(
		_CLSID_FileOpenDialog, 0, 0x17, // CLSCTX_ALL
		iIDiFileOpenDialog, uintptr(unsafe.Pointer(&dialog)))

	if int32(hr) < 0 {
		return browseForFolder(ctx, s) // use fallback..
	}
	defer dialog.Call(dialog.vtbl.Release) //nolint: errcheck

	var flags int
	hr, _, _ = dialog.Call(dialog.vtbl.GetOptions, uintptr(unsafe.Pointer(&flags)))
	if int32(hr) < 0 {
		err = cm.ErrorF("Failed call 'dialog.GetOptions': error '%v'", syscall.Errno(hr))

		return
	}

	if s.MultipleSelection {
		flags |= 0x200 // FOS_ALLOWMULTISELECT
	}

	if s.ShowHidden {
		flags |= 0x10000000 // FOS_FORCESHOWHIDDEN
	}

	hr, _, _ = dialog.Call(dialog.vtbl.SetOptions, uintptr(flags|0x68)) // nolint: gomnd
	// FOS_NOCHANGEDIR|FOS_PICKFOLDERS|FOS_FORCEFILESYSTEM
	if int32(hr) < 0 {
		err = cm.ErrorF("Failed call 'dialog.SetOptions': error '%v'", syscall.Errno(hr))

		return
	}

	if strs.IsNotEmpty(s.Title) {
		ptr, e := syscall.UTF16PtrFromString(s.Title)
		cm.AssertNoErrorPanic(e, "Conversion string to UTF16 failed")

		hr, _, _ = dialog.Call(dialog.vtbl.SetTitle, uintptr(unsafe.Pointer(ptr)))
		if int32(hr) < 0 {
			err = cm.ErrorF("Failed call 'dialog.SetTitle': error '%v'", syscall.Errno(hr))

			return
		}
	}

	var item *iShellItem
	ptr, err := syscall.UTF16PtrFromString(s.Root)
	cm.AssertNoErrorPanic(err, "Converting string to UTF16 failed.")

	hr, _, _ = shCreateItemFromParsingName.Call(
		uintptr(unsafe.Pointer(ptr)), 0,
		iIDiShellItem,
		uintptr(unsafe.Pointer(&item)))
	if int32(hr) < 0 {
		err = cm.ErrorF("Failed call 'shCreateItemFromParsingName': error '%v'", syscall.Errno(hr))

		return
	}

	if item != nil {
		hr, _, _ = dialog.Call(dialog.vtbl.SetFolder, uintptr(unsafe.Pointer(item)))
		if int32(hr) < 0 {
			err = cm.ErrorF("Failed call 'dialog.SetFolder': error '%v'", syscall.Errno(hr))

			return
		}

		item.Call(item.vtbl.Release) // nolint: errcheck
	}

	// Make the context destroying the window by hooking it.
	if ctx != nil {
		unhook, e := hookDialog(ctx, nil)
		if err = e; err != nil {
			return
		}
		defer unhook()
	}

	activate()

	hr, _, _ = dialog.Call(dialog.vtbl.Show, 0)

	if ctx != nil && ctx.Err() != nil {
		err = ctx.Err()

		return
	}

	if hr == 0x800704c7 { //nolint: gomnd // ERROR_CANCELLED
		return res.File{General: res.CancelResult()}, nil
	}

	if int32(hr) < 0 {
		err = cm.ErrorF("Failed call 'dialog.Show': error '%v'", syscall.Errno(hr))

		return
	}

	var pathsSelected []string

	shellItemPath := func(obj *comObject, trap uintptr, a ...uintptr) {
		var item *iShellItem
		hr, _, _ := obj.Call(trap, append(a, uintptr(unsafe.Pointer(&item)))...)
		if int32(hr) < 0 {
			err = cm.ErrorF("Failed call 'dialog.GetItem': error '%v'", syscall.Errno(hr))

			return
		}

		defer item.Call(item.vtbl.Release) //nolint: errcheck

		var ptr uintptr
		hr, _, _ = item.Call(
			item.vtbl.GetDisplayName,
			0x80058000, // SIGDN_FILESYSPATH
			uintptr(unsafe.Pointer(&ptr)))
		if int32(hr) < 0 {
			err = cm.ErrorF("Failed call 'dialog.GetDisplayName': error '%v'", syscall.Errno(hr))

			return
		}

		defer coTaskMemFree.Call(ptr) //nolint: errcheck

		var res []uint16
		hdr := (*reflect.SliceHeader)(unsafe.Pointer(&res))
		hdr.Data, hdr.Len, hdr.Cap = ptr, maxPath, maxPath
		pathsSelected = append(pathsSelected, syscall.UTF16ToString(res))
	}

	if s.MultipleSelection {
		var items *iShellItemArray
		hr, _, _ = dialog.Call(dialog.vtbl.GetResults, uintptr(unsafe.Pointer(&items)))
		if int32(hr) < 0 {
			err = cm.ErrorF("Failed call 'dialog.GetResults': error '%v'", syscall.Errno(hr))

			return
		}

		defer items.Call(items.vtbl.Release) //nolint: errcheck

		var count uint32
		hr, _, _ = items.Call(items.vtbl.GetCount, uintptr(unsafe.Pointer(&count)))
		if int32(hr) < 0 {
			err = cm.ErrorF("Failed call 'dialog.GetCount': error '%v'", syscall.Errno(hr))

			return
		}

		for i := uintptr(0); i < uintptr(count); i++ {
			shellItemPath(&items.comObject, items.vtbl.GetItemAt, i)
		}

	} else {
		shellItemPath(&dialog.comObject, dialog.vtbl.GetResult)
	}

	return res.File{
		General: res.OkResult(),
		Paths:   pathsSelected}, nil
}

func browseForFolder(ctx context.Context, s *sets.FileSelection) (r res.File, err error) {

	var args BrowseInfo
	args.Flags = 0x1 // BIF_RETURNONLYFSDIRS

	if strs.IsNotEmpty(s.Title) {
		args.Title, err = syscall.UTF16PtrFromString(s.Title)
		cm.AssertNoErrorPanic(err, "Conversion string to UTF16 failed")
	}

	if strs.IsNotEmpty(s.Filename) {
		ptr, err := syscall.UTF16PtrFromString(s.Filename)
		cm.AssertNoErrorPanic(err, "Conversion string to UTF16 failed")

		args.LParam = uintptr(unsafe.Pointer(ptr))

		args.CallbackFunc =
			syscall.NewCallback(
				func(wnd uintptr, msg uint32, lparam, data uintptr) uintptr {
					if msg == 1 { // BFFMiNITIALIZED
						sendMessage.Call( //nolint: errcheck
							wnd, 1024+103, //nolint:  gomnd
							/* BFFM_SETSELECTIONW */
							1, /* TRUE */
							data)
					}

					return 0
				})
	}

	// Make the context destroying the window by hooking it.
	if ctx != nil {
		unhook, e := hookDialog(ctx, nil)
		if err = e; err != nil {
			return
		}

		defer unhook()
	}

	activate()

	ptr, _, _ := shBrowseForFolder.Call(uintptr(unsafe.Pointer(&args)))
	if ctx != nil && ctx.Err() != nil {
		err = ctx.Err()

		return
	}

	if ptr == 0 {
		return res.File{General: res.CancelResult()}, nil
	}
	defer coTaskMemFree.Call(ptr) // nolint: errcheck

	path := make([]uint16, maxPath)
	success, _, _ := shGetPathFromIDListEx.Call(ptr, uintptr(unsafe.Pointer(&path[0])), uintptr(len(path)), 0)
	cm.AssertOrPanic(success == 1, "Could not get path by 'shGetPathFromIDListEx'")

	return res.File{
		General: res.OkResult(),
		Paths:   []string{syscall.UTF16ToString(path)}}, nil
}

func uuid(s string) uintptr {
	return (*reflect.StringHeader)(unsafe.Pointer(&s)).Data
}

var (
	iIDiShellItem         = uuid("\x1e\x6d\x82\x43\x18\xe7\xee\x42\xbc\x55\xa1\xe2\x61\xc3\x7b\xfe")
	iIDiFileOpenDialog    = uuid("\x88\x72\x7c\xd5\xad\xd4\x68\x47\xbe\x02\x9d\x96\x95\x32\xd9\x60")
	_CLSID_FileOpenDialog = uuid("\x9c\x5a\x1c\xdc\x8a\xe8\xde\x4d\xa5\xa1\x60\xf8\x2a\x20\xae\xf7")
)

//nolint: structcheck
type iFileOpenDialog struct {
	comObject
	vtbl *iFileOpenDialogVtbl
}

//nolint: structcheck
type iShellItem struct {
	comObject
	vtbl *iShellItemVtbl
}

//nolint: structcheck
type iShellItemArray struct {
	comObject
	vtbl *iShellItemArrayVtbl
}

//nolint: structcheck
type iFileOpenDialogVtbl struct {
	iFileDialogVtbl
	GetResults       uintptr
	GetSelectedItems uintptr
}

//nolint: structcheck
type iFileDialogVtbl struct {
	iModalWindowVtbl
	SetFileTypes        uintptr
	SetFileTypeIndex    uintptr
	GetFileTypeIndex    uintptr
	Advise              uintptr
	Unadvise            uintptr
	SetOptions          uintptr
	GetOptions          uintptr
	SetDefaultFolder    uintptr
	SetFolder           uintptr
	GetFolder           uintptr
	GetCurrentSelection uintptr
	SetFileName         uintptr
	GetFileName         uintptr
	SetTitle            uintptr
	SetOkButtonLabel    uintptr
	SetFileNameLabel    uintptr
	GetResult           uintptr
	AddPlace            uintptr
	SetDefaultExtension uintptr
	Close               uintptr
	SetClientGuid       uintptr
	ClearClientData     uintptr
	SetFilter           uintptr
}

//nolint: structcheck
type iModalWindowVtbl struct {
	unknownVtbl
	Show uintptr
}

//nolint: structcheck
type iShellItemVtbl struct {
	unknownVtbl
	BindToHandler  uintptr
	GetParent      uintptr
	GetDisplayName uintptr
	GetAttributes  uintptr
	Compare        uintptr
}

//nolint: structcheck
type iShellItemArrayVtbl struct {
	unknownVtbl
	BindToHandler              uintptr
	GetPropertyStore           uintptr
	GetPropertyDescriptionList uintptr
	GetAttributes              uintptr
	GetCount                   uintptr
	GetItemAt                  uintptr
	EnumItems                  uintptr
}
