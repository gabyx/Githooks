package common

import (
	cm "github.com/gabyx/githooks/githooks/common"
)

type CmdContext struct {

	// Exit code of the dialog app.
	ExitCode ExitCode

	// Report as JSON
	ReportAsJSON bool

	// Log context.
	Log cm.ILogContext
}

type ExitCode = uint

const (
	// ExitCodeCanceled is the exit code if the user cancled or closed the dialog.
	ExitCodeCanceled = 1
)
