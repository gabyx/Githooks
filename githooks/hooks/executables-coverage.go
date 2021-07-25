// +build coverage

package hooks

import (
	"os"
	"path"
	"runtime"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/coverage"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// GetCLIExecutable gets the global Githooks CLI executable.
func GetCLIExecutable(installDir string) cm.Executable {
	p := path.Join(GetBinaryDir(installDir), "cli")
	if runtime.GOOS == cm.WindowsOsName {
		p += cm.WindowsExecutableSuffix
	}

	coverDir, _, covData := coverage.ReadCoverData("cli")

	if _, exists := os.LookupEnv("GH_DEPLOY_SOURCE_IS_PROD"); exists {
		return cm.Executable{Cmd: p}
	} else {
		return cm.Executable{
			Cmd: p,
			Args: []string{"-test.coverprofile",
				path.Join(coverDir, strs.Fmt("cli-%v.cov", covData.Counter)),
				"githooksCoverage"}}
	}
}

// GetInstallerExecutable gets the global Githooks installer executable (cli with args).
func GetInstallerExecutable(installDir string) cm.Executable {
	exec := GetCLIExecutable(installDir)
	exec.Args = append(exec.Args, "installer")

	return exec
}

// GetUninstallerExecutable gets the global Githooks uninstaller executable (cli with args).
func GetUninstallerExecutable(installDir string) cm.Executable {
	exec := GetCLIExecutable(installDir)
	exec.Args = append(exec.Args, "uninstaller")

	return exec
}
