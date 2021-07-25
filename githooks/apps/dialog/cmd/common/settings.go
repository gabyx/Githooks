package common

import (
	cm "github.com/gabyx/githooks/githooks/common"
)

// CmdContext is the command context for the dialog executable.
type CmdContext struct {

	// Exit code of the dialog app.
	ExitCode ExitCode

	// Report as JSON
	ReportAsJSON bool

	// Log context.
	Log cm.ILogContext
}

// ExitCode is the exit code type of the executable.
type ExitCode = uint

const (
	// ExitCodeCanceled is the exit code if the user cancled or closed the dialog.
	ExitCodeCanceled = 1
)
