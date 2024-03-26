package installer

import (
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// InstallSet is a type wrapper for installed repo data.
type InstallSet = strs.StringSet

// Settings are the settings for the installer.
type Settings struct {
	GitX            *git.Context        // The git command context.
	InstallDir      string              // The install directory.
	InstallDirRaw   string              // The raw install directory with env. variables inside.
	CloneDir        string              // The release clone dir inside the install dir.
	TempDir         string              // The temporary directory inside the install dir.
	LFSHooksCache   hooks.LFSHooksCache // LFS hooks cache if 'git lfs' is available.
	HookTemplateDir string              // The chosen hook template directory.

	// Registered Repos loaded from the install dir.
	// New registered repos will be added here.
	RegisteredGitDirs hooks.RegisterRepos

	// All repositories Git directories where Githooks run-wrappers have been installed.
	InstalledGitDirs InstallSet
}
