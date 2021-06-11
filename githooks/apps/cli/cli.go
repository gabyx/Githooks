package main

import (
	"os"
	"path/filepath"

	"github.com/gabyx/githooks/githooks/cmd"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/hooks"
)

func mainRun() (exitCode int) {
	cwd, err := os.Getwd()
	cm.AssertNoErrorPanic(err, "Could not get current working dir.")
	cwd = filepath.ToSlash(cwd)

	log, err := cm.CreateLogContext(false)
	cm.AssertOrPanic(err == nil, "Could not create log")

	// Handle all panics and report the error
	defer func() {
		r := recover()
		if cm.HandleCLIErrors(r, cwd, log, hooks.GetBugReportingInfo) {
			exitCode = 1
		}
	}()

	cmd.Run(log, log)

	return
}

func main() {
	os.Exit(mainRun())
}
