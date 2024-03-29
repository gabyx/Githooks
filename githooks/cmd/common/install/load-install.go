package install

import (
	"path"
	"path/filepath"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/mitchellh/go-homedir"
)

// LoadInstallDir loads the install directory and uses a default if
// it does not exist.
func LoadInstallDir(log cm.ILogContext, gitx *git.Context) (installDir string) {

	installDir = hooks.GetInstallDir(gitx)

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
