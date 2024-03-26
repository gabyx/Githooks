package install

import (
	"os"
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// CheckTemplateDir checks the target directory and if valid
// returns the target hook template directory otherwise empty.
// If an error occurs the directory is empty.
func CheckTemplateDir(targetDir string, subFolderIfExists string) (string, error) {
	if strs.IsNotEmpty(targetDir) {

		targetDir, err := cm.ReplaceTilde(targetDir, false)
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
// `init.templateDir` and `core.hooksPath` can already be
// set and controlled by the user.
func FindHookTemplateDir(gitx *git.Context, installMode InstallModeType) (hooksTemplateDir string, err error) {

	switch installMode {
	case InstallModeTypeV.Manual:

		hooksTemplateDir, err = CheckTemplateDir(
			gitx.GetConfig(hooks.GitCKManualTemplateDir, git.GlobalScope), "hooks")

	case InstallModeTypeV.CoreHooksPath:

		hooksTemplateDir, err = CheckTemplateDir(
			gitx.GetConfig(git.GitCKCoreHooksPath, git.GlobalScope), "")

	case InstallModeTypeV.None:
		fallthrough
	case InstallModeTypeV.TemplateDir:

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
		hooksTemplateDir, err = CheckTemplateDir(
			gitx.GetConfig(git.GitCKInitTemplateDir, git.GlobalScope), "hooks")

		if err != nil {
			return
		} else if strs.IsNotEmpty(hooksTemplateDir) {
			return
		}

		// 3. Try setup from the default location
		hooksTemplateDir, err = CheckTemplateDir(path.Join(git.GetDefaultTemplateDir(), "hooks"), "")
	}

	return
}
