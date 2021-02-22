package main

import (
	dcm "gabyx/githooks/apps/dialog/cmd/common"
	"gabyx/githooks/apps/dialog/cmd/entry"
	"gabyx/githooks/apps/dialog/cmd/file"
	"gabyx/githooks/apps/dialog/cmd/message"
	"gabyx/githooks/apps/dialog/cmd/notify"
	"gabyx/githooks/apps/dialog/cmd/options"
	"gabyx/githooks/build"
	ccm "gabyx/githooks/cmd/common"
	cm "gabyx/githooks/common"
	"gabyx/githooks/hooks"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/spf13/cobra"
)

func addSubCommands(cmd *cobra.Command, ctx *dcm.CmdContext) {
	cmd.AddCommand(options.NewCmd(ctx))
	cmd.AddCommand(message.NewCmd(ctx))
	cmd.AddCommand(entry.NewCmd(ctx))
	cmd.AddCommand(notify.NewCmd(ctx))
	cmd.AddCommand(file.NewCmd(ctx)...)
}

// makeDialogCtl returns the root command of the Githooks dialog executable.
func makeDialogCtl(ctx *dcm.CmdContext) (rootCmd *cobra.Command) {

	fmt := ctx.Log.GetInfoFormatter(false)
	title := fmt("Githooks Dialog CLI [version: %s]", build.BuildVersion)
	firstPrefix := " ▶ "
	ccm.InitTemplates(title, firstPrefix, ctx.Log.GetIndent())

	rootCmd = &cobra.Command{
		Use:   "dialog", // Contains a en-space (utf-8: U+2002) to make it work...
		Short: "Githooks dialog application",
		Long:  "See further information at https://github.com/gabyx/githooks/blob/main/README.md"}

	ccm.ModifyTemplate(rootCmd, ctx.Log.GetIndent())

	rootCmd.SetOut(cm.ToInfoWriter(ctx.Log))
	rootCmd.SetErr(cm.ToErrorWriter(ctx.Log))
	rootCmd.Version = build.BuildVersion

	addSubCommands(rootCmd, ctx)

	ccm.SetCommandDefaults(ctx.Log, rootCmd)

	return rootCmd
}

func mainRun() (exitCode int) {

	// Without handling the exit code
	// would match with SIGINT.
	// At least on Windows, which does not seem to add it to 128.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
	}()

	go func() {
		<-c
		os.Exit(1) // Return canceled always...
	}()

	cwd, err := os.Getwd()
	cm.AssertNoErrorPanic(err, "Could not get current working dir.")
	cwd = filepath.ToSlash(cwd)

	// We only log to stderr, because we need stdout for the output.
	log, err := cm.CreateLogContext(true)
	cm.AssertOrPanic(err == nil, "Could not create log")

	// Handle all panics and report the error
	defer func() {
		r := recover()
		if cm.HandleCLIErrors(r, cwd, log, hooks.GetBugReportingInfo) {
			exitCode = 111
		}
	}()

	ctx := dcm.CmdContext{Log: log}
	cmd := makeDialogCtl(&ctx)

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
