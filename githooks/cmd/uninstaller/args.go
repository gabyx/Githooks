package uninstaller

// Arguments repesents all CLI arguments for the uninstaller.
type Arguments struct {
	Config string

	InternalPostDispatch bool

	NonInteractive bool

	FullUninstallFromRepos bool

	UseStdin bool
}
