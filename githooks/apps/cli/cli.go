package main

import (
	"os"
	"os/signal"
	"path/filepath"

	"github.com/gabyx/githooks/githooks/cmd"
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/hooks"
)

func mainRun() (exitCode int) {
	cwd, err := os.Getwd()
	cm.AssertNoErrorPanic(err, "Could not get current working dir.")
	cwd = filepath.ToSlash(cwd)

	log, err := cm.CreateLogContext(false)
	cm.AssertOrPanic(err == nil, "Could not create log")

	exitCode = 1
	panicExitCode := 1
	wrapPanicExitCode := func() {
		panicExitCode = 111
	}

	// Install signal handling
	var cleanUpX cm.InterruptContext
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
	}()

	go func() {
		<-c
		cleanUpX.RunHandlers()
		os.Exit(1) // Return 1 := canceled always...
	}()

	// Handle all panics and report the error
	defer func() {
		r := recover()
		if cm.HandleCLIErrors(r, cwd, log, hooks.GetBugReportingInfo) {
			exitCode = panicExitCode
		}
	}()

	err = cmd.Run(log, log, wrapPanicExitCode, &cleanUpX)

	// Overwrite the exit code if its a command exit error.
	if v, ok := err.(ccm.CmdExit); ok {
		exitCode = v.ExitCode
	} else if err == nil {
		exitCode = 0
	}

	return
}

func main() {
	os.Exit(mainRun())
}
