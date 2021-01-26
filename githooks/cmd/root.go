package cmd

import (
	inst "gabyx/githooks/apps/install"
	"gabyx/githooks/build"
	ccm "gabyx/githooks/cmd/common"
	"gabyx/githooks/cmd/config"
	"gabyx/githooks/cmd/disable"
	"gabyx/githooks/cmd/ignore"
	"gabyx/githooks/cmd/install"
	"gabyx/githooks/cmd/list"
	"gabyx/githooks/cmd/readme"
	"gabyx/githooks/cmd/shared"
	"gabyx/githooks/cmd/tools"
	"gabyx/githooks/cmd/trust"
	"gabyx/githooks/cmd/update"
	cm "gabyx/githooks/common"
	"gabyx/githooks/git"
	"gabyx/githooks/hooks"
	"gabyx/githooks/prompt"
	"os"

	"github.com/spf13/cobra"
)

func NewSettings(log cm.ILogContext) ccm.CmdContext {

	var promptCtx prompt.IContext
	var err error

	cwd, err := os.Getwd()
	log.AssertNoErrorPanic(err, "Could not get current working directory.")

	promptCtx, err = prompt.CreateContext(log, &cm.ExecContext{}, nil, false, false)
	log.AssertNoErrorF(err, "Prompt setup failed -> using fallback.")

	installDir := inst.LoadInstallDir(log)

	return ccm.CmdContext{
		Cwd:        cwd,
		GitX:       git.CtxC(cwd),
		InstallDir: installDir,
		CloneDir:   hooks.GetReleaseCloneDir(installDir),
		PromptCtx:  promptCtx,
		Log:        log}
}

func addSubCommands(cmd *cobra.Command, ctx *ccm.CmdContext) {
	cmd.AddCommand(config.NewCmd(ctx))
	cmd.AddCommand(disable.NewCmd(ctx))
	cmd.AddCommand(ignore.NewCmd(ctx))
	cmd.AddCommand(install.NewCmd(ctx)...)
	cmd.AddCommand(list.NewCmd(ctx))
	cmd.AddCommand(readme.NewCmd(ctx))
	cmd.AddCommand(shared.NewCmd(ctx))
	cmd.AddCommand(tools.NewCmd(ctx))
	cmd.AddCommand(trust.NewCmd(ctx))
	cmd.AddCommand(update.NewCmd(ctx))
}

func MakeGithooksCtl(ctx *ccm.CmdContext) (rootCmd *cobra.Command) {

	fmt := ctx.Log.GetInfoFormatter(false)
	title := fmt("Githooks CLI [version: %s]", build.BuildVersion)
	firstPrefix := " ▶ "
	ccm.InitTemplates(title, firstPrefix, ctx.Log.GetIndent())

	rootCmd = &cobra.Command{
		Use:   "git hooks", // Contains a en-space (utf-8: U+2002) to make it work...
		Short: "Githooks CLI application",
		Long:  "See further information at https://github.com/gabyx/githooks/blob/master/README.md"}

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

func Run(log cm.ILogContext) {

	ctx := NewSettings(log)
	cmd := MakeGithooksCtl(&ctx)

	err := cmd.Execute()
	if err != nil {
		_ = cmd.Help()
	}

	ctx.Log.AssertNoErrorPanic(err, "Command failed.")
}
