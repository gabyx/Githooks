package install

import (
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	inst "github.com/gabyx/githooks/githooks/cmd/common/install"
	"github.com/gabyx/githooks/githooks/cmd/installer"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"

	"github.com/spf13/cobra"
)

func runInstallIntoRepo(ctx *ccm.CmdContext, maintainedHooks []string, nonInteractive bool) {
	_, gitDir, _ := ccm.AssertRepoRoot(ctx)
	uiSettings := inst.UISettings{PromptCtx: ctx.PromptCtx}

	_, installMode := inst.GetInstallMode(ctx.GitX)
	ctx.Log.PanicIfF(installMode == inst.InstallModeTypeV.Centralized,
		"Githooks is installed in '%s' mode and\n"+
			"installing into the current repository has no effect.",
		inst.InstallModeTypeV.Centralized.Name())

	lfsHooksCache, err := hooks.NewLFSHooksCache(hooks.GetTemporaryDir(ctx.InstallDir))
	ctx.Log.AssertNoErrorPanicF(err, "Could not create LFS hooks cache.")

	var hooksToMaintain []string
	if maintainedHooks != nil {
		maintainedHooks, err = hooks.CheckHookNames(maintainedHooks)
		ctx.Log.AssertNoErrorPanic(err, "Maintained hooks are not valid.")
		hooksToMaintain, err = hooks.UnwrapHookNames(maintainedHooks)
		ctx.Log.AssertNoErrorPanic(err, "Maintained hooks are not valid.")
	}

	installed := inst.InstallIntoRepo(
		ctx.Log, gitDir,
		lfsHooksCache, hooksToMaintain,
		nonInteractive, false, false, &uiSettings)
	ctx.Log.PanicIf(!installed, "Install had errors.")

	if maintainedHooks != nil {
		err = hooks.SetMaintainedHooks(ctx.GitX, maintainedHooks, git.LocalScope)
		ctx.Log.AssertNoErrorPanic(err, "Could not set maintained hooks config value.")
	}

	err = hooks.RegisterRepo(gitDir, ctx.InstallDir, false, false)
	ctx.Log.AssertNoError(err, "Could not register repository '%s'.", gitDir)
	err = hooks.MarkRepoRegistered(ctx.GitX)
	ctx.Log.AssertNoError(err, "Could not mark repository '%s' as registered.", gitDir)
}

func runUninstallFromRepo(ctx *ccm.CmdContext, fullUninstall bool) {
	_, gitDir, _ := ccm.AssertRepoRoot(ctx)

	_, installMode := inst.GetInstallMode(ctx.GitX)
	ctx.Log.WarnIfF(installMode == inst.InstallModeTypeV.Centralized,
		"Githooks is installed in '%s' mode and\n"+
			"uninstalling from the current repository has no effect.",
		inst.InstallModeTypeV.Centralized.Name())

	// Read registered file if existing.
	var registeredGitDirs hooks.RegisterRepos
	err := registeredGitDirs.Load(ctx.InstallDir, true, true)
	ctx.Log.AssertNoErrorPanicF(err, "Could not load register file in '%s'.",
		ctx.InstallDir)

	lfsHooksCache, err := hooks.NewLFSHooksCache(hooks.GetTemporaryDir(ctx.InstallDir))
	ctx.Log.AssertNoErrorPanicF(err, "Could not create LFS hooks cache.")

	if inst.UninstallFromRepo(ctx.Log, gitDir, lfsHooksCache, fullUninstall) {

		registeredGitDirs.Remove(gitDir)
		err := registeredGitDirs.Store(ctx.InstallDir)
		ctx.Log.AssertNoErrorPanicF(err, "Could not store register file in '%s'.",
			ctx.InstallDir)
	}
}

func runUninstall(ctx *ccm.CmdContext, fullUninstall bool) {
	runUninstallFromRepo(ctx, fullUninstall)
}

func runInstall(ctx *ccm.CmdContext, maintainedHooks []string, nonInteractive bool) {
	runInstallIntoRepo(ctx, maintainedHooks, nonInteractive)
}

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) []*cobra.Command {

	var maintainedHooks *[]string
	nonInteractive := false

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Installs Githooks run-wrappers into the current repository.",
		Long: `Installs the Githooks run-wrappers and Git config settings
into the current repository.`,
		Run: func(cmd *cobra.Command, args []string) {
			runInstall(ctx, *maintainedHooks, nonInteractive)
		},
	}

	installCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "Install non-interactively.")
	maintainedHooks = installCmd.Flags().StringSlice(
		"maintained-hooks", nil,
		"A set of hook names which are maintained in this repository.\n"+
			installer.MaintainedHooksDesc)

	fullUninstall := false
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstalls Githooks run-wrappers into the current repository.",
		Long: `Uninstall the Githooks run-wrappers and Git config settings
into the current repository.`,
		Run: func(cmd *cobra.Command, args []string) {
			runUninstall(ctx, fullUninstall)
		},
	}

	installCmd.Flags().BoolVar(&fullUninstall, "full", false,
		"Uninstall also Git config values of Githooks and cached\n"+
			"settings (checksums etc.) inside the repository.")

	installCmd.PersistentPostRun = func(_ *cobra.Command, _ []string) {
		ccm.CheckGithooksSetup(ctx.Log, ctx.GitX)
	}

	return []*cobra.Command{installCmd, uninstallCmd}
}
