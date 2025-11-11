package trust

import (
	"os"

	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/hooks"

	"github.com/spf13/cobra"
)

type trustOption = int

const (
	trustAdd    trustOption = 0
	trustRevoke trustOption = 1
	trustForget trustOption = 2
	trustDelete trustOption = 3
)

func runTrust(ctx *ccm.CmdContext, opt trustOption) {
	repoRoot, _, _ := ccm.AssertRepoRoot(ctx)
	file := hooks.GetTrustMarkerFile(repoRoot)

	switch opt {
	case trustAdd:
		err := cm.TouchFile(file, true)
		ctx.Log.AssertNoErrorPanicF(err, "Could not touch trust marker '%s'.", file)
		ctx.Log.Info("The trust marker is added to the repository.")

		err = hooks.SetTrustAllSetting(ctx.GitX, true, false)
		ctx.Log.AssertNoErrorPanic(err, "Could not set trust settings.")
		ctx.Log.Info("The current repository is now trusted.")

		if !ctx.GitX.IsBareRepo() {
			ctx.Log.Info("Do not forget to commit and push it!")
		}
	case trustForget:
		_, isSet := hooks.GetTrustAllSetting(ctx.GitX)
		if !isSet {
			ctx.Log.Info("The current repository does not have trust settings.")
		} else {
			err := hooks.SetTrustAllSetting(ctx.GitX, false, true)
			ctx.Log.AssertNoErrorPanic(err, "Could not unset trust settings.")
		}

		ctx.Log.Info("The current repository is no longer trusted.")

	case trustRevoke:
		fallthrough
	case trustDelete:
		err := hooks.SetTrustAllSetting(ctx.GitX, false, false)
		ctx.Log.AssertNoErrorPanicF(err, "Could not set trust settings.")
		ctx.Log.Info("The current repository is no longer trusted.")
	}

	if opt == trustDelete {
		err := os.RemoveAll(file)
		ctx.Log.AssertNoErrorPanicF(err, "Could not remove trust marker '%s'.", file)

		ctx.Log.Info("The trust marker is removed from the repository.")

		if !ctx.GitX.IsBareRepo() {
			ctx.Log.Info("Do not forget to commit and push it!")
		}
	}
}

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) *cobra.Command {
	trustCmd := &cobra.Command{
		Use:   "trust",
		Short: "Manages settings related to trusted repositories.",
		Long: `Sets up, or reverts the trusted setting for the local repository.

When called without arguments, it marks the local repository as trusted.

The 'revoke' argument resets the already accepted trust setting,
and the 'delete' argument also deletes the trust marker.

The 'forget' option unsets the trust setting, asking for accepting
it again next time, if the repository is marked as trusted.`,
		Run: func(cmd *cobra.Command, args []string) {
			runTrust(ctx, trustAdd)
		}}

	trustRevokeCmd := &cobra.Command{
		Use:   "revoke",
		Short: `Revoke repository trust settings.`,
		Run: func(cmd *cobra.Command, args []string) {
			runTrust(ctx, trustRevoke)
		}}

	trustForgetCmd := &cobra.Command{
		Use:   "forget",
		Short: `Forget repository trust settings.`,
		Run: func(cmd *cobra.Command, args []string) {
			runTrust(ctx, trustForget)
		}}

	trustDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: `Delete repository trust settings.`,
		Run: func(cmd *cobra.Command, args []string) {
			runTrust(ctx, trustDelete)
		}}

	trustCmd.AddCommand(
		ccm.SetCommandDefaults(ctx.Log, trustRevokeCmd),
		ccm.SetCommandDefaults(ctx.Log, trustForgetCmd),
		ccm.SetCommandDefaults(ctx.Log, trustDeleteCmd),
		ccm.SetCommandDefaults(ctx.Log, NewTrustHooksCmd(ctx)))

	trustCmd.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		ccm.CheckGithooksSetup(ctx.Log, ctx.GitX)
	}

	return ccm.SetCommandDefaults(ctx.Log, trustCmd)
}
