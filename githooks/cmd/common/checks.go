package ccm

import (
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
		hasHooksDir, _ := cm.IsPathExisting(hooks.GetGithooksDir(repoRoot))

		if hasHooksDir && localCoreHooksPathSet && localCoreHooksPath != pathToUse {
			log.WarnF(
				"Local Git config 'core.hooksPath' is set to:\n"+
					"'%s',\n"+
					"Githooks however uses the maintained run-wrappers in path:\n"+
					"'%s'\n."+
					"Hooks configured for Githooks in this repository will not run.",
				localCoreHooksPath, pathToUse)
		}
	}

	if installMode == install.InstallModeTypeV.UseGlobalCoreHooksPath &&
		globalCoreHooksPathSet && globalCoreHooksPath != pathToUse {

		log.ErrorF("Githooks install is corrupt: \n"+
			"Global 'core.hooksPath' is set to:\n"+
			"'%s'\n"+
			"Githooks however uses the maintained run-wrappers in path:\n"+
			"'%s'\n."+
			"Hooks configured for Githooks might not run.",
			globalCoreHooksPath, pathToUse)
	}
}
