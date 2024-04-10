package install

import (
	"os"
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// InstallIntoRepo set the `core.hooksPath` or
// installs run-wrappers into a repositories.
//
// Setting `core.hooksPath` with `useCoreHooksPath` to use the
// Githooks maintained hooks directory is the
// preferred way. It can not be combined with `hookNames`
// since this only works with installing run-wrappers directly.
// Otherwise we install run-wrappers directly.
// It prompts for disabling detected LFS hooks and offers to
// setup a README file.
func InstallIntoRepo(
	log cm.ILogContext,
	repoGitDir string,
	lfsHooksCache hooks.LFSHooksCache,
	hookNames []string,
	nonInteractive bool,
	dryRun bool,
	skipReadme bool,
	uiSettings *UISettings,
) bool {

	hookDir := path.Join(repoGitDir, "hooks")
	if !cm.IsDirectory(hookDir) {
		err := os.MkdirAll(hookDir, cm.DefaultFileModeDirectory)
		log.AssertNoErrorPanic(err,
			"Could not create hook directory in '%s'.", repoGitDir)
	}
	gitx := git.NewCtxAt(repoGitDir)
	isBare := gitx.IsBareRepo()

	// Check if this repository is setup to install only run-wrappers.
	// We switch to run-wrappers if we install a set of maintained hooks.
	installRunWrappers, _ := cm.IsPathExisting(path.Join(hookDir, ".githooks-contains-run-wrappers"))
	installRunWrappers = installRunWrappers || len(hookNames) != 0

	var err error
	if len(hookNames) == 0 {
		hookNames, _, err = hooks.GetMaintainedHooks(gitx, git.LocalScope)
		log.AssertNoErrorF(err, "Could not get maintained hooks.")
	}

	if isBare {
		// Filter out all non-relevant hooks for bare repositories.
		hookNames = strs.Filter(hookNames, func(s string) bool { return strs.Includes(hooks.ManagedServerHookNames, s) })
		// LFS hooks also do not need to be reinstalled
		lfsHooksCache = nil
	}

	if dryRun {
		log.InfoF("[dry run] Hooks would have been installed into\n'%s'.",
			repoGitDir)

		return false
	}

	if installRunWrappers {

		nLFSHooks, err := hooks.InstallRunWrappers(
			hookDir, hookNames,
			nil,
			GetHookDisableCallback(log, gitx, nonInteractive, uiSettings),
			lfsHooksCache,
			nil)

		log.AssertNoErrorPanicF(err, "Could not install run-wrappers into '%s'.", hookDir)

		if nLFSHooks != 0 {
			log.InfoF("Installed '%v' Githooks run-wrapper(s) and '%v' missing LFS hooks into '%s'.",
				len(hookNames), nLFSHooks, hookDir)
		} else {
			log.InfoF("Installed '%v' Githooks run-wrapper(s) into '%s'",
				len(hookNames), hookDir)
		}

	} else {
		err := hooks.InstallLinkRunWrappers(gitx, hookDir)
		log.AssertNoErrorPanicF(err, "Could not install run-wrapper link into '%s'.", repoGitDir)

		log.InfoF("Installed Githooks run-wrapper link ('%s') into '%s'",
			git.GitCKCoreHooksPath, hookDir)
	}

	// Offer to setup the intro README if running in interactive mode
	// Let's skip this in non-interactive mode or in a bare repository
	// to avoid polluting the repos with README files
	if !skipReadme && !nonInteractive && !isBare {
		setupReadme(log, repoGitDir, dryRun, uiSettings)
	}

	return true
}

func cleanArtefactsInRepo(log cm.ILogContext, gitDir string) {

	// Remove checksum files...
	cacheDir := hooks.GetChecksumDirectoryGitDir(gitDir)
	if cm.IsDirectory(cacheDir) {
		log.AssertNoErrorF(os.RemoveAll(cacheDir),
			"Could not delete checksum cache dir '%s'.", cacheDir)
	}

	ignoreFile := hooks.GetHookIgnoreFileGitDir(gitDir)
	if cm.IsDirectory(ignoreFile) {
		log.AssertNoErrorF(os.RemoveAll(ignoreFile),
			"Could not delete ignore file '%s'.", ignoreFile)
	}
}

func cleanGitConfigInRepo(log cm.ILogContext, gitDir string) {
	gitx := git.NewCtxAt(gitDir)

	for _, k := range hooks.GetLocalGitConfigKeys() {

		log.AssertNoErrorF(gitx.UnsetConfig(k, git.LocalScope),
			"Could not unset Git config '%s' in '%s'.", k, gitDir)

	}
}

func unregisterRepo(log cm.ILogContext, gitDir string) {
	gitx := git.NewCtxAt(gitDir)

	log.AssertNoErrorF(hooks.UnmarkRepoRegistered(gitx),
		"Could not unregister Git repo '%s'.", gitDir)
}

// UninstallFromRepo uninstalls run-wrappers from the repositories Git directory.
// LFS hooks will be reinstalled if available.
func UninstallFromRepo(
	log cm.ILogContext,
	gitDir string,
	lfsHooksCache hooks.LFSHooksCache,
	cleanArtefacts bool) bool {

	hookDir := path.Join(gitDir, "hooks")
	gitx := git.NewCtxAt(gitDir)

	var err error
	var nLfsCount int

	if cm.IsDirectory(hookDir) {

		// We always uninstasll run-wrappers if any are existing.
		// no need to check the marker file `.githooks-contains-run-wrappers`.
		nLfsCount, err = hooks.UninstallRunWrappers(hookDir, lfsHooksCache)

		log.AssertNoErrorF(err,
			"Could not uninstall Githooks run-wrappers from\n'%s'.",
			hookDir)
	}

	err = hooks.UninstallLinkRunWrappers(gitx)
	log.AssertNoErrorPanicF(err, "Could not uninstall run-wrapper link in '%s'.", gitDir)

	// Always unregister repo.
	unregisterRepo(log, gitDir)

	if cleanArtefacts {
		cleanArtefactsInRepo(log, gitDir)
		cleanGitConfigInRepo(log, gitDir)
	}

	if nLfsCount != 0 {
		log.InfoF("Githooks uninstalled from '%s'.\nLFS hooks have been reinstalled.", gitDir)
	} else {
		log.InfoF("Githooks uninstalled from '%s'.", gitDir)
	}

	return true
}
