package hooks

// Git config keys for globals config.
const (
	GitCKInstallDir = "githooks.installDir"
	GitCKRunner     = "githooks.runner"
	GitCKDialog     = "githooks.dialog"

	GitCKDisable = "githooks.disable"

	GitCKAutoUpdateEnabled        = "githooks.autoUpdateEnabled"
	GitCKAutoUpdateCheckTimestamp = "githooks.autoUpdateCheckTimestamp"
	GitCKAutoUpdateUsePrerelease  = "githooks.autoUpdateUsePrerelease"

	GitCKBugReportInfo = "githooks.bugReportInfo"

	GitCKCloneBranch     = "githooks.cloneBranch"
	GitCKCloneURL        = "githooks.cloneUrl"
	GitCKBuildFromSource = "githooks.buildFromSource"
	GitCKGoExecutable    = "githooks.goExecutable"

	GitCKDeleteDetectedLFSHooksAnswer = "githooks.deleteDetectedLFSHooks"

	GitCKUseManual         = "githooks.useManual"
	GitCKManualTemplateDir = "githooks.manualTemplateDir"

	GitCKUseCoreHooksPath        = "githooks.useCoreHooksPath"
	GitCKPathForUseCoreHooksPath = "githooks.pathForUseCoreHooksPath"

	GitCKPreviousSearchDir = "githooks.previousSearchDir"
	GitCKNumThreads        = "githooks.numThreads"

	GitCKAliasHooks = "alias.hooks"

	GitCKBuildImagesOnSharedUpdate = "githooks.buildImagesOnSharedUpdate"
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
)

// GetGlobalGitConfigKeys gets all global git config keys relevant for Githooks.
func GetGlobalGitConfigKeys() []string {
	return []string{
		GitCKInstallDir,
		GitCKRunner,
		GitCKDialog,

		GitCKDisable,

		GitCKMaintainedHooks,
		GitCKPreviousSearchDir,

		GitCKAutoUpdateEnabled,
		GitCKAutoUpdateCheckTimestamp,
		GitCKAutoUpdateUsePrerelease,

		GitCKBugReportInfo,

		GitCKCloneBranch,
		GitCKCloneURL,
		GitCKGoExecutable,
		GitCKBuildFromSource,

		GitCKDeleteDetectedLFSHooksAnswer,

		GitCKUseManual,
		GitCKManualTemplateDir,

		GitCKUseCoreHooksPath,
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
		GitCKContainerizedHooksEnabled,
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
	}
}

// var filterRegex = regexp.MustCompile(`^(githooks\.|alias.hooks|core.hook|init.template)`)

// FilterGitConfigCache filters  for filtering the Git config cache.
func FilterGitConfigCache(key string) bool {
	return true
	// Cannot filter, because `hooks.runner` needs all variables due to replacements.`
}
