package hooks

import (
	cm "gabyx/githooks/common"
	"gabyx/githooks/git"
	strs "gabyx/githooks/strings"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

// HooksDirName denotes the directory name used for repository specific hooks.
const HooksDirName = ".githooks"

// GithooksWebpage is the main Githooks webpage.
const GithooksWebpage = "https://github.com/gabyx/githooks"

// DefaultBugReportingURL is the default url to report errors.
const DefaultBugReportingURL = "https://github.com/gabyx/githooks/issues"

// ManagedHookNames are hook names managed by Githooks for normal repositories.
var ManagedHookNames = []string{
	"applypatch-msg",
	"pre-applypatch",
	"post-applypatch",
	"pre-commit",
	"pre-merge-commit",
	"prepare-commit-msg",
	"commit-msg",
	"post-commit",
	"pre-rebase",
	"post-checkout",
	"post-merge",
	"pre-push",
	"pre-receive",
	"update",
	"post-receive",
	"post-update",
	"reference-transaction",
	"push-to-checkout",
	"pre-auto-gc",
	"post-rewrite",
	"sendemail-validate",
	"post-index-change"}

// ManagedServerHookNames are hook names managed by Githooks for bare repositories.
var ManagedServerHookNames = []string{
	"pre-push",
	"pre-receive",
	"update",
	"post-receive",
	"post-update",
	"reference-transaction",
	"push-to-checkout",
	"pre-auto-gc"}

// LFSHookNames are the hook names of all Large File System (LFS) hooks.
var LFSHookNames = [4]string{
	"post-checkout",
	"post-commit",
	"post-merge",
	"pre-push"}

// StagedFilesHookNames are the hook names on which staged files are exported.
var StagedFilesHookNames = [3]string{"pre-commit", "prepare-commit-msg", "commit-msg"}

// EnvVariableStagedFiles is the environment variable which holds the staged files.
const EnvVariableStagedFiles = "STAGED_FILES"

// GetBugReportingInfo gets the default bug reporting url. Argument 'repoPath' can be empty.
func GetBugReportingInfo(repoPath string) (info string, err error) {
	var exists bool

	// Set default if needed
	defer func() {
		if strs.IsEmpty(info) {
			info = strs.Fmt("-> Report this bug to: '%s'", DefaultBugReportingURL)
		}
	}()

	// Check in the repo if possible
	if !strs.IsEmpty(repoPath) {
		file := path.Join(GetGithooksDir(repoPath), ".bug-report")
		exists, err = cm.IsPathExisting(file)

		if exists {
			data, err := ioutil.ReadFile(file)
			if err == nil {
				info = string(data)
			}
		}
	}

	// Check global Git config
	info = git.Ctx().GetConfig(GitCKBugReportInfo, git.GlobalScope)

	return
}

// GetGithooksDir gets the hooks directory for Githooks inside a repository (bare, non-bare).
func GetGithooksDir(repoDir string) string {
	return path.Join(repoDir, HooksDirName)
}

// GetSharedGithooksDir gets the hooks directory for Githooks inside a shared repository.
func GetSharedGithooksDir(repoDir string) string {
	return path.Join(repoDir, "githooks")
}

// GetInstallDir returns the Githooks install directory.
func GetInstallDir() string {
	return filepath.ToSlash(git.Ctx().GetConfig(GitCKInstallDir, git.GlobalScope))
}

// SetInstallDir sets the global Githooks install directory.
func SetInstallDir(path string) error {
	return git.Ctx().SetConfig(GitCKInstallDir, path, git.GlobalScope)
}

// GetBinaryDir returns the Githooks binary directory inside the install directory.
func GetBinaryDir(installDir string) string {
	return path.Join(installDir, "bin")
}

// GetTemporaryDir returns the Githooks temporary directory inside the install directory.
func GetTemporaryDir(installDir string) string {
	cm.DebugAssert(strs.IsNotEmpty(installDir))

	return path.Join(installDir, "tmp")
}

// AssertTemporaryDir returns the Githooks temporary directory inside the install directory.
func AssertTemporaryDir(installDir string) (tempDir string, err error) {
	tempDir = GetTemporaryDir(installDir)
	err = os.MkdirAll(GetTemporaryDir(installDir), cm.DefaultFileModeDirectory)

	return
}

func removeTemporaryDir(installDir string) (err error) {
	cm.DebugAssert(strs.IsNotEmpty(installDir))
	tempDir := GetTemporaryDir(installDir)

	if err = os.RemoveAll(tempDir); err != nil {
		return
	}

	return
}

// CleanTemporaryDir returns the Githooks temporary directory inside the install directory.
func CleanTemporaryDir(installDir string) (string, error) {
	if err := removeTemporaryDir(installDir); err != nil {
		return "", err
	}

	return AssertTemporaryDir(installDir)
}

// GetRunnerExecutable gets the installed Githooks runner executable.
func GetRunnerExecutable(installDir string) (p string) {
	p = path.Join(GetBinaryDir(installDir), "runner")
	if runtime.GOOS == cm.WindowsOsName {
		p += cm.WindowsExecutableSuffix
	}

	return
}

// SetRunnerExecutableAlias sets the global Githooks runner executable.
func SetRunnerExecutableAlias(path string) error {
	if !cm.IsFile(path) {
		return cm.ErrorF("Runner executable '%s' does not exist.", path)
	}

	return git.Ctx().SetConfig(GitCKRunner, path, git.GlobalScope)
}

// GetDialogExecutable gets the installed Githooks dialog executable.
func GetDialogExecutable(installDir string) (p string) {
	p = path.Join(GetBinaryDir(installDir), "dialog")
	if runtime.GOOS == cm.WindowsOsName {
		p += cm.WindowsExecutableSuffix
	}

	return
}

// SetDialogExecutableConfig sets the global Githooks dialog executable.
func SetDialogExecutableConfig(path string) error {
	if !cm.IsFile(path) {
		return cm.ErrorF("Dialog executable '%s' does not exist.", path)
	}

	return git.Ctx().SetConfig(GitCKDialog, path, git.GlobalScope)
}

// SetCLIExecutableAlias sets the global Githooks runner executable.
func SetCLIExecutableAlias(path string) error {
	if !cm.IsFile(path) {
		return cm.ErrorF("CLI executable '%s' does not exist.", path)
	}

	return git.Ctx().SetConfig(GitCKAliasHooks, strs.Fmt("!\"%s\"", path), git.GlobalScope)
}

// GetReleaseCloneDir get the release clone directory inside the install dir.
func GetReleaseCloneDir(installDir string) string {
	cm.DebugAssert(strs.IsNotEmpty(installDir), "Empty install dir.")

	return path.Join(installDir, "release")
}

// GetLFSRequiredFile gets the LFS-Required file inside the repository.
func GetLFSRequiredFile(repoDir string) string {
	return path.Join(GetGithooksDir(repoDir), ".lfs-required")
}

// IsGithooksDisabled checks if Githooks is disabled in
// any config starting from the working dir given by the git context or
// optional also by the env. variable `GITHOOKS_DISABLE`.
func IsGithooksDisabled(gitx *git.Context, checkEnv bool) bool {

	if checkEnv {
		env := os.Getenv("GITHOOKS_DISABLE")
		if env != "" &&
			env != "0" &&
			env != "false" && env != "off" {
			return true
		}
	}

	disabled := gitx.GetConfig(GitCKDisable, git.Traverse)

	return disabled == "true" || // nolint: goconst
		disabled == "y" || // Legacy
		disabled == "Y" // Legacy
}
