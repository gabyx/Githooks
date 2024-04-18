//go:build !coverage

package hooks

import (
	"path"
	"runtime"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// GetCLIExecutable gets the global Githooks CLI executable.
// If `installDir` is not given the executable name is returned.
func GetCLIExecutable(installDir string) cm.Executable {
	p := CLIName
	if strs.IsNotEmpty(installDir) {
		p = path.Join(GetBinaryDir(installDir), p)
	}

	if runtime.GOOS == cm.WindowsOsName {
		p += cm.WindowsExecutableSuffix
	}

	return cm.Executable{Cmd: p}
}

// GetInstallerExecutable gets the global Githooks installer executable (cli with args).
// If `installDir` is not given the executable name is returned.
func GetInstallerExecutable(installDir string) cm.Executable {
	exec := GetCLIExecutable(installDir)
	exec.Args = []string{"installer"}

	return exec
}

// GetUninstallerExecutable gets the global Githooks uninstaller executable (cli with args).
// If `installDir` is not given the executable name is returned.
func GetUninstallerExecutable(installDir string) cm.Executable {
	exec := GetCLIExecutable(installDir)
	exec.Args = []string{"uninstaller"}

	return exec
}
