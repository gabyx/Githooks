package install

import (
	cm "gabyx/githooks/common"
	"gabyx/githooks/hooks"
	strs "gabyx/githooks/strings"
	"path"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

// LoadInstallDir loads the install directory and uses a default if
// it does not exist.
func LoadInstallDir(log cm.ILogContext) (installDir string) {

	installDir = hooks.GetInstallDir()

	if !cm.IsDirectory(installDir) {

		if strs.IsNotEmpty(installDir) {
			log.WarnF("Install directory '%s' does not exist.\n"+
				"Githooks installation is corrupt!\n"+
				"Using default location '~/.githooks'.", installDir)
		}

		home, err := homedir.Dir()
		cm.AssertNoErrorPanic(err, "Could not get home directory.")
		installDir = path.Join(filepath.ToSlash(home), hooks.HooksDirName)
	}

	return
}
