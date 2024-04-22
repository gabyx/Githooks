package installer

// Arguments represents all CLI arguments for the installer.
type Arguments struct {
	Config string

	Log                  string // The log file.
	InternalTempDir      string // The temporary directory to use.
	InternalPostDispatch bool   // If the installer has already dispatched itself to the downloaded/build installer.

	InternalUpdateFromVersion string   // Build version we are updating from.
	InternalUpdateTo          string   // Commit SHA to update local branch to remote.
	InternalBinaries          []string // Binaries which need to get installed.

	DryRun         bool
	NonInteractive bool

	// Directly update to the latest possible tag on the clone branch.
	// Before `2.3.3` that was always true.
	Update bool

	SkipInstallIntoExisting bool // Skip install into existing repositories.

	MaintainedHooks []string // Maintain hooks by Githooks.

	// Use install mode with the global `core.hooksPath` for the hook run wrappers.
	Centralized bool

	InstallPrefix string // Install prefix for Githooks.
	HooksDir      string // The directory to use to install the global maintained run-wrappers.

	// If the hooks directory from `GIT_TEMPLATE_DIR` or `init.templateDir` or the Git default is taken.
	HooksDirUseTemplateDir bool

	CloneURL       string // Clone URL of the Githooks repository.
	CloneBranch    string // Clone branch for Githooks repository.
	DeployAPI      string // Deploy API to use for auto detection of deploy settings.
	DeploySettings string // Deploy settings YAML file.

	BuildFromSource bool     // If we build the update from source.
	BuildTags       []string // Go build tags.

	UsePreRelease bool // If also pre-release versions should be considered.

	GitConfigNoAbsPath bool // If certain Githooks Git config values are not using absolute paths.
	UseStdin           bool
}
