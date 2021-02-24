package cmd

import (
	"gabyx/githooks/apps/dialog/cmd/entry"
	"gabyx/githooks/apps/dialog/cmd/file"
	"gabyx/githooks/apps/dialog/cmd/message"
	"gabyx/githooks/apps/dialog/cmd/notify"
	"gabyx/githooks/apps/dialog/cmd/options"
	"gabyx/githooks/build"

	dcm "gabyx/githooks/apps/dialog/cmd/common"
	ccm "gabyx/githooks/cmd/common"
	cm "gabyx/githooks/common"

	"github.com/spf13/cobra"
)

func addSubCommands(cmd *cobra.Command, ctx *dcm.CmdContext) {
	cmd.AddCommand(options.NewCmd(ctx))
	cmd.AddCommand(message.NewCmd(ctx))
	cmd.AddCommand(entry.NewCmd(ctx))
	cmd.AddCommand(notify.NewCmd(ctx))
	cmd.AddCommand(file.NewCmd(ctx)...)
}

// MakeDialogCtl returns the root command of the Githooks dialog executable.
func MakeDialogCtl(ctx *dcm.CmdContext) (rootCmd *cobra.Command) {

	fmt := ctx.Log.GetInfoFormatter(false)
	title := fmt("Githooks Dialog CLI [version: '%s']", build.BuildVersion)
	firstPrefix := " ▶ "
	ccm.InitTemplates(title, firstPrefix, ctx.Log.GetIndent())

	rootCmd = &cobra.Command{
		Use:   "dialog",
		Short: "Githooks dialog application similar to 'zenity'.",
		Long:  "See further information at https://github.com/gabyx/githooks/blob/main/README.md"}

	ccm.ModifyTemplate(rootCmd, ctx.Log.GetIndent())

	rootCmd.SetOut(cm.ToInfoWriter(ctx.Log))
	rootCmd.SetErr(cm.ToErrorWriter(ctx.Log))
	rootCmd.Version = build.BuildVersion

	addSubCommands(rootCmd, ctx)

	ccm.SetCommandDefaults(ctx.Log, rootCmd)

	return rootCmd
}
