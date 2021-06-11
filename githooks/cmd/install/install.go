package install

import (
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	inst "github.com/gabyx/githooks/githooks/cmd/common/install"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"

	"github.com/spf13/cobra"
)

func runInstallIntoRepo(ctx *ccm.CmdContext, nonInteractive bool) {
	_, gitDir, _ := ccm.AssertRepoRoot(ctx)

	// Check if useCoreHooksPath or core.hooksPath is set
	// and if so error out.
	value, exists := ctx.GitX.LookupConfig(git.GitCKCoreHooksPath, git.Traverse)
	ctx.Log.PanicIfF(exists, "You are using already '%s' = '%s'\n"+
		"Installing Githooks run-wrappers into '%s'\n"+
		"has no effect.",
		git.GitCKCoreHooksPath, value, gitDir)

	value, exists = ctx.GitX.LookupConfig(hooks.GitCKUseCoreHooksPath, git.GlobalScope)
	ctx.Log.PanicIfF(exists && value == git.GitCVTrue, "It appears you are using Githooks in 'core.hooksPath' mode\n"+
		"('%s' = '%s'). Installing Githooks run-wrappers into '%s'\n"+
		"may have no effect.",
		hooks.GitCKUseCoreHooksPath, value, gitDir)

	uiSettings := inst.UISettings{PromptCtx: ctx.PromptCtx}
	inst.InstallIntoRepo(ctx.Log, gitDir, nonInteractive, false, false, &uiSettings)
}

func runUninstallFromRepo(ctx *ccm.CmdContext) {
	_, gitDir, _ := ccm.AssertRepoRoot(ctx)

	// Read registered file if existing.
	var registeredGitDirs hooks.RegisterRepos
	err := registeredGitDirs.Load(ctx.InstallDir, true, true)
	ctx.Log.AssertNoErrorPanicF(err, "Could not load register file in '%s'.",
		ctx.InstallDir)

	lfsIsAvailable := git.IsLFSAvailable()
	if inst.UninstallFromRepo(ctx.Log, gitDir, lfsIsAvailable, false) {

		registeredGitDirs.Remove(gitDir)
		err := registeredGitDirs.Store(ctx.InstallDir)
		ctx.Log.AssertNoErrorPanicF(err, "Could not store register file in '%s'.",
			ctx.InstallDir)
	}
}

func runUninstall(ctx *ccm.CmdContext) {
	runUninstallFromRepo(ctx)
}

func runInstall(ctx *ccm.CmdContext, nonInteractive bool) {
	runInstallIntoRepo(ctx, nonInteractive)
}

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) []*cobra.Command {

	nonInteractive := false

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Installs Githooks run-wrappers into the current repository.",
		Long: `Installs the Githooks run-wrappers and Git config settings
into the current repository.`,
		Run: func(cmd *cobra.Command, args []string) {
			runInstall(ctx, nonInteractive)
		},
	}

	installCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Install non-interactively.")

	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstalls Githooks run-wrappers into the current repository.",
		Long: `Uninstall the Githooks run-wrappers and Git config settings
into the current repository.`,
		Run: func(cmd *cobra.Command, args []string) {
			runUninstall(ctx)
		},
	}

	return []*cobra.Command{installCmd, uninstallCmd}
}
