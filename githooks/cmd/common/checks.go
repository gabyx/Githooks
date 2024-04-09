package ccm

import (
	"github.com/gabyx/githooks/githooks/cmd/common/install"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
)

// CheckGithooksSetup tests if 'core.hooksPath' is in alignment with 'git.GitCKUseCoreHooksPath'.
func CheckGithooksSetup(gitx *git.Context) (err error) {

	useCoreHooksPath := gitx.GetConfig(hooks.GitCKInstallMode, git.GlobalScope) ==
		install.InstallModeTypeV.UseGlobalCoreHooksPath.Name()

	coreHooksPath, coreHooksPathSet := gitx.LookupConfig(git.GitCKCoreHooksPath, git.Traverse)

	if coreHooksPathSet {
		if useCoreHooksPath {
			err = cm.ErrorF(
				"Git config 'core.hooksPath' is set and has value:\n"+
					"'%s',\n"+
					"but Githooks is not configured to use that folder.\n"+
					"This could mean the hooks in this repository are not run by Githooks.", coreHooksPath)
		}
	} else {
		if useCoreHooksPath {
			err = cm.ErrorF(
				"Githooks is configured to consider Git config 'core.hooksPath'\n" +
					"but that setting is not currently set.\n" +
					"This could mean the hooks in this repository are not run by Githooks.")
		}
	}

	return
}
