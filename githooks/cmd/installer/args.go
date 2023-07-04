package installer

// Arguments represents all CLI arguments for the installer.
type Arguments struct {
	Config string

	Log                  string // The log file.
	InternalAutoUpdate   bool   // If the installer is run from the runner.
	InternalPostDispatch bool   // If the installer has already dispatched itself to the downloaded/build installer.

	InternalUpdateFromVersion string   // Build version we are updating from.
	InternalUpdateTo          string   // Commit SHA to update local branch to remote.
	InternalBinaries          []string // Binaries which need to get installed.

	DryRun         bool
	NonInteractive bool

	Update bool // Directly update to the latest possible tag on the clone branch.
	// Before `2.3.3` that was always true.

	SkipInstallIntoExisting bool // Skip install into existing repositories.

	MaintainedHooks []string // Maintain hooks by Githooks.

	UseTemplateDir   bool // Use install mode: `init.templateDir`
	UseCoreHooksPath bool // Use install mode: `core.hooksPath` for the template dir.
	UseManual        bool // Use install mode: manual -> no `core.hooksPath` nor `init.templateDir`

	InstallPrefix string // Install prefix for Githooks.
	TemplateDir   string // Template dir to use for the hooks.

	CloneURL       string // Clone URL of the Githooks repository.
	CloneBranch    string // Clone branch for Githooks repository.
	DeployAPI      string // Deploy API to use for auto detection of deploy settings.
	DeploySettings string // Deploy settings YAML file.

	BuildFromSource bool     // If we build the install/update from source.
	BuildTags       []string // Go build tags.

	UsePreRelease bool // If also pre-release versions should be considered.

	UseStdin bool
}
