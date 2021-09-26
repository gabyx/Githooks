package uninstaller

import (
	"os"
	"runtime"
	"strings"

	"github.com/gabyx/githooks/githooks/build"
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	"github.com/gabyx/githooks/githooks/cmd/common/install"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	"github.com/gabyx/githooks/githooks/prompt"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) *cobra.Command {

	vi := viper.New()

	var cmd = &cobra.Command{
		Use:   "uninstaller",
		Short: "Githooks uninstaller application.",
		Long: "Githooks uninstaller application\n" +
			"See further information at https://github.com/gabyx/githooks/blob/main/README.md",
		PreRun: ccm.PanicIfAnyArgs(ctx.Log),
		Run: func(cmd *cobra.Command, _ []string) {
			runUninstall(ctx, vi)
		}}

	defineArguments(cmd, vi)

	return ccm.SetCommandDefaults(ctx.Log, cmd)
}

func initArgs(log cm.ILogContext, args *Arguments, vi *viper.Viper) {

	config := vi.GetString("config")
	if strs.IsNotEmpty(config) {
		vi.SetConfigFile(config)
		err := vi.ReadInConfig()
		log.AssertNoErrorPanicF(err, "Could not read config file '%s'.", config)
	}

	err := vi.Unmarshal(&args)
	log.AssertNoErrorPanicF(err, "Could not unmarshal parameters.")
}

func writeArgs(log cm.ILogContext, file string, args *Arguments) {
	err := cm.StoreJSON(file, args)
	log.AssertNoErrorPanicF(err, "Could not write arguments to '%s'.", file)
}

