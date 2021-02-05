package disable

import (
	ccm "gabyx/githooks/cmd/common"
	"gabyx/githooks/git"
	"gabyx/githooks/hooks"

	"github.com/spf13/cobra"
)

type disableOptions struct {
	Reset  bool
	Global bool
}

// RunDisable disables Githooks completely.
func RunDisable(ctx *ccm.CmdContext, reset bool, onlyPrint bool, global bool) {

	var scope git.ConfigScope
	var fmt string

	if global {
		scope = git.GlobalScope
		fmt = "globally"
	} else {
		ccm.AssertRepoRoot(ctx)
		scope = git.LocalScope
		fmt = "in the current repository"
	}

	if onlyPrint {
		conf := ctx.GitX.GetConfig(hooks.GitCKDisable, scope)
		if conf == "true" {
			ctx.Log.InfoF("Githooks is disabled %s.", fmt)
		} else {
			ctx.Log.InfoF("Githooks is not disabled %s.", fmt)
		}

		return
	}

	if reset {
		err := ctx.GitX.UnsetConfig(hooks.GitCKDisable, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not unset '%s' Git config '%s'.", scope, hooks.GitCKDisable)
		ctx.Log.InfoF("Enabled Githooks %s.", fmt)

	} else {
		err := ctx.GitX.SetConfig(hooks.GitCKDisable, true, scope)
		ctx.Log.AssertNoErrorPanicF(err, "Could not set '%s' Git config '%s'.", scope, hooks.GitCKDisable)
		ctx.Log.InfoF("Disabled Githooks %s.", fmt)
	}
}

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) *cobra.Command {

	var disableOpts disableOptions

	disableCmd := &cobra.Command{
		Use:   "disable [flags]",
		Short: "Disables Githooks in the current repository or globally.",
		Long: `Disables running any Githooks in the current repository or globally.

LFS hooks and replaced previous hooks are still executed.`,
		PreRun: ccm.PanicIfAnyArgs(ctx.Log),
		Run: func(cmd *cobra.Command, args []string) {
			RunDisable(ctx, disableOpts.Reset, false, disableOpts.Global)
		}}

	disableCmd.Flags().BoolVar(&disableOpts.Reset, "reset", false,
		`Resets the disable setting and enables running
hooks by Githooks again.`)

	disableCmd.Flags().BoolVar(&disableOpts.Global, "global", false,
		`Enable/disable Githooks globally instead of locally.`)

	return ccm.SetCommandDefaults(ctx.Log, disableCmd)
}
