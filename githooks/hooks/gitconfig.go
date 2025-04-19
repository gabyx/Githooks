package hooks

// Git config keys for globals config.
const (
	GitCKInstallDir = "githooks.installDir"
	GitCKRunner     = "githooks.runner"
	GitCKDialog     = "githooks.dialog"

	GitCKDisable = "githooks.disable"

	GitCKUpdateCheckEnabled       = "githooks.updateCheckEnabled"
	GitCKUpdateCheckUsePrerelease = "githooks.updateCheckUsePrerelease"

	GitCKBugReportInfo = "githooks.bugReportInfo"

	GitCKCloneBranch     = "githooks.cloneBranch"
	GitCKCloneURL        = "githooks.cloneUrl"
	GitCKBuildFromSource = "githooks.buildFromSource"
	GitCKGoExecutable    = "githooks.goExecutable"

	GitCKDeleteDetectedLFSHooksAnswer = "githooks.deleteDetectedLFSHooks"

	// Install modes.
	GitCKInstallMode = "githooks.installMode"

	GitCKPathForUseCoreHooksPath = "githooks.pathForUseCoreHooksPath"

	GitCKPreviousSearchDir = "githooks.previousSearchDir"
	GitCKNumThreads        = "githooks.numThreads"

	GitCKAliasHooks = "alias.hooks"
)

// Git config keys for local config.
const (
	GitCKRegistered = "githooks.registered"
	GitCKTrustAll   = "githooks.trustAll"
)

// Git config keys for local/global config.
const (
	GitCKMaintainedHooks = "githooks.maintainedHooks"

	GitCKShared                        = "githooks.shared"
	GitCKSharedUpdateTriggers          = "githooks.sharedHooksUpdateTriggers"
	GitCKAutoUpdateSharedHooksDisabled = "githooks.autoUpdateSharedHooksDisabled"

	GitCKSkipNonExistingSharedHooks = "githooks.skipNonExistingSharedHooks"
	GitCKSkipUntrustedHooks         = "githooks.skipUntrustedHooks"

	GitCKRunnerIsNonInteractive = "githooks.runnerIsNonInteractive"

	GitCKContainerizedHooksEnabled     = "githooks.containerizedHooksEnabled"
	GitCKContainerManager              = "githooks.containerManager"
	GitCKContainerImageUpdateAutomatic = "githooks.containerImageUpdateAutomatic"

	GitCKExportStagedFilesAsFile = "githooks.exportStagedFilesAsFile"

	GitCKChecksumsDir = "githooks.checksumsDir"
)

// GetGlobalGitConfigKeys gets all global git config keys relevant for Githooks.
func GetGlobalGitConfigKeys() []string {
	return []string{
		GitCKInstallMode,
		GitCKInstallDir,
		GitCKRunner,
		GitCKDialog,

		GitCKDisable,

		GitCKMaintainedHooks,
		GitCKPreviousSearchDir,

		GitCKUpdateCheckEnabled,
		GitCKUpdateCheckUsePrerelease,

		GitCKBugReportInfo,

		GitCKCloneBranch,
		GitCKCloneURL,
		GitCKGoExecutable,
		GitCKBuildFromSource,

		GitCKDeleteDetectedLFSHooksAnswer,

		GitCKPathForUseCoreHooksPath,

		GitCKNumThreads,

		GitCKAliasHooks,

		// Local & global.
		GitCKShared,
		GitCKSharedUpdateTriggers,
		GitCKAutoUpdateSharedHooksDisabled,

		GitCKSkipNonExistingSharedHooks,
		GitCKSkipUntrustedHooks,

		GitCKRunnerIsNonInteractive,

		GitCKContainerManager,
		GitCKExportStagedFilesAsFile,

		GitCKContainerizedHooksEnabled,

		GitCKChecksumsDir,
	}
}

// GetLocalGitConfigKeys gets all local git config keys relevant for Githooks.
func GetLocalGitConfigKeys() []string {
	return []string{
		GitCKRegistered,
		GitCKTrustAll,

		GitCKMaintainedHooks,

		GitCKShared,
		GitCKSharedUpdateTriggers,
		GitCKAutoUpdateSharedHooksDisabled,

		GitCKSkipNonExistingSharedHooks,
		GitCKSkipUntrustedHooks,

		GitCKRunnerIsNonInteractive,

		GitCKContainerManager,
		GitCKContainerizedHooksEnabled,

		GitCKExportStagedFilesAsFile,
		GitCKChecksumsDir,
	}
}

// GetLocalGitConfigKeysNonMinUninstall gets all local keys which should always be uninstalled
// in registered repos.
func GetLocalGitConfigKeysNonMinUninstall() []string {
	return []string{
		GitCKRegistered,
	}
}

// var filterRegex = regexp.MustCompile(`^(githooks\.|alias.hooks|core.hook|init.template)`)

// FilterGitConfigCache filters  for filtering the Git config cache.
func FilterGitConfigCache(key string) bool {
	return true
	// Cannot filter, because `hooks.runner` needs all variables due to replacements.`
}
