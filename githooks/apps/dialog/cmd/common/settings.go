package common

import (
	cm "gabyx/githooks/common"
)

type CmdContext struct {

	// Exit code of the dialog app.
	ExitCode ExitCode

	// Log context.
	Log cm.ILogContext
}

type ExitCode = uint

const (
	// ExitCodeCanceled is the exit code if the user cancled or closed the dialog.
	ExitCodeCanceled = 1
)
