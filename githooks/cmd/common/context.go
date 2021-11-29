package ccm

import (
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/prompt"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// CmdContext is the context for the CLI.
type CmdContext struct {
	Cwd  string       // The current working directory.
	GitX *git.Context // The git context in the current working directory.

	InstallDir string // The install directory.
	CloneDir   string // The release clone dir inside the install dir.

	Log      cm.ILogContext // The log context.
	LogStats cm.ILogStats   // The statistics of the log context.

	PromptCtx prompt.IContext // The general prompt context (will be different for install/uninstall).

	WrapPanicExitCode func() // Wraps the panic exit code to 111 instead of 1.

	CleanupX *cm.InterruptContext // Crucial tasks to perform when singal is received.
}

// CmdExit is generic exit error with exit code.
type CmdExit struct {
	ExitCode int // The exit code.
}

// Error returns the error string.
func (e CmdExit) Error() string {
	return strs.Fmt("Exit code: '%v'.", e.ExitCode)
}

// NewCmdExit creates a new command error with exit code.
// The error is logged directly.
func (c *CmdContext) NewCmdExit(ec int, format string, args ...interface{}) error {
	c.Log.ErrorF(format, args...)

	return CmdExit{ExitCode: ec}
}
