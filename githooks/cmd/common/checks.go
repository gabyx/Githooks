package ccm

import (
	"path"

	"github.com/gabyx/githooks/githooks/cmd/common/install"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// CheckGithooksSetup tests if 'core.hooksPath' is in alignment with the install.
func CheckGithooksSetup(log cm.ILogContext, gitx *git.Context) {
	repoRoot, _, _, err := gitx.GetRepoRoot()
	insideRepo := strs.IsNotEmpty(repoRoot) && err == nil

	haveInstall, installMode := install.GetInstallMode(gitx)
	pathToUse := gitx.GetConfig(hooks.GitCKPathForUseCoreHooksPath, git.GlobalScope)
	globalCoreHooksPath, globalCoreHooksPathSet := gitx.LookupConfig(git.GitCKCoreHooksPath, git.GlobalScope)
	localCoreHooksPath, localCoreHooksPathSet := gitx.LookupConfig(git.GitCKCoreHooksPath, git.LocalScope)

	if !haveInstall {
		log.WarnF("Githooks seems not installed. Please install it.")

		return
	} else if strs.IsEmpty(pathToUse) {
		log.WarnF("Githooks install is corrupt: Global Git config '%s' is empty.", hooks.GitCKPathForUseCoreHooksPath)

		return
	}

	if insideRepo {
		hasHooksConfigured, _ := cm.IsPathExisting(hooks.GetGithooksDir(repoRoot))

		if hasHooksConfigured && localCoreHooksPathSet && localCoreHooksPath != pathToUse {
			log.WarnF(
				"Local Git config 'core.hooksPath' is set to:\n"+
					"'%s',\n"+
					"Githooks however uses the maintained run-wrappers in path:\n"+
					"'%s'.\n"+
					"Hooks configured for Githooks in this repository will not run!",
				localCoreHooksPath, pathToUse)
		}

		gitDir, e := gitx.GetGitDirCommon()
		log.AssertNoErrorF(e, "Could not determine common Git dir.")
		hasRunWrappers, _ := cm.IsPathExisting(path.Join(gitDir, "hooks", hooks.RunWrapperMarkerFileName))

		if hasHooksConfigured &&
			!localCoreHooksPathSet && !globalCoreHooksPathSet &&
			!hasRunWrappers {
			log.WarnF("Githooks are configured but Githooks seems not installed in '%v'.\n"+
				"Neither 'core.hooksPath' set nor run-wrappers installed.\n"+
				"Hooks might not run!", gitDir)
		}
	}

	if installMode == install.InstallModeTypeV.Centralized &&
		globalCoreHooksPathSet && globalCoreHooksPath != pathToUse {
		log.ErrorF("Githooks install is corrupt: \n"+
			"Global Git config 'core.hooksPath' is set to:\n"+
			"'%s'\n"+
			"Githooks however uses the maintained run-wrappers in path:\n"+
			"'%s'.\n"+
			"Hooks configured for Githooks might not run!",
			globalCoreHooksPath, pathToUse)
	}
}
