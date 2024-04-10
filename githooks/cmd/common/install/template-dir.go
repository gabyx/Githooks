package install

import (
	"os"
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// CheckDirAccess checks the target directory and if valid
// returns the target hook template directory otherwise empty.
// If an error occurs the directory is empty.
func CheckDirAccess(targetDir string, subFolderIfExists string) (string, error) {
	if strs.IsNotEmpty(targetDir) {
		targetDir, err := cm.ReplaceTilde(targetDir)
		if err != nil {
			return "", cm.ErrorF("Could not replace tilde '~' in '%s'.", targetDir)
		}

		if cm.IsWritable(targetDir) {
			return path.Join(targetDir, subFolderIfExists), nil
		}
	}

	return "", nil
}

// FindHooksDir finds the hook directory.
// either set in `GitCKPathForUseCoreHooksPath` or if no install
// use the defaults `GIT_TEMPLATE_DIR` or `init.templateDir` or the default directory.
func FindHooksDir(log cm.ILogContext, gitx *git.Context, haveInstall bool) (hooksDir string, err error) {
	if haveInstall {
		log.Info("Check Githooks installation.")
		path := gitx.GetConfig(hooks.GitCKPathForUseCoreHooksPath, git.GlobalScope)
		hooksDir, err = CheckDirAccess(path, "")
	} else {

		// 1. Try setup from environment variables
		log.Info("Check env. variable 'GIT_TEMPLATE_DIR'.")
		gitTempDir, exists := os.LookupEnv("GIT_TEMPLATE_DIR")
		if exists {
			if hooksDir, err = CheckDirAccess(gitTempDir, "hooks"); err != nil {
				return
			} else if strs.IsNotEmpty(hooksDir) {
				return
			}
		}

		// 2. Try setup from git config
		log.InfoF("Check Git config '%s'.", git.GitCKInitTemplateDir)
		hooksDir, err = CheckDirAccess(
			gitx.GetConfig(git.GitCKInitTemplateDir, git.GlobalScope), "hooks")

		if err != nil {
			return
		} else if strs.IsNotEmpty(hooksDir) {
			return
		}

		// 3. Try setup from the default location
		d := git.GetDefaultTemplateDir()
		log.InfoF("Check Git default template directory '%s'.", d)
		hooksDir, err = CheckDirAccess(path.Join(d, "hooks"), "")
	}

	return
}
