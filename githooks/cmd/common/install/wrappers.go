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
// If `hookNames` is given (not nil, but can be empty), run-wrappers
// are installed.
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
	hooksDir := path.Join(repoGitDir, "hooks")
	if !cm.IsDirectory(hooksDir) {
		err := os.MkdirAll(hooksDir, cm.DefaultFileModeDirectory)
		log.AssertNoErrorPanic(err,
			"Could not create hook directory in '%s'.", repoGitDir)
	}
	gitx := git.NewCtxAt(repoGitDir)
	isBare := gitx.IsBareRepo()
	var err error

	// Check if this repository is setup to install only run-wrappers instead of using
	// a link `core.hooksPath`.
	// We switch to run-wrappers if we install a set of maintained hooks.
	// or repository settings have maintained hooks set.
	installRunWrappers, _ := cm.IsPathExisting(path.Join(hooksDir, hooks.RunWrapperMarkerFileName))
	log.DebugF("Marker file for run-wrappers detected: '%v'.", installRunWrappers)
	installRunWrappers = installRunWrappers || hookNames != nil

	var isSet bool
	if hookNames == nil {
		// Will default to all hooks if unset or wrong.
		hookNames, _, isSet, err = hooks.GetMaintainedHooks(gitx, git.LocalScope)
		log.AssertNoErrorF(err, "Could not get maintained hooks.")

		// If maintained hooks are set we install run-wrappers.
		installRunWrappers = installRunWrappers || isSet
		log.DebugF("Detected customized maintained hooks: '%v'.", isSet)
	}

	log.DebugF("Install run-wrappers: '%v'.", installRunWrappers)

	if dryRun {
		log.InfoF("[dry run] Hooks would have been installed into\n'%s'.",
			repoGitDir)

		return false
	}

	if installRunWrappers {
		gcp, gcpSet := gitx.LookupConfig(git.GitCKCoreHooksPath, git.GlobalScope)
		if gcpSet {
			log.WarnF("Global Git config '%s=%s' is set\n"+
				"which circumvents Githooks run-wrappers.\n"+
				"Not going to install run-wrappers in '%s'.\n"+
				"Did you install Githooks in 'centralized' mode?", git.GitCKCoreHooksPath, gcp, hooksDir)

			return false
		}

		lcp, lcpSet := gitx.LookupConfig(git.GitCKCoreHooksPath, git.LocalScope)
		pathToUse := gitx.GetConfig(hooks.GitCKPathForUseCoreHooksPath, git.GlobalScope)
		if lcpSet && pathToUse != lcp {
			log.WarnF("Local Git config '%s=%s' is set\n"+
				"and not maintained by Githooks ('%s').\n"+
				"This circumvents Githooks run-wrappers.\n"+
				"Not going to install run-wrappers in '%s'.", git.GitCKCoreHooksPath, lcp, pathToUse, hooksDir)

			return false
		} else if lcpSet {
			// We can safely delete this local config an then install run-wrappers.
			e := gitx.UnsetConfig(git.GitCKCoreHooksPath, git.LocalScope)
			log.AssertNoErrorPanicF(e, "Could not uset local Git config '%s'.", git.GitCKCoreHooksPath)
		}

		if isBare {
			// Filter out all non-relevant hooks for bare repositories.
			hookNames = strs.Filter(hookNames, func(s string) bool { return strs.Includes(hooks.ManagedServerHookNames, s) })
			// LFS hooks also do not need to be reinstalled
			lfsHooksCache = nil
		}

		nLFSHooks, e := hooks.InstallRunWrappers(
			hooksDir, hookNames,
			nil,
			GetHookDisableCallback(log, gitx, nonInteractive, uiSettings),
			lfsHooksCache,
			nil)

		log.AssertNoErrorPanicF(e, "Could not install run-wrappers into '%s'.", hooksDir)

		if nLFSHooks != 0 {
			log.InfoF("Installed '%v' Githooks run-wrapper(s) and '%v' missing LFS hooks into '%s'.",
				len(hookNames), nLFSHooks, hooksDir)
		} else {
			log.InfoF("Installed '%v' Githooks run-wrapper(s) into '%s'",
				len(hookNames), hooksDir)
		}
	} else {
		e := hooks.InstallRunWrappersLink(log, gitx, hooksDir, lfsHooksCache)
		log.AssertNoErrorPanicF(e, "Could not install run-wrapper link into '%s'.", repoGitDir)

		log.InfoF("Installed Githooks run-wrapper link ('%s') into '%s'",
			git.GitCKCoreHooksPath, hooksDir)
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

func cleanGitConfigInRepo(log cm.ILogContext, gitDir string, minimal bool) {
	gitx := git.NewCtxAt(gitDir)

	var names []string
	if !minimal {
		names = hooks.GetLocalGitConfigKeys()
	} else {
		names = hooks.GetLocalGitConfigKeysNonMinUninstall()
	}

	for _, k := range names {
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
	fullUninstall bool) bool {
	hooksDir := path.Join(gitDir, "hooks")
	err := os.MkdirAll(hooksDir, cm.DefaultFileModeDirectory)
	log.AssertNoErrorPanicF(err, "Could not create directory '%s'.", hooksDir)

	gitx := git.NewCtxAt(gitDir)

	var nLfsCount int

	// We always uninstall run-wrappers if any are existing.
	// Also reinstalls LFS hooks.
	// No need to check the marker file `RunWrapperMarkerFileName`.
	nLfsCount, err = hooks.UninstallRunWrappers(hooksDir, lfsHooksCache)
	log.InfoF("Githooks has reinstalled '%v' LFS hooks into '%s'.", nLfsCount, hooksDir)

	log.AssertNoErrorF(err,
		"Could not uninstall Githooks run-wrappers from\n'%s'.",
		hooksDir)

	err = hooks.UninstallRunWrappersLink(gitx)
	log.AssertNoErrorPanicF(err, "Could not uninstall run-wrapper link in '%s'.", gitDir)

	// Always unregister repo.
	unregisterRepo(log, gitDir)

	if fullUninstall {
		cleanArtefactsInRepo(log, gitDir)
		cleanGitConfigInRepo(log, gitDir, false)
	} else {
		cleanGitConfigInRepo(log, gitDir, true)
	}

	if nLfsCount != 0 {
		log.InfoF("Githooks uninstalled from '%s'.\nLFS hooks have been reinstalled.", gitDir)
	} else {
		log.InfoF("Githooks uninstalled from '%s'.", gitDir)
	}

	return true
}
