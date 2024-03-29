package main

import (
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// HookSettings defines hooks related settings for this run.
type HookSettings struct {
	Args               []string       // Rest arguments.
	ExecX              cm.ExecContext // Execution context for executables (working dir is this repository).
	GitX               *git.Context   // Git context to execute commands (working dir is this repository).
	RepositoryDir      string         // Repository directory (bare, non-bare).
	RepositoryHooksDir string         // Directory with hooks for this repository.
	GitDirWorktree     string         // Git directory. (for worktrees this points to the worktree Git dir).
	InstallDir         string         // Install directory.

	HookPath      string // Absolute path of the hook executing this runner.
	HookName      string // Name of the hook.
	HookDir       string // Directory of the hook.
	HookNamespace string // Namespace of this repositorie's Githooks.

	IsRepoTrusted              bool // If the repository is a trusted repository.
	SkipNonExistingSharedHooks bool // If Githooks should skip non-existing shared hooks.
	SkipUntrustedHooks         bool // If Githooks should skip active untrusted hooks.
	NonInteractive             bool // If all non-fatal prompts should be default answered.
	ContainerizedHooksEnabled  bool // If all hooks should run containerized (if they are setup for it).
	Disabled                   bool // If Githooks has been disabled.

	StagedFilesFile string // The temporary file where all staged files are written to.
}

func (s HookSettings) toString() string {
	return strs.Fmt(
		" • Args: '%q'\n"+
			" • Repo Path: '%s'\n"+
			" • Repo Hooks: '%s'\n"+
			" • Git Dir Worktree: '%s'\n"+
			" • Install Dir: '%s'\n"+
			" • Hook Path: '%s'\n"+
			" • Hook Name: '%s'\n"+
			" • Trusted: '%v'\n"+
			" • ContainerizedEnabled: '%v'",
		s.Args, s.RepositoryDir,
		s.RepositoryHooksDir, s.GitDirWorktree,
		s.InstallDir, s.HookPath, s.HookName, s.IsRepoTrusted,
		s.ContainerizedHooksEnabled)
}
