package cmd

import (
	"github.com/gabyx/githooks/githooks/apps/dialog/cmd/entry"
	"github.com/gabyx/githooks/githooks/apps/dialog/cmd/file"
	"github.com/gabyx/githooks/githooks/apps/dialog/cmd/message"
	"github.com/gabyx/githooks/githooks/apps/dialog/cmd/notify"
	"github.com/gabyx/githooks/githooks/apps/dialog/cmd/options"
	"github.com/gabyx/githooks/githooks/build"

	dcm "github.com/gabyx/githooks/githooks/apps/dialog/cmd/common"
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	cm "github.com/gabyx/githooks/githooks/common"

	"github.com/spf13/cobra"
)

func addSubCommands(cmd *cobra.Command, ctx *dcm.CmdContext) {
	cmd.AddCommand(options.NewCmd(ctx))
	cmd.AddCommand(message.NewCmd(ctx))
	cmd.AddCommand(entry.NewCmd(ctx))
	cmd.AddCommand(notify.NewCmd(ctx))
	cmd.AddCommand(file.NewCmd(ctx)...)

	cmd.PersistentFlags().BoolVar(&ctx.ReportAsJSON, "json", false,
		`Report the result as a JSON object on stdout.
Exit code:
	- '0' for success, and
	- '> 0' if creating the dialog failed.`)
}

// MakeDialogCtl returns the root command of the Githooks dialog executable.
func MakeDialogCtl(ctx *dcm.CmdContext) (rootCmd *cobra.Command) {

	title := cm.FormatInfoF("Githooks Dialog CLI [version: '%s']", build.BuildVersion)
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
