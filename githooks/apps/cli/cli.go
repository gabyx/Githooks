package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gabyx/githooks/githooks/cmd"
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/hooks"
)

func installSignalHandling() *cm.InterruptContext {
	var cleanUpX cm.InterruptContext

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)

	go func() {
		<-c
		cleanUpX.RunHandlers()
		os.Exit(1) // Return 1 := canceled always...
	}()

	return &cleanUpX
}

func mainRun(cleanUpX *cm.InterruptContext) (exitCode int) {

	log, err := cm.CreateLogContext(false, false)
	cm.AssertOrPanic(err == nil, "Could not create log")

	exitCode = 1
	panicExitCode := 1
	wrapPanicExitCode := func() {
		panicExitCode = 111
	}

	// Handle all panics and report the error
	defer func() {
		r := recover()
		if cm.HandleCLIErrors(r, log, hooks.GetBugReportingInfo) {
			exitCode = panicExitCode
		}
	}()

	err = cmd.Run(log, log, wrapPanicExitCode, cleanUpX)

	// Overwrite the exit code if its a command exit error.
	if v, ok := err.(ccm.CmdExit); ok {
		exitCode = v.ExitCode
	} else if err == nil {
		exitCode = 0
	}

	return
}

func main() {
	cleanUpX := installSignalHandling()
	exitCode := mainRun(cleanUpX)
	cleanUpX.RunHandlers()
	os.Exit(exitCode)
}
