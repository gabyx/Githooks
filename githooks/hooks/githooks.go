package hooks

import (
	"os"
	"path"
	"path/filepath"
	"runtime"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"
)

const RunnerName = "githooks-runner"
const CLIName = "githooks-cli"
const DialogName = "githooks-dialog"

// HooksDirName denotes the directory name used for repository specific hooks.
const HooksDirName = ".githooks"
const HooksDirNameShared = "githooks"

// GithooksWebpage is the main Githooks webpage.
const GithooksWebpage = "https://github.com/gabyx/githooks"

// DefaultBugReportingURL is the default url to report errors.
const DefaultBugReportingURL = "https://github.com/gabyx/githooks/issues"

// All Git hook names.
var AllHookNames = []string{
	"applypatch-msg",
	"commit-msg",
	"fsmonitor-watchman",
	"p4-changelist",
	"p4-post-changelist",
	"p4-prepare-changelist",
	"p4-pre-submit",
	"post-applypatch",
	"post-checkout",
	"post-commit",
	"post-index-change",
	"post-merge",
	"post-receive",
	"post-rewrite",
	"post-update",
	"pre-applypatch",
	"pre-auto-gc",
	"pre-commit",
	"pre-merge-commit",
	"prepare-commit-msg",
	"pre-push",
	"pre-rebase",
	"pre-receive",
	"proc-receive",
	"push-to-checkout",
	"reference-transaction",
	"sendemail-validate",
	"update",
}

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

// EnvVariableOs is the environment variable which holds runtime operating system name.
const EnvVariableOs = "GITHOOKS_OS"

// EnvVariableArch is the environment variable which holds runtime architecture name.
const EnvVariableArch = "GITHOOKS_ARCH"

// EnvVariableStagedFiles is the environment variable which holds the staged files.
const EnvVariableStagedFiles = "STAGED_FILES"

// EnvVariableStagedFilesFile is the environment variable pointing to a
// file which holds the staged files relative
// to the repository where Githooks runs.
const EnvVariableStagedFilesFile = "STAGED_FILES_FILE"

// GetGithooksEnvVariables gets all Githooks env variables.
// `EnvVariableStagedFilesFile` variable's value is modified optionaly.
func GetGithooksEnvVariables(newStagedFilesFile string) []string {
	var env []string

	names := []string{EnvVariableOs, EnvVariableArch}
	for i := range names {
		env = append(env, strs.Fmt("%s=%s", names[i], os.Getenv(names[i])))
	}

	names = []string{EnvVariableStagedFiles, EnvVariableStagedFilesFile}
	for i := range names {
		if val, exists := os.LookupEnv(names[i]); exists {

			// Modify the file name.
			if names[i] == EnvVariableStagedFilesFile &&
				strs.IsNotEmpty(newStagedFilesFile) {
				val = newStagedFilesFile
			}

			env = append(env, strs.Fmt("%s=%s", names[i], val))
		}
	}

	return env
}

// GetBugReportingInfo gets the default bug reporting url. Argument 'repoPath' can be empty.
func GetBugReportingInfo() (info string) {

	// Set default if needed
	defer func() {
		if strs.IsEmpty(info) {
			info = strs.Fmt("-> Report this bug to: '%s'", DefaultBugReportingURL)
		}
	}()

	// Check global Git config
	info = git.NewCtx().GetConfig(GitCKBugReportInfo, git.GlobalScope)

	return
}

// GetGithooksDir gets the hooks directory for Githooks inside a repository (bare, non-bare).
func GetGithooksDir(repoDir string) string {
	return path.Join(repoDir, HooksDirName)
}

// GetSharedGithooksDir gets the hooks directory for Githooks inside a shared repository.
func GetSharedGithooksDir(repoDir string) (dir string) {

	// 1. priority has non-dot folder 'githooks'
	// 2. priority is the normal '.githooks' folder.
	// This is second, to allow internal development Githooks inside shared repos.
	// 3. Fallback to the whole repository.

	if dir = path.Join(repoDir, HooksDirNameShared); cm.IsDirectory(dir) {
		return
	} else if dir = GetGithooksDir(repoDir); cm.IsDirectory(dir) {
		return
	}

	dir = repoDir

	return
}

