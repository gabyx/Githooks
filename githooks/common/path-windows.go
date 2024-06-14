//go:build windows

package common

import (
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/hectane/go-acl"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetBinaryTypeW = kernel32.NewProc("GetBinaryTypeW")
)

//nolint: deadcode,varcheck
const (
	scs32BitBinary = 0
	scs64BitBinary = 6
	scsDOSBinary   = 1
	scsOS216Binary = 5
	scsPIFBinary   = 3
	scsPOSIXBinary = 4
	scsWOWBinary   = 2
)

// IsExecutable tests if a `path` is an executable.
// This wraps: https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-getbinarytypew
func IsExecutable(path string) bool {

	if len([]rune(path)) > syscall.MAX_PATH {
		// If the path is to big we send it directly to the filesystem with a prefix `\\?\`
		// see https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file
		p, err := filepath.Abs(path)
		if err != nil {
			return false
		}
		path = `\\?\` + p
	}

	var t uint32
	pPath, err := syscall.UTF16PtrFromString(path)
	AssertNoErrorPanicF(err, "Path cannot be converted to UTF16 Pointer '%s'", path)

	_, _, e := syscall.Syscall(
		procGetBinaryTypeW.Addr(),
		2, //nolint: mnd
		uintptr(unsafe.Pointer(pPath)),
		uintptr(unsafe.Pointer(&t)),
		0)

	return e == 0 && (t == scs32BitBinary || t == scs64BitBinary)
}

// IsWritable tests if a `path` is writable.
func IsWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	if !info.IsDir() {
		return false
	}

	// Check if the user bit is enabled in file permission
	if info.Mode().Perm()&(1<<(uint(7))) == 0 { //nolint: mnd
		return false
	}

	return true
}

// MakeExecutable makes a file executable.
func MakeExecutable(path string) error {
	// On Windows this does not make sense.
	return nil
}

// Chmod is a wrapper around the Windows ACL.
func Chmod(filePath string, mode os.FileMode) error {
	return acl.Chmod(filePath, mode)
}
