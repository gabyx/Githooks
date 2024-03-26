package main

import (
	"os"
	"os/signal"

	"github.com/gabyx/githooks/githooks/apps/dialog/cmd"
	dcm "github.com/gabyx/githooks/githooks/apps/dialog/cmd/common"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/hooks"
)

func mainRun() (exitCode int) {

	// Without handling the exit code
	// would match with SIGINT on Windows, which does not have signals
	// and would call exit(SIGINT), so handle it explicitly.
	// Also on Unix, if SIGINT is received -> return 1 := cancel too.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
	}()

	go func() {
		<-c
		os.Exit(1) // Return 1 := canceled always...
	}()
	// ===============================================================

	// We only log to stderr, because we need stdout for the output.
	log, err := cm.CreateLogContext(true, false)
	cm.AssertOrPanic(err == nil, "Could not create log")

	// Handle all panics and report the error
	defer func() {
		r := recover()
		if cm.HandleCLIErrors(r, log, hooks.GetBugReportingInfo) {
			exitCode = 111
		}
	}()

	ctx := dcm.CmdContext{Log: log}
	cmd := cmd.MakeDialogCtl(&ctx)

	err = cmd.Execute()
	if err != nil {
		_ = cmd.Help()
	}
	ctx.Log.AssertNoErrorPanic(err, "Command failed.")

	return int(ctx.ExitCode)
}

func main() {
	os.Exit(mainRun())
}
