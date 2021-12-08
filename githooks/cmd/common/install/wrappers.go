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

	isBare := git.CtxC(repoGitDir).IsBareRepo()

	var hookNames []string
	if isBare {
		hookNames = hooks.ManagedServerHookNames
	} else {
		hookNames = hooks.ManagedHookNames
	}

	if dryRun {
		log.InfoF("[dry run] Hooks would have been installed into\n'%s'.",
			repoGitDir)
	} else {

		err := hooks.InstallRunWrappers(
			hookDir, hookNames,
			nil,
			GetHookDisableCallback(log, gitx, nonInteractive, uiSettings),
			nil)

		log.AssertNoErrorPanicF(err, "Could not install run-wrappers into '%s'.", hookDir)
		log.InfoF("Githooks run-wrappers installed into '%s'.", hookDir)
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
	lfsAvailable bool,
	cleanArtefacts bool) bool {

	hookDir := path.Join(gitDir, "hooks")

	if cm.IsDirectory(hookDir) {

		err := hooks.UninstallRunWrappers(hookDir, hooks.ManagedHookNames)

		log.AssertNoErrorF(err,
			"Could not uninstall Githooks run-wrappers from\n'%s'.",
			hookDir)

		if err == nil {

			if lfsAvailable {
				err = hooks.InstallLFSHooks(gitDir)

				log.AssertNoErrorF(err,
					"Could not reinstall Git LFS hooks in\n"+
						"'%[1]s'.\n"+
						"Please try manually by invoking:\n"+
						"  $ git -C '%[1]s' lfs install", gitDir)

			}
		}
	}

	// Always unregister repo.
	unregisterRepo(log, gitDir)

	if cleanArtefacts {
		cleanArtefactsInRepo(log, gitDir)
		cleanGitConfigInRepo(log, gitDir)
	}

	log.InfoF("Githooks uninstalled from '%s'.", gitDir)

	return true
}
