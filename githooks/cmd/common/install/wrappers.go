package install

import (
	"os"
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
)

// InstallIntoRepo installs run-wrappers into a repositories
// It prompts for disabling detected LFS hooks and offers to
// setup a README file.
//nolint
func InstallIntoRepo(
	log cm.ILogContext,
	gitx *git.Context,
	repoGitDir string,
	lfsHooksCache hooks.LFSHooksCache,
	nonInteractive bool,
	dryRun bool,
	skipReadme bool,
	uiSettings *UISettings) bool {

	hookDir := path.Join(repoGitDir, "hooks")
	if !cm.IsDirectory(hookDir) {
		err := os.MkdirAll(hookDir, cm.DefaultFileModeDirectory)
		log.AssertNoErrorPanic(err,
			"Could not create hook directory in '%s'.", repoGitDir)
	}
	gitxR := git.CtxC(repoGitDir)
	isBare := gitxR.IsBareRepo()

	hookNames, err := hooks.GetMaintainedHooks(gitxR, git.Traverse)
	log.AssertNoErrorF(err, "Could not get maintained hooks.")

	if dryRun {
		log.InfoF("[dry run] Hooks would have been installed into\n'%s'.",
			repoGitDir)
	} else {

		nLFSHooks, err := hooks.InstallRunWrappers(
			hookDir, hookNames,
			nil,
			GetHookDisableCallback(log, gitx, nonInteractive, uiSettings),
			lfsHooksCache,
			nil)

		log.AssertNoErrorPanicF(err, "Could not install run-wrappers into '%s'.", hookDir)

		if nLFSHooks != 0 {
			log.InfoF("Installed '%v' Githooks run-wrapper(s) and '%v' missing LFS hooks.",
				len(hookNames), nLFSHooks)
		} else {
			log.InfoF("Installed '%v' Githooks run-wrapper(s).",
				len(hookNames))
		}
	}

	// Offer to setup the intro README if running in interactive mode
	// Let's skip this in non-interactive mode or in a bare repository
	// to avoid polluting the repos with README files
	if !skipReadme && !nonInteractive && !isBare {
		setupReadme(log, repoGitDir, dryRun, uiSettings)
	}

	return !dryRun
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
	gitx := git.CtxC(gitDir)

	for _, k := range hooks.GetLocalGitConfigKeys() {

		log.AssertNoErrorF(gitx.UnsetConfig(k, git.LocalScope),
			"Could not unset Git config '%s' in '%s'.", k, gitDir)

	}
}

func unregisterRepo(log cm.ILogContext, gitDir string) {
	gitx := git.CtxC(gitDir)

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

	var err error
	var nLfsCount int

	if cm.IsDirectory(hookDir) {

		nLfsCount, err = hooks.UninstallRunWrappers(hookDir, lfsHooksCache)

		log.AssertNoErrorF(err,
			"Could not uninstall Githooks run-wrappers from\n'%s'.",
			hookDir)
	}

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
