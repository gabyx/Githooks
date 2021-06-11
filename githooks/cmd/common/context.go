package ccm

import (
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/prompt"
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
}