// GetInstallDir returns the Githooks install directory.
func GetInstallDir(gitx *git.Context) string {
	return filepath.ToSlash(gitx.GetConfig(GitCKInstallDir, git.GlobalScope))
}

// SetInstallDir sets the global Githooks install directory.
func SetInstallDir(gitx *git.Context, path string) error {
	return gitx.SetConfig(GitCKInstallDir, path, git.GlobalScope)
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
	err = os.MkdirAll(tempDir, cm.DefaultFileModeDirectory)

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

	return GetTemporaryDir(installDir), nil
}

// GetRunnerExecutable gets the installed Githooks runner executable.
func GetRunnerExecutable(installDir string) (p string) {
	p = path.Join(GetBinaryDir(installDir), RunnerName)
	if runtime.GOOS == cm.WindowsOsName {
		p += cm.WindowsExecutableSuffix
	}

	return
}

// SetRunnerExecutableAlias sets the global Githooks runner executable.
func SetRunnerExecutableAlias(path string) error {
	return git.NewCtx().SetConfig(GitCKRunner, path, git.GlobalScope)
}

// GetDialogExecutable gets the installed Githooks dialog executable.
func GetDialogExecutable(installDir string) (p string) {
	p = path.Join(GetBinaryDir(installDir), DialogName)
	if runtime.GOOS == cm.WindowsOsName {
		p += cm.WindowsExecutableSuffix
	}

	return
}

// SetDialogExecutableConfig sets the global Githooks dialog executable.
func SetDialogExecutableConfig(path string) error {
	return git.NewCtx().SetConfig(GitCKDialog, path, git.GlobalScope)
}

// SetCLIExecutableAlias sets the global Githooks runner executable.
func SetCLIExecutableAlias(path string) error {
	return git.NewCtx().SetConfig(GitCKAliasHooks, strs.Fmt("!\"%s\"", path), git.GlobalScope)
}

// GetReleaseCloneDir get the release clone directory inside the install dir.
func GetReleaseCloneDir(installDir string) string {
	cm.DebugAssert(strs.IsNotEmpty(installDir), "Empty install dir.")

	return path.Join(installDir, "release")
}

// GetLFSRequiredFile gets the LFS-Required file inside the repository
// and `true` if existing.
func GetLFSRequiredFile(repoDir string) (string, bool) {
	s := path.Join(GetGithooksDir(repoDir), ".lfs-required")

	return s, cm.IsFile(s)
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

	return disabled == git.GitCVTrue || // nolint: goconst
		disabled == "y" || // Legacy
		disabled == "Y" // Legacy
}

// IsContainerizedHooksEnabled returns if containerized hooks are enabled.
func IsContainerizedHooksEnabled(gitx *git.Context, checkEnv bool) bool {
	if checkEnv {
		env := os.Getenv("GITHOOKS_CONTAINERIZED_HOOKS_ENABLED")
		if env != "" &&
			env != "0" &&
			env != "false" && env != "off" {
			return true
		}
	}

	enabled := gitx.GetConfig(GitCKContainerizedHooksEnabled, git.Traverse)

	return enabled == git.GitCVTrue
}

// IsRunnerNonInteractive tells if the runner should run in non-interactive mode
// meaning all non-fatal prompts will be skipped with default answering
// and fatal prompts still need to be configured to pass.
func IsRunnerNonInteractive(gitx *git.Context, scope git.ConfigScope) bool {
	return gitx.GetConfig(GitCKRunnerIsNonInteractive, scope) == "true"
}

// SetRunnerNonInteractive sets the runner to non-interactive mode.
func SetRunnerNonInteractive(gitx *git.Context, enable bool, reset bool, scope git.ConfigScope) error {
	switch {
	case reset:
		return gitx.UnsetConfig(GitCKRunnerIsNonInteractive, scope)
	default:
		return gitx.SetConfig(GitCKRunnerIsNonInteractive, enable, scope)
	}
}
