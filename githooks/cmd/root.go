package cmd

import (
	"os"

	"github.com/gabyx/githooks/githooks/build"
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	inst "github.com/gabyx/githooks/githooks/cmd/common/install"
	"github.com/gabyx/githooks/githooks/cmd/config"
	"github.com/gabyx/githooks/githooks/cmd/disable"
	"github.com/gabyx/githooks/githooks/cmd/ignore"
	"github.com/gabyx/githooks/githooks/cmd/install"
	"github.com/gabyx/githooks/githooks/cmd/installer"
	"github.com/gabyx/githooks/githooks/cmd/list"
	"github.com/gabyx/githooks/githooks/cmd/readme"
	"github.com/gabyx/githooks/githooks/cmd/shared"
	"github.com/gabyx/githooks/githooks/cmd/trust"
	"github.com/gabyx/githooks/githooks/cmd/uninstaller"
	"github.com/gabyx/githooks/githooks/cmd/update"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	"github.com/gabyx/githooks/githooks/prompt"

	"github.com/spf13/cobra"
)

// NewSettings creates common settings to use for all commands.
func NewSettings(log cm.ILogContext, logStats cm.ILogStats) ccm.CmdContext {

	var promptCtx prompt.IContext
	var err error

	cwd, err := os.Getwd()
	log.AssertNoErrorPanic(err, "Could not get current working directory.")

	installDir := inst.LoadInstallDir(log)

	promptCtx, err = prompt.CreateContext(log, false, false)
	log.AssertNoErrorF(err, "Prompt setup failed -> using fallback.")

	return ccm.CmdContext{
		Cwd:        cwd,
		GitX:       git.CtxC(cwd),
		InstallDir: installDir,
		CloneDir:   hooks.GetReleaseCloneDir(installDir),
		PromptCtx:  promptCtx,
		Log:        log,
		LogStats:   logStats}
}

func addSubCommands(cmd *cobra.Command, ctx *ccm.CmdContext) {
	cmd.AddCommand(config.NewCmd(ctx))
	cmd.AddCommand(disable.NewCmd(ctx))
	cmd.AddCommand(ignore.NewCmd(ctx))
	cmd.AddCommand(install.NewCmd(ctx)...)
	cmd.AddCommand(list.NewCmd(ctx))
	cmd.AddCommand(readme.NewCmd(ctx))
	cmd.AddCommand(shared.NewCmd(ctx))
	cmd.AddCommand(trust.NewCmd(ctx))
	cmd.AddCommand(update.NewCmd(ctx))

	cmd.AddCommand(installer.NewCmd(ctx))
	cmd.AddCommand(uninstaller.NewCmd(ctx))

}

// MakeGithooksCtl returns the root command of the Githooks CLI executable.
func MakeGithooksCtl(ctx *ccm.CmdContext) (rootCmd *cobra.Command) {

	fmt := ctx.Log.GetInfoFormatter(false)
	title := fmt("Githooks CLI [version: '%s']", build.BuildVersion)
	firstPrefix := " ▶ "
	ccm.InitTemplates(title, firstPrefix, ctx.Log.GetIndent())

	rootCmd = &cobra.Command{
		Use:   "git hooks", // Contains a en-space (utf-8: U+2002) to make it work...
		Short: "Githooks CLI application",
		Long:  "See further information at https://github.com/gabyx/githooks/blob/main/README.md"}

	ccm.ModifyTemplate(rootCmd, ctx.Log.GetIndent())

	rootCmd.SetOut(cm.ToInfoWriter(ctx.Log))
	rootCmd.SetErr(cm.ToErrorWriter(ctx.Log))
	rootCmd.Version = build.BuildVersion

	addSubCommands(rootCmd, ctx)

	ccm.SetCommandDefaults(ctx.Log, rootCmd)
	cobra.OnInitialize(initArgs)

	return rootCmd
}

func initArgs() {
	// Initialize from config , ENV -> viper
	// not yet needed...
}

// Run executes the main CLI function.
func Run(log cm.ILogContext, logStats cm.ILogStats) {

	ctx := NewSettings(log, logStats)
	cmd := MakeGithooksCtl(&ctx)

	err := cmd.Execute()
	if err != nil {
		_ = cmd.Help()
	}

	ctx.Log.AssertNoErrorPanic(err, "Command failed.")
}
