package install

import (
	"os"
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// CheckTemplateDir checks the target directory and if valid
// returns the target hook template directory otherwise empty.
// If an error occures the directory is empty.
func CheckTemplateDir(targetDir string, subFolderIfExists string) (string, error) {
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

// FindHookTemplateDir finds the hook template directory.
func FindHookTemplateDir(gitx *git.Context, useCoreHooksPath bool) (hooksTemplateDir string, err error) {
	// 1. Try setup from environment variables
	gitTempDir, exists := os.LookupEnv("GIT_TEMPLATE_DIR")
	if exists {
		if hooksTemplateDir, err = CheckTemplateDir(gitTempDir, "hooks"); err != nil {
			return
		} else if strs.IsNotEmpty(hooksTemplateDir) {
			return
		}
	}

	// 2. Try setup from git config
	if useCoreHooksPath {
		hooksTemplateDir, err = CheckTemplateDir(
			gitx.GetConfig(git.GitCKCoreHooksPath, git.GlobalScope), "")
	} else {
		hooksTemplateDir, err = CheckTemplateDir(
			gitx.GetConfig(git.GitCKInitTemplateDir, git.GlobalScope), "hooks")
	}

	if err != nil {
		return
	} else if strs.IsNotEmpty(hooksTemplateDir) {
		return
	}

	// 3. Try setup from the default location
	hooksTemplateDir, err = CheckTemplateDir(path.Join(git.GetDefaultTemplateDir(), "hooks"), "")

	return
}
