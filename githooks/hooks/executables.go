// +build !coverage

package hooks

import (
	"path"
	"runtime"

	cm "github.com/gabyx/githooks/githooks/common"
)

// GetCLIExecutable gets the global Githooks CLI executable.
func GetCLIExecutable(installDir string) cm.Executable {
	p := path.Join(GetBinaryDir(installDir), "cli")
	if runtime.GOOS == cm.WindowsOsName {
		p += cm.WindowsExecutableSuffix
	}

	return cm.Executable{Cmd: p}
}

// GetInstallerExecutable gets the global Githooks installer executable (cli with args).
func GetInstallerExecutable(installDir string) cm.Executable {
	exec := GetCLIExecutable(installDir)
	exec.Args = []string{"installer"}

	return exec
}

// GetUninstallerExecutable gets the global Githooks uninstaller executable (cli with args).
func GetUninstallerExecutable(installDir string) cm.Executable {
	exec := GetCLIExecutable(installDir)
	exec.Args = []string{"uninstaller"}

	return exec
}