func defineArguments(cmd *cobra.Command, vi *viper.Viper) {
	// Internal commands
	cmd.PersistentFlags().String("config", "",
		"JSON config according to the 'Arguments' struct.")
	cm.AssertNoErrorPanic(cmd.MarkPersistentFlagDirname("config"))
	cm.AssertNoErrorPanic(cmd.PersistentFlags().MarkHidden("config"))

	// User commands
	cmd.PersistentFlags().Bool(
		"non-interactive", false,
		"Run the uninstallation non-interactively\n"+
			"without showing prompts.")

	cm.AssertNoErrorPanic(
		vi.BindPFlag("config", cmd.PersistentFlags().Lookup("config")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("nonInteractive", cmd.PersistentFlags().Lookup("non-interactive")))

	setupMockFlags(cmd, vi)
}

func setupSettings(log cm.ILogContext, gitx *git.Context, args *Arguments) (Settings, UISettings) {

	var promptx prompt.IContext
	var err error

	log.AssertNoErrorPanic(err, "Could not get current working directory.")

	if !args.NonInteractive {
		promptx, err = prompt.CreateContext(log, false, args.UseStdin)
		log.AssertNoErrorF(err, "Prompt setup failed -> using fallback.")
	}

	installDir := install.LoadInstallDir(log, gitx)

	// Safety check.
	log.PanicIfF(!strings.Contains(installDir, ".githooks"),
		"Uninstall path at '%s' needs to contain '.githooks'.")

	// Remove temporary directory if existing
	tempDir, err := hooks.CleanTemporaryDir(installDir)
	log.AssertNoErrorPanicF(err,
		"Could not clean temporary directory in '%s'", installDir)

	lfsHooksCache, err := hooks.NewLFSHooksCache(hooks.GetTemporaryDir(installDir))
	log.AssertNoErrorPanicF(err, "Could not create LFS hooks cache.")

	return Settings{
			Gitx:               gitx,
			InstallDir:         installDir,
			CloneDir:           hooks.GetReleaseCloneDir(installDir),
			TempDir:            tempDir,
			UninstalledGitDirs: make(UninstallSet, 10), // nolint: gomnd
			LFSHooksCache:      lfsHooksCache},
		UISettings{PromptCtx: promptx}
}

func prepareDispatch(log cm.ILogContext, settings *Settings, args *Arguments) bool {

	uninstaller := hooks.GetUninstallerExecutable(settings.InstallDir)
	if !cm.IsFile(uninstaller.Cmd) {
		log.WarnF("There is no existing Githooks executable present\n"+
			"in install dir '%s'.\n"+
			"Your installation is corrupt.\n"+
			"We will continue to uninstall agnostically with this installer.",
			settings.InstallDir)

		return false
	}

	// Set variables for further uninstall procedure.
	args.InternalPostDispatch = true

	runUninstaller(log, &uninstaller, args)

	return true
}

func runUninstaller(log cm.ILogContext, uninstaller cm.IExecutable, args *Arguments) {

	log.Info("Dispatching to uninstaller ...")

	file, err := os.CreateTemp("", "*uninstall-config.json")
	log.AssertNoErrorPanicF(err, "Could not create temporary file in '%s'.")
	defer os.Remove(file.Name())

	// Write the config to
	// make the uninstaller gettings all settings
	writeArgs(log, file.Name(), args)

	// Run the uninstaller binary
	err = cm.RunExecutable(
		&cm.ExecContext{},
		uninstaller,
		cm.UseStreams(os.Stdin, log.GetInfoWriter(), log.GetInfoWriter()),
		"--config", file.Name())

	log.AssertNoErrorPanic(err, "Running uninstaller failed.")
}

func thankYou(log cm.ILogContext) {
	log.InfoF(
		"All done! Enjoy!\n"+
			"If you ever want to reinstall the hooks, just follow\n"+
			"the install instructions at '%s'.", hooks.GithooksWebpage)
}

func uninstallFromExistingRepos(
	log cm.ILogContext,
	gitx *git.Context,
	lfsHooksCache hooks.LFSHooksCache,
	nonInteractive bool,
	uninstalledRepos UninstallSet,
	registeredRepos *hooks.RegisterRepos,
	uiSettings *UISettings) {

	// Show prompt and run callback.
	install.PromptExistingRepos(
		log,
		gitx,
		nonInteractive,
		true,
		uiSettings.PromptCtx,
		func(gitDir string) {

			if install.UninstallFromRepo(log, gitDir, lfsHooksCache, true) {

				registeredRepos.Remove(gitDir)
				uninstalledRepos.Insert(gitDir)
			}
		})
}

func uninstallFromRegisteredRepos(
	log cm.ILogContext,
	lfsHooksCache hooks.LFSHooksCache,
	nonInteractive bool,
	uninstalledRepos UninstallSet,
	registeredRepos *hooks.RegisterRepos,
	uiSettings *UISettings) {

	if len(registeredRepos.GitDirs) == 0 {
		return
	}

	dirsWithNoUninstalls := strs.Filter(registeredRepos.GitDirs,
		func(s string) bool {
			return !uninstalledRepos.Exists(s)
		})

	// Show prompt and run callback.
	install.PromptRegisteredRepos(
		log,
		dirsWithNoUninstalls,
		nonInteractive,
		true,
		uiSettings.PromptCtx,
		func(gitDir string) {
			if install.UninstallFromRepo(log, gitDir, lfsHooksCache, true) {

				registeredRepos.Remove(gitDir)
				uninstalledRepos.Insert(gitDir)
			}
		})
}

func cleanTemplateDir(log cm.ILogContext, gitx *git.Context, lfsHooksCache hooks.LFSHooksCache) {
	installUsesCoreHooksPath := gitx.GetConfig(hooks.GitCKUseCoreHooksPath, git.GlobalScope)

	hookTemplateDir, err := install.FindHookTemplateDir(gitx, installUsesCoreHooksPath == git.GitCVTrue)
	log.AssertNoErrorF(err, "Error while determining default hook template directory.")

	if strs.IsEmpty(hookTemplateDir) {
		log.ErrorF(
			"Git hook templates directory not found.\n" +
				"Installation is corrupt!")
	} else {
		_, err = hooks.UninstallRunWrappers(hookTemplateDir, lfsHooksCache)
		log.AssertNoErrorF(err, "Could not uninstall Githooks run-wrappers in\n'%s'.", hookTemplateDir)
	}
}

func cleanSharedClones(log cm.ILogContext, installDir string) {
	sharedDir := hooks.GetSharedDir(installDir)

	if cm.IsDirectory(sharedDir) {
		err := os.RemoveAll(sharedDir)
		log.AssertNoErrorF(err,
			"Could not delete shared directory '%s'.", sharedDir)
	}
}

func deleteDir(log cm.ILogContext, dir string, tempDir string) {
	if runtime.GOOS == cm.WindowsOsName {
		// On Windows we cannot move binaries which we execute at the moment.
		// We move everything to a new random folder inside tempDir
		// and notify the user.

		tmp := cm.GetTempPath(tempDir, "old-binaries")
		err := os.Rename(dir, tmp)
		log.AssertNoErrorF(err, "Could not move dir\n'%s' to '%s'.", dir, tmp)

	} else {
		// On Unix system we can simply remove the binary dir,
		// even if we are running the installer
		err := os.RemoveAll(dir)
		log.AssertNoErrorF(err, "Could not delete dir '%s'.", dir)
	}
}

func cleanBinaries(
	log cm.ILogContext,
	installDir string,
	tempDir string) {

	binDir := hooks.GetBinaryDir(installDir)

	if cm.IsDirectory(binDir) {
		deleteDir(log, binDir, tempDir)
	}
}

func cleanReleaseClone(
	log cm.ILogContext,
	installDir string) {

	cloneDir := hooks.GetReleaseCloneDir(installDir)

	if cm.IsDirectory(cloneDir) {
		err := os.RemoveAll(cloneDir)
		log.AssertNoErrorF(err,
			"Could not delete release clone directory '%s'.", cloneDir)
	}
}

func cleanGitConfig(log cm.ILogContext, gitx *git.Context) {

	// Remove core.hooksPath if we are using it.
	pathForUseCoreHooksPath := gitx.GetConfig(hooks.GitCKPathForUseCoreHooksPath, git.GlobalScope)
	coreHooksPath := gitx.GetConfig(git.GitCKCoreHooksPath, git.GlobalScope)

	if coreHooksPath == pathForUseCoreHooksPath {
		err := gitx.UnsetConfig(git.GitCKCoreHooksPath, git.GlobalScope)
		log.AssertNoError(err, "Could not unset global Git config 'core.hooksPath'.")
	}

	// Remove all global configs
	for _, k := range hooks.GetGlobalGitConfigKeys() {

		log.AssertNoErrorF(gitx.UnsetConfig(k, git.GlobalScope),
			"Could not unset global Git config '%s'.", k)
	}

	// Remove legacy values
	k := "githooks.checksumCacheDir"
	log.AssertNoErrorF(gitx.UnsetConfig(k, git.GlobalScope),
		"Could not unset global Git config '%s'.", k)
}

func cleanRegister(log cm.ILogContext, installDir string) {

	registerFile := hooks.GetRegisterFile(installDir)

	if cm.IsFile(registerFile) {
		err := os.Remove(registerFile)
		log.AssertNoError(err,
			"Could not delete register file '%s'.", registerFile)
	}
}

func runUninstallSteps(
	log cm.ILogContext,
	settings *Settings,
	uiSettings *UISettings,
	args *Arguments) {

	// Read registered file if existing.
	// We ensured during load, that only existing Git directories are listed.
	err := settings.RegisteredGitDirs.Load(settings.InstallDir, true, true)
	log.AssertNoErrorPanicF(err, "Could not load register file in '%s'.",
		settings.InstallDir)

	log.InfoF("Running uninstall at version '%s' ...", build.BuildVersion)

	uninstallFromExistingRepos(
		log,
		settings.Gitx,
		settings.LFSHooksCache,
		args.NonInteractive,
		settings.UninstalledGitDirs,
		&settings.RegisteredGitDirs,
		uiSettings)

	uninstallFromRegisteredRepos(
		log,
		settings.LFSHooksCache,
		args.NonInteractive,
		settings.UninstalledGitDirs,
		&settings.RegisteredGitDirs,
		uiSettings)

	cleanTemplateDir(log, settings.Gitx, settings.LFSHooksCache)

	cleanSharedClones(log, settings.InstallDir)
	cleanReleaseClone(log, settings.InstallDir)
	cleanBinaries(log, settings.InstallDir, settings.TempDir)
	cleanRegister(log, settings.InstallDir)

	cleanGitConfig(log, settings.Gitx)
}

func runUninstall(ctx *ccm.CmdContext, vi *viper.Viper) {

	log := ctx.Log
	args := Arguments{}

	log.InfoF("Githooks Uninstaller [version: %s]", build.BuildVersion)

	initArgs(log, &args, vi)

	log.DebugF("Arguments: %+v", args)

	settings, uiSettings := setupSettings(log, ctx.GitX, &args)

	if !args.InternalPostDispatch {
		if isDispatched := prepareDispatch(log, &settings, &args); isDispatched {
			return
		}
	}

	runUninstallSteps(log, &settings, &uiSettings, &args)

	if ctx.LogStats.ErrorCount() == 0 {
		thankYou(log)
	} else {
		log.ErrorF("Tried my best at uninstalling, but\n"+
			" • %v errors\n"+
			" • %v warnings\n"+
			"occurred!", ctx.LogStats.ErrorCount(), ctx.LogStats.WarningCount())
	}
}
