package update

import (
	"gabyx/githooks/build"
	ccm "gabyx/githooks/cmd/common"
	"gabyx/githooks/cmd/config"
	cm "gabyx/githooks/common"
	"gabyx/githooks/prompt"
	strs "gabyx/githooks/strings"
	"gabyx/githooks/updates"

	"github.com/spf13/cobra"
)

func runUpdate(ctx *ccm.CmdContext, setOpts *config.SetOptions, answer string) {

	switch {
	case setOpts.Set || setOpts.Unset:
		config.RunUpdate(ctx, setOpts)

	default:

		var promptCtx prompt.IContext
		if strs.IsEmpty(answer) {
			promptCtx = ctx.PromptCtx
		}

		updateAvailable, accepted, err := updates.RunUpdate(
			ctx.InstallDir,
			updates.DefaultAcceptUpdateCallback(ctx.Log, promptCtx, answer == "y"),
			&cm.ExecContext{},
			cm.UseStreams(nil, ctx.Log.GetInfoWriter(), ctx.Log.GetErrorWriter()))

		ctx.Log.AssertNoErrorPanic(err, "Running update failed.")

		switch {
		case updateAvailable:
			if accepted {
				ctx.Log.Info("Update successfully dispatched.")
			} else {
				ctx.Log.Info("Update declined.")
			}
		default:
			ctx.Log.InfoF("Githooks is at the latest version '%s'",
				build.GetBuildVersion().String())
		}
	}
}

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) *cobra.Command {

	yes := false
	no := false

	setOpts := config.SetOptions{}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Performs an update check.",
		Long: `
Executes an update check for a newer Githooks version.

If it finds one and the user accepts the prompt (or '--yes' is used)
the installer is executed to update to the latest version.

The '--enable' and '--disable' options enable or disable
the automatic checks that would normally run daily
after a successful commit event.`,
		Run: func(cmd *cobra.Command, args []string) {

			answer := ""
			if yes {
				answer = "y"
			}

			if no {
				answer = "n"
			}

			runUpdate(ctx, &setOpts, answer)
		},
	}

	updateCmd.Flags().BoolVar(&yes, "yes", false, "Always accepts a new update (non-interactive).")
	updateCmd.Flags().BoolVar(&no, "no", false, "Always deny an update and only check for it.")

	updateCmd.Flags().BoolVar(&setOpts.Set, "enable", false, "Enable daily Githooks update checks.")
	updateCmd.Flags().BoolVar(&setOpts.Unset, "disable", false, "Disable daily Githooks update checks.")

	return ccm.SetCommandDefaults(ctx.Log, updateCmd)
}
