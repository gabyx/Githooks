package uninstaller

import (
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	"github.com/gabyx/githooks/githooks/prompt"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// UninstallSet is a typedef for tracking uninstalled repos.
type UninstallSet = strs.StringSet

// Settings are the settings for the installer.
type Settings struct {
	Gitx       *git.Context // The git command context.
	InstallDir string       // The install directory.
	CloneDir   string       // The release clone dir inside the install dir.
	TempDir    string       // The temporary directory inside the install dir.

	PromptCtx prompt.IContext // The prompt context for UI prompts.

	HookTemplateDir string // The chosen hook template directory.

	// Registered Repos loaded from the install dir.
	RegisteredGitDirs hooks.RegisterRepos

	// All repositories Git directories where Githooks run-wrappers have been installed.
	// Bool indicates if it is already registered.
	UninstalledGitDirs UninstallSet

	// LFS hooks cache if 'git-lfs' is installed.
	LFSHooksCache hooks.LFSHooksCache
}
