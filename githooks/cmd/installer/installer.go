package installer

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabyx/githooks/githooks/build"
	"github.com/gabyx/githooks/githooks/builder"
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	"github.com/gabyx/githooks/githooks/cmd/common/install"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	"github.com/gabyx/githooks/githooks/prompt"
	strs "github.com/gabyx/githooks/githooks/strings"
	"github.com/gabyx/githooks/githooks/updates"
	"github.com/gabyx/githooks/githooks/updates/download"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) *cobra.Command {

	vi := viper.New()

	var cmd = &cobra.Command{
		Use:   "installer [flags]",
		Short: "Githooks installer application.",
		Long: `Githooks installer application.
It downloads the Githooks artifacts of the current version
from a deploy source and verifies its checksums and signature.
Then it calls the installer on the new version which
will then run the installation procedure for Githooks.

See further information at https://github.com/gabyx/githooks/blob/main/README.md`,
		PreRun: ccm.PanicIfAnyArgs(ctx.Log),
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInstall(cmd, ctx, vi)
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

var MaintainedHooksDesc = "Any argument can be a hook name '<hookName>', 'all' or 'server'.\n" +
	"An optional prefix '!' means subtraction from the current set.\n" +
	"The initial value of the internally built set defaults\n" +
	"to all hook names if 'all' or 'server' is not given as first argument:\n" +
	"  - 'all' : All hooks supported by Githooks.\n" +
	"  - 'server' : Only server hooks supported by Githooks.\n" +
	"You can list them separately or comma-separated in one argument."

func defineArguments(cmd *cobra.Command, vi *viper.Viper) {
	// Internal commands
	cmd.PersistentFlags().String("config", "",
		"JSON config according to the 'Arguments' struct.")
	cm.AssertNoErrorPanic(cmd.MarkPersistentFlagDirname("config"))
	cm.AssertNoErrorPanic(cmd.PersistentFlags().MarkHidden("config"))

	cmd.PersistentFlags().String("log", "", "Log file path (only for installer).")
	cm.AssertNoErrorPanic(cmd.MarkPersistentFlagFilename("log"))

	// User commands
	cmd.PersistentFlags().Bool("dry-run", false,
		"Dry run the installation showing what's being done.")
	cmd.PersistentFlags().Bool(
		"non-interactive", false,
		"Run the installation non-interactively\n"+
			"without showing prompts.")
	cmd.PersistentFlags().Bool(
		"update", false,
		"Install and update directly to the latest\n"+
			"possible tag on the clone branch.")
	cmd.PersistentFlags().Bool(
		"skip-install-into-existing", false,
		"Skip installation into existing repositories\n"+
			"defined by a search path.")
	cmd.PersistentFlags().String(
		"prefix", "",
		"Githooks installation prefix such that\n"+
			"'<prefix>/.githooks' will be the installation directory.")
	cm.AssertNoErrorPanic(cmd.MarkPersistentFlagDirname("prefix"))

	cmd.PersistentFlags().String(
		"hooks-dir", "",
		"The preferred directory to use for maintaining Githook's hook run-wrappers\n"+
			"e.g. `~/.githooks/templates/hooks`.")
	cmd.PersistentFlags().StringSlice(
		"maintained-hooks", nil,
		"A set of hook names which are maintained in the template directory.\n"+
			MaintainedHooksDesc)

	cmd.PersistentFlags().Bool(
		"centralized", false,
		"If the install mode 'centralized' should be used which\n"+
			"sets the global 'core.hooksPath'.")

	cmd.PersistentFlags().String(
		"clone-url", "",
		"The clone url from which Githooks should clone\n"+
			"and install/update itself. Githooks tries to\n"+
			"auto-detect the deploy setting for downloading binaries.\n"+
			"You can however provide a deploy settings file yourself if\n"+
			"the auto-detection does not work (see '--deploy-settings').")
	cmd.PersistentFlags().String(
		"clone-branch", "",
		"The clone branch from which Githooks should\n"+
			"clone and install/update itself.")
	cmd.PersistentFlags().String(
		"deploy-api", "",
		"The deploy api type (e.g. ['gitea', 'github']) to use for updates\n"+
			"of the specified 'clone-url' for helping the deploy settings\n"+
			"auto-detection. For Github urls, this is not needed.")
	cmd.PersistentFlags().String(
		"deploy-settings", "",
		"The deploy settings YAML file to use for updates of the specified\n"+
			"'--clone-url'. See the documentation for further details.")

	cmd.PersistentFlags().Bool(
		"build-from-source", false,
		"If the binaries are built from source on updates instead of\n"+
			"downloaded from the deploy url.")
	cmd.PersistentFlags().StringSlice(
		"build-tags", nil,
		"Build tags for building from source (get extended with defaults).\n"+
			"You can list them separately or comma-separated in one argument.")

	cmd.PersistentFlags().Bool(
		"use-pre-release", false,
		"When fetching the latest installer, also consider pre-release versions.")

	cm.AssertNoErrorPanic(
		vi.BindPFlag("config", cmd.PersistentFlags().Lookup("config")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("log", cmd.PersistentFlags().Lookup("log")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("dryRun", cmd.PersistentFlags().Lookup("dry-run")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("nonInteractive", cmd.PersistentFlags().Lookup("non-interactive")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("update", cmd.PersistentFlags().Lookup("update")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("skipInstallIntoExisting", cmd.PersistentFlags().Lookup("skip-install-into-existing")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("maintainedHooks", cmd.PersistentFlags().Lookup("maintained-hooks")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("centralized", cmd.PersistentFlags().Lookup("centralized")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("cloneURL", cmd.PersistentFlags().Lookup("clone-url")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("cloneBranch", cmd.PersistentFlags().Lookup("clone-branch")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("deploySettings", cmd.PersistentFlags().Lookup("deploy-settings")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("deployAPI", cmd.PersistentFlags().Lookup("deploy-api")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("buildFromSource", cmd.PersistentFlags().Lookup("build-from-source")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("buildTags", cmd.PersistentFlags().Lookup("build-tags")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("usePreRelease", cmd.PersistentFlags().Lookup("use-pre-release")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("installPrefix", cmd.PersistentFlags().Lookup("prefix")))
	cm.AssertNoErrorPanic(
		vi.BindPFlag("hooksDir", cmd.PersistentFlags().Lookup("hooks-dir")))

	setupMockFlags(cmd, vi)
}

func validateArgs(log cm.ILogContext, cmd *cobra.Command, args *Arguments) {

	// Check all parsed flags to not have empty value!
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		log.PanicIfF(f.Changed && strs.IsEmpty(f.Value.String()),
			"Flag '%s' needs an non-empty value.", f.Name)
	})

	// Check deploy-settings and deploy-api are not given together.
	log.PanicIfF(strs.IsNotEmpty(args.DeployAPI) &&
		strs.IsNotEmpty(args.DeploySettings),
		"You cannot specify a deploy api type together with\n"+
			"a deploy settings file.")

	log.PanicIf(args.BuildFromSource &&
		(strs.IsNotEmpty(args.DeployAPI) || strs.IsNotEmpty(args.DeploySettings)),
		"You cannot build binaries from source together with specifying\n",
		"a deploy settings file or deploy api.")

	var err error
	args.MaintainedHooks, err = hooks.CheckHookNames(args.MaintainedHooks)
	log.AssertNoErrorPanic(err,
		"Maintained hooks are not valid.")
}

func setupSettings(
	log cm.ILogContext,
	gitx *git.Context,
	args *Arguments,
) (Settings, install.UISettings) {

	var promptx prompt.IContext
	var err error

	log.AssertNoErrorPanic(err, "Could not get current working directory.")

	if !args.NonInteractive {
		promptx, err = prompt.CreateContext(log, true, args.UseStdin)
		promptx.AddFileWriter(log.GetFileWriter())
		log.AssertNoErrorF(err, "Prompt setup failed -> using fallback.")
	}

	var installDir string
	// First check if we already have
	// an install directory set (from --prefix)
	if strs.IsNotEmpty(args.InstallPrefix) {
		var err error
		args.InstallPrefix, err = cm.ReplaceTilde(filepath.ToSlash(args.InstallPrefix))
		log.AssertNoErrorPanic(err, "Could not replace '~' character in path.")
		installDir = path.Join(args.InstallPrefix, ".githooks")

	} else {
		installDir = install.LoadInstallDir(log, gitx)
	}

	// Remove temporary directory if existing
	tempDir, err := hooks.CleanTemporaryDir(installDir)
	log.AssertNoErrorPanicF(err,
		"Could not clean temporary directory in '%s'", installDir)

	lfsHooksCache, err := hooks.NewLFSHooksCache(args.InternalTempDir)
	log.AssertNoErrorPanicF(err, "Could not setup LFS hooks cache.")

	cloneDir := hooks.GetReleaseCloneDir(installDir)
	if args.DryRun {
		// Do not clone into install dir.
		cloneDir = path.Join(args.InternalTempDir, "release")
	}

	return Settings{
			GitX:             gitx,
			InstallDir:       installDir,
			CloneDir:         cloneDir,
			TempDir:          tempDir,
			LFSHooksCache:    lfsHooksCache,
			InstalledGitDirs: make(InstallSet, 10)}, // nolint: gomnd
		install.UISettings{PromptCtx: promptx}
}

func setInstallDir(log cm.ILogContext, gitx *git.Context, installDir string) {
	log.AssertNoErrorPanic(hooks.SetInstallDir(gitx, installDir),
		"Could not set install dir '%s'", installDir)
}

func buildFromSource(
	log cm.ILogContext,
	cleanUpX *cm.InterruptContext,
	buildTags []string,
	tempDir string,
	url string,
	branch string,
	commitSHA string) updates.Binaries {

	log.InfoF("Building binaries from source at commit '%s'.", commitSHA)

	// Clone another copy of the release clone into temporary directory
	log.InfoF("Clone to temporary build directory '%s'", tempDir)
	err := git.Clone(tempDir, url, branch, -1)
	log.AssertNoErrorPanicF(err, "Could not clone release branch into '%s'.", tempDir)

	// Checkout the remote commit sha
	log.InfoF("Checkout out commit '%s'", commitSHA[0:6])
	gitx := git.NewCtxSanitizedAt(tempDir)
	err = gitx.Check("checkout",
		"-b", "update-to-"+commitSHA[0:6],
		commitSHA)

	log.AssertNoErrorPanicF(err,
		"Could not checkout update commit '%s' in '%s'.",
		commitSHA, tempDir)

	tag, _ := gitx.Get("describe", "--tags", "--abbrev=6")
	log.InfoF("Building binaries at '%s'", tag)

	// Build the binaries.
	binPath, err := builder.Build(gitx, buildTags, cleanUpX)
	log.AssertNoErrorPanicF(err, "Could not build release branch in '%s'.", tempDir)

	bins, err := cm.GetAllFiles(binPath)
	log.AssertNoErrorPanicF(err, "Could not get files in path '%s'.", binPath)

	binaries := updates.Binaries{BinDir: binPath}
	strs.Map(bins, func(s string) string {
		if cm.IsExecutable(s) {
			if strings.HasPrefix(path.Base(s), "githooks-cli") {
				binaries.Cli = s
			} else {
				binaries.Others = append(binaries.Others, s)
			}
			binaries.All = append(binaries.All, s)
		}

		return s
	})

	log.InfoF(
		"Successfully built %v binaries:\n - %s",
		len(binaries.All),
		strings.Join(
			strs.Map(binaries.All, func(s string) string { return strs.Fmt("'%s'", path.Base(s)) }),
			"\n - "))

	log.PanicIfF(
		len(binaries.All) == 0 ||
			strs.IsEmpty(binaries.Cli),
		"No binaries or Githooks executable found in '%s'", binPath)

	// Remember to build from source
	err = gitx.SetConfig(hooks.GitCKBuildFromSource, true, git.GlobalScope)
	log.AssertNoErrorF(err, "Could not store Git config '%s'.", hooks.GitCKBuildFromSource)

	return binaries
}

func getDeploySettings(
	log cm.ILogContext,
	installDir string,
	cloneURL string,
	args *Arguments) download.IDeploySettings {

	var err error
	var deploySettings download.IDeploySettings

	installDeploySettings := download.GetDeploySettingsFile(installDir)
	fileToLoad := ""
	switch {
	case strs.IsNotEmpty(args.DeploySettings):
		// If the user specified a deploy settings file use this.
		cm.DebugAssert(cm.IsFile(args.DeploySettings))
		fileToLoad = args.DeploySettings
	case strs.IsEmpty(args.DeployAPI):
		// If the user did not specify a deploy api type,
		// load the deploy settings from install dir.
		fileToLoad = installDeploySettings
	}

	if cm.IsFile(fileToLoad) {
		deploySettings, err = download.LoadDeploySettings(fileToLoad)
		log.AssertNoErrorPanicF(err, "Could not load deploy settings '%s'.", fileToLoad)
	}

	// If nothing is specified yet, try to detect it.
	if deploySettings == nil {
		deploySettings, err = detectDeploySettings(cloneURL, args.DeployAPI)
		log.AssertNoErrorF(err, "Could not auto-detect deploy settings.")
	}

	if deploySettings != nil {
		err := download.StoreDeploySettings(installDeploySettings, deploySettings)
		log.AssertNoErrorPanicF(err, "Could not store deploy settings '%s'.", installDeploySettings)
	}

	return deploySettings
}

func runInstallDispatched(
	log cm.ILogContext,
	gitx *git.Context,
	settings *Settings,
	args Arguments,
	cleanUpX *cm.InterruptContext) (bool, error) {

	var status updates.ReleaseStatus
	var err error

	log.Info("Running dispatched installer.")

	log.Info("Fetching Githooks clone...")
	setGitConfig := !args.DryRun
	status, err = updates.FetchUpdates(
		settings.CloneDir,
		args.CloneURL,
		args.CloneBranch,
		build.BuildTag,
		true,
		updates.RecloneOnWrongRemote,
		args.UsePreRelease,
		setGitConfig)

	log.AssertNoErrorPanicF(err,
		"Could not assert release clone '%s' existing",
		settings.CloneDir)

	log.DebugF("Status: %v", status)

	installer := hooks.GetInstallerExecutable(settings.InstallDir)
	log.InfoF("Githooks update available: '%v'", status.IsUpdateAvailable)

	if cm.PackageManagerEnabled {
		// We do not do either updating or downloading binaries
		// because they already exist on the system
		// at the correct place
		return false, nil
	}

	// We download/build the binaries always.
	// Only do an update if enabled and we either have
	// given the update flag or its an auto-update
	// call.
	doUpdate := updates.Enabled && args.Update && status.IsUpdateAvailable

	tag := ""
	commit := ""

	if doUpdate {
		tag = status.UpdateTag
		commit = status.UpdateCommitSHA
	} else {
		tag = status.LocalTag
		commit = status.LocalCommitSHA
	}

	log.InfoF("Getting Githooks binaries at version '%s' ...", tag)

	tempDir, err := os.MkdirTemp(args.InternalTempDir, "githooks-update-*")
	log.AssertNoErrorPanic(err, "Can not create temporary update dir in '%s'", args.InternalTempDir)

	buildFromSrc := args.BuildFromSource ||
		gitx.GetConfig(hooks.GitCKBuildFromSource, git.GlobalScope) == git.GitCVTrue

	binaries := updates.Binaries{}

	if buildFromSrc {
		log.Info("Building from source...")
		binaries = buildFromSource(
			log,
			cleanUpX,
			args.BuildTags,
			tempDir,
			status.RemoteURL,
			status.Branch,
			commit)
	}

	// We need to run deploy code too when running coverage because
	// it builds a non-instrumented binary.
	if !buildFromSrc || IsRunningCoverage {
		log.InfoF("Download '%s' from deploy source...", tag)

		deploySettings := getDeploySettings(log, settings.InstallDir, status.RemoteURL, &args)
		binaries = downloadBinaries(log, deploySettings, tempDir, tag)
	}

	installer.Cmd = binaries.Cli

	// Set variables for further update procedure...
	// Note: `args` is passed by value.
	args.InternalPostDispatch = true
	args.InternalBinaries = binaries.All
	if status.IsUpdateAvailable {
		args.InternalUpdateFromVersion = build.BuildVersion
		args.InternalUpdateTo = status.UpdateCommitSHA
	}

	if DevIsDispatchSkipped {
		return false, nil
	}

	log.PanicIfF(!cm.IsFile(installer.Cmd),
		"Githooks executable '%s' is not existing.", installer)

	return true, dispatchToInstaller(log, &installer, &args)
}

func dispatchToInstaller(log cm.ILogContext, installer cm.IExecutable, args *Arguments) error {

	log.Info("Dispatching to new installer ...")

	file, err := os.CreateTemp("", "*install-config.json")
	log.AssertNoErrorPanicF(err, "Could not create temporary file in '%s'.")
	defer os.Remove(file.Name())

	// Write the config to
	// make the installer gettings all settings
	writeArgs(log, file.Name(), args)

	// Run the installer binary
	return cm.RunExecutable(
		&cm.ExecContext{Env: os.Environ()},
		installer,
		cm.UseStreams(os.Stdin, os.Stdout, os.Stderr),
		"--config", file.Name())
}

// findHookTemplateDir returns the Git hook template directory.
func findHookTemplateDir(
	log cm.ILogContext,
	gitx *git.Context,
	installDir string,
	installMode install.InstallModeType,
	haveInstall bool,
	nonInteractive bool,
	promptx prompt.IContext) string {

	log.InfoF("Find hooks template dir for install mode '%s'.",
		installMode.Name())

	hooksDir, err := install.FindHooksDir(log, gitx, haveInstall)
	log.AssertNoErrorF(err, "Error while determining default hook template directory.")

	if err == nil && strs.IsNotEmpty(hooksDir) {
		return hooksDir
	}

	// No folder found: Try setup a new folder.
	if nonInteractive {
		return setupNewHooksDir(log, installDir, nil)
	}

	// Try to search for it on disk (only normal install mode)
	answer, err := promptx.ShowOptions(
		"Could not find the Git hooks directory.\n"+
			"Do you want to search for it?",
		"(yes, No)",
		"y/N",
		"Yes", "No")
	log.AssertNoErrorF(err, "Could not show prompt.")

	if answer == "y" {
		templateDir := searchHooksDirOnDisk(log, promptx)

		if strs.IsNotEmpty(templateDir) {
			return path.Join(templateDir, "hooks")
		}
	}

	// Try to set up as new
	answer, err = promptx.ShowOptions(
		"Do you want to set up a new Git hooks directory?",
		"(yes, No)",
		"y/N",
		"Yes", "No")
	log.AssertNoErrorF(err, "Could not show prompt.")

	if answer == "y" {
		return setupNewHooksDir(log, installDir, promptx)
	}

	return ""
}

func searchPreCommitFile(log cm.ILogContext, startDirs []string, promptx prompt.IContext) (result string) {

	for _, dir := range startDirs {

		log.InfoF("Searching for potential locations in '%s'...", dir)

		settings := cm.CreateDefaultProgressSettings(
			"Searching ...", "Still searching ...")
		taskIn := install.PreCommitSearchTask{Dir: dir}

		resultTask, err := cm.RunTaskWithProgress(&taskIn, log, 300*time.Second, settings) //nolint: gomnd
		if err != nil {
			log.AssertNoError(err, "Searching failed.")
			return //nolint: nlreturn
		}

		taskOut := resultTask.(*install.PreCommitSearchTask)
		cm.DebugAssert(taskOut != nil, "Wrong output.")

		for _, match := range taskOut.Matches { //nolint: staticcheck

			templateDir := path.Dir(path.Dir(filepath.ToSlash(match)))

			answer, err := promptx.ShowOptions(
				strs.Fmt("--> Is it '%s'", templateDir),
				"(yes, No)",
				"y/N",
				"Yes", "No")
			log.AssertNoErrorF(err, "Could not show prompt.")

			if answer == "y" {
				result = templateDir

				break //nolint: nlreturn
			}
		}
	}

	return
}

func searchHooksDirOnDisk(log cm.ILogContext, promptx prompt.IContext) string {

	first, second := GetDefaultTemplateSearchDir()

	templateDir := searchPreCommitFile(log, first, promptx)

	if strs.IsEmpty(templateDir) {

		answer, err := promptx.ShowOptions(
			"Git hooks directory not found\n"+
				"Do you want to keep searching?",
			"(yes, No)",
			"y/N",
			"Yes", "No")

		log.AssertNoErrorF(err, "Could not show prompt.")

		if answer == "y" {
			templateDir = searchPreCommitFile(log, second, promptx)
		}
	}

	return templateDir
}

func setupNewHooksDir(log cm.ILogContext, installDir string, promptx prompt.IContext) string {
	hooksDir := path.Join(installDir, "templates", "hooks")

	homeDir, err := homedir.Dir()
	cm.AssertNoErrorPanic(err, "Could not get home directory.")

	if promptx != nil {
		var err error
		hooksDir, err = promptx.ShowEntry(
			"Enter the target folder ('~' allowed)",
			hooksDir,
			nil)
		log.AssertNoErrorF(err, "Could not show prompt.")
	}

	hooksDir = cm.ReplaceTildeWith(hooksDir, homeDir)
	log.AssertNoErrorPanicF(err, "Could not replace tilde '~' in '%s'.", hooksDir)

	return hooksDir
}

func setupInstallMode(
	log cm.ILogContext,
	gitx *git.Context,
	installDir string,
	givenHooksDir string,
	haveInstall bool,
	installMode install.InstallModeType,
	nonInteractive bool,
	dryRun bool,
	promptx prompt.IContext) (hooksDir string) {

	switch {
	case strs.IsNotEmpty(givenHooksDir):
		// Template directory given, use this.
		hooksDir = givenHooksDir

	case strs.IsEmpty(givenHooksDir):
		hooksDir = findHookTemplateDir(
			log,
			gitx,
			installDir,
			installMode,
			haveInstall,
			nonInteractive,
			promptx)

		log.PanicIfF(strs.IsEmpty(hooksDir),
			"Could not determine Git hook template directory.")
	}

	log.InfoF("Hooks dir set to '%s'.", hooksDir)

	// Set the global Git configuration.
	setDirectoryForInstallMode(log, gitx, installMode, hooksDir, dryRun)

	return
}

func setDirectoryForInstallMode(
	log cm.ILogContext,
	gitx *git.Context,
	installMode install.InstallModeType,
	hooksDir string,
	dryRun bool) {

	prefix := "Setting"
	if dryRun {
		prefix = "[dry run] Would set"
	}

	warnOnTemplateHooks := func() string {
		tD := gitx.GetConfig(git.GitCKInitTemplateDir, git.GlobalScope)
		msg := ""
		if strs.IsNotEmpty(tD) && cm.IsDirectory(path.Join(tD, "hooks")) {
			d := path.Join(tD, "hooks")
			files, err := cm.GetAllFiles(d)
			log.AssertNoErrorPanicF(err, "Could not get files in '%s'.", d)

			if len(files) > 0 {
				msg = strs.Fmt(
					"The 'init.templateDir' setting is currently set to\n"+
						"'%s'\n"+ // nolint: gGetInstallModeNameoconst
						"and contains '%v' potential hooks.\n", tD, len(files))
			}
		}

		tDEnv := os.Getenv("GIT_TEMPLATE_DIR")
		if strs.IsNotEmpty(tDEnv) && cm.IsDirectory(path.Join(tDEnv, "hooks")) {
			d := path.Join(tDEnv, "hooks")
			files, err := cm.GetAllFiles(d)
			log.AssertNoErrorPanicF(err, "Could not get files in '%s'.", d)

			if len(files) > 0 {
				msg += strs.Fmt(
					"The environment variable 'GIT_TEMPLATE_DIR' is currently set to\n"+
						"'%s'\n"+
						"and contains '%v' potential hooks.\n", tDEnv, len(files))
			}
		}

		return msg
	}

	if !dryRun {
		err := os.MkdirAll(hooksDir, cm.DefaultFileModeDirectory)
		log.AssertNoErrorPanicF(err,
			"Could not assert directory '%s' exists",
			hooksDir)
	}

	switch installMode {
	case install.InstallModeTypeV.UseGlobalCoreHooksPath:
		log.InfoF("%s '%s' to '%s'.", prefix, git.GitCKCoreHooksPath, hooksDir)

		if !dryRun {
			err := gitx.SetConfig(hooks.GitCKInstallMode,
				install.InstallModeTypeV.UseGlobalCoreHooksPath.Name(), git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")

			err = gitx.SetConfig(hooks.GitCKPathForUseCoreHooksPath,
				hooksDir, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")

			err = gitx.SetConfig(git.GitCKCoreHooksPath, hooksDir, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")
		}

		// Warnings:
		// Check if hooks might not run...
		msg := warnOnTemplateHooks()
		log.WarnIf(strs.IsNotEmpty(msg),
			msg+
				"These hooks might get installed but\n"+
				"ignored because '%s' is also set.\n"+
				"It is recommended to either remove the files or run\n"+
				"the Githooks installation without the '--centralized'\n"+
				"parameter.", git.GitCKCoreHooksPath)

	case install.InstallModeTypeV.Manual:
		log.InfoF("%s '%s' to '%s'.", prefix, hooks.GitCKPathForUseCoreHooksPath, hooksDir)

		if !dryRun {
			err := gitx.SetConfig(hooks.GitCKInstallMode, install.InstallModeTypeV.Manual.Name(), git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")

			err = gitx.SetConfig(hooks.GitCKPathForUseCoreHooksPath, hooksDir, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")
		}

		msg := warnOnTemplateHooks()
		log.WarnIf(strs.IsNotEmpty(msg),
			msg+
				"These hooks might get installed but\n"+
				"defeats the purpose of running Githooks in manual install mode."+
				"It is recommended to remove the files.")
	}
}

func setupHookTemplates(
	log cm.ILogContext,
	gitx *git.Context,
	hookTemplateDir string,
	maintainedHooks []string,
	lfsHooksCache hooks.LFSHooksCache,
	nonInteractive bool,
	dryRun bool,
	uiSettings *install.UISettings) {

	if dryRun {
		log.InfoF("[dry run] Would install Githooks run-wrappers into '%s'.",
			hookTemplateDir)

		return
	}

	log.InfoF("Saving Githooks run-wrapper to '%s' :", hookTemplateDir)

	var err error
	var hookNames []string
	if len(maintainedHooks) == 0 {
		hookNames, maintainedHooks, err = hooks.GetMaintainedHooks(gitx, git.GlobalScope)
		log.AssertNoError(err, "Could not get maintained hooks config.")
	} else {
		hookNames, err = hooks.UnwrapHookNames(maintainedHooks)
		log.AssertNoErrorPanic(err, "Could not build maintained hook list.")
	}

	nLFSHooks, err := hooks.InstallRunWrappers(
		hookTemplateDir,
		hookNames,
		func(dest string) {
			log.InfoF(" %s '%s'", cm.ListItemLiteral, path.Base(dest))
		},
		install.GetHookDisableCallback(log, gitx, nonInteractive, uiSettings),
		lfsHooksCache,
		log)
	log.AssertNoErrorPanicF(err, "Could not install run-wrappers into '%s'.", hookTemplateDir)

	if nLFSHooks != 0 {
		log.InfoF("Installed '%v' Githooks run-wrappers and '%v' missing LFS hooks into '%s'.",
			len(hookNames), nLFSHooks, hookTemplateDir)
	} else {
		log.InfoF("Installed '%v' Githooks run-wrappers into '%s'.",
			len(hookNames), hookTemplateDir)
	}

	// Set maintained hooks in global settings, such that
	// local repository Githooks installs are in alignment
	// to the setup template directory.
	err = hooks.SetMaintainedHooks(gitx, maintainedHooks, git.GlobalScope)
	log.AssertNoError(err, "Could not set git config.")
}

func installBinaries(
	log cm.ILogContext,
	tempDir string,
	installDir string,
	binaries []string,
	dryRun bool) {

	binDir := hooks.GetBinaryDir(installDir)
	err := os.MkdirAll(binDir, cm.DefaultFileModeDirectory)
	log.AssertNoErrorPanicF(err, "Could not create binary dir '%s'.", binDir)

	msg := strs.Map(binaries, func(s string) string { return strs.Fmt(" • '%s'", path.Base(s)) })
	if dryRun {
		log.InfoF("[dry run] Would install binaries:\n%s\n"+"to '%s'.", msg, binDir)
		return // nolint:nlreturn
	}

	log.InfoF("Installing binaries:\n%s\n"+"to '%s'.", strings.Join(msg, "\n"), binDir)

	// Remove old legacy-named binaries, just in case if they are still there.
	for _, name := range []string{"runner", "cli", "dialog"} {
		_ = os.Remove(path.Join(binDir, name))
		_ = os.Remove(path.Join(binDir, name+".exe"))
	}

	for _, binary := range binaries {
		dest := path.Join(binDir, path.Base(binary))

		err := os.MkdirAll(tempDir, cm.DefaultFileModeDirectory)
		log.AssertNoErrorPanicF(err,
			"Could not create backup folder for old binaries")

		err = cm.CopyFileWithBackup(binary, dest, tempDir, false)
		log.AssertNoErrorPanicF(err,
			"Could not move file '%s' to '%s'.", binary, dest)
	}

	// Set CLI executable alias.
	cli := hooks.GetCLIExecutable(installDir)
	err = hooks.SetCLIExecutableAlias(cli.Cmd)
	log.AssertNoErrorPanicF(err,
		"Could not set Git config 'alias.hooks' to '%s'.", cli.Cmd)

	runner := hooks.GetRunnerExecutable(installDir)
	err = hooks.SetRunnerExecutableAlias(runner)
	log.AssertNoErrorPanic(err,
		"Could not set runner executable alias '%s'.", runner)

	dialog := hooks.GetDialogExecutable(installDir)
	err = hooks.SetDialogExecutableConfig(dialog)
	log.AssertNoErrorPanic(err,
		"Could not set dialog executable to '%s'.", dialog)
}

func setupAutomaticUpdateChecks(
	log cm.ILogContext,
	gitx *git.Context,
	nonInteractive bool,
	dryRun bool,
	promptx prompt.IContext) {

	enabled, isSet := updates.GetUpdateCheckSettings(gitx)
	promptMsg := ""

	switch {
	case !isSet:
		promptMsg = "Would you like to enable automatic update checks,\ndone once a day after a commit?"
	case enabled:
		return // Already enabled.
	default:
		log.Info("Automatic update checks are currently disabled.")
		if nonInteractive {
			return
		}
		promptMsg = "Would you like to re-enable them,\ndone once a day after a commit?"
	}

	var activate bool

	if nonInteractive {
		activate = true
	} else {
		answer, err := promptx.ShowOptions(
			promptMsg,
			"(Yes, no)",
			"Y/n", "Yes", "No")
		log.AssertNoErrorF(err, "Could not show prompt.")

		activate = answer == "y"
	}

	if activate {
		if dryRun {
			log.Info("[dry run] Would enable automatic update checks.")
		} else {

			err := updates.SetUpdateCheckSettings(true, false)
			if log.AssertNoErrorF(err, "Failed to enable automatic update checks.") {
				log.Info("Automatic update checks are now enabled.")
			}
		}
	} else {
		log.Info(
			"If you change your mind in the future, you can enable it by running:",
			"  $ git hooks update --enable")
	}
}

func installIntoExistingRepos(
	log cm.ILogContext,
	gitx *git.Context,
	lfsHooksCache hooks.LFSHooksCache,
	nonInteractive bool,
	dryRun bool,
	skipReadme bool,
	installedRepos InstallSet,
	registeredRepos *hooks.RegisterRepos,
	uiSettings *install.UISettings) {

	// Show prompt and run callback.
	install.PromptExistingRepos(
		log,
		gitx,
		nonInteractive,
		false,
		uiSettings.PromptCtx,

		func(gitDir string) {

			if install.InstallIntoRepo(
				log, gitDir, lfsHooksCache, nil,
				nonInteractive, dryRun,
				skipReadme, uiSettings) {

				registeredRepos.Insert(gitDir)
				installedRepos.Insert(gitDir)
			}
		})

}

func installIntoRegisteredRepos(
	log cm.ILogContext,
	lfsHooksCache hooks.LFSHooksCache,
	nonInteractive bool,
	dryRun bool,
	skipReadme bool,
	installedRepos InstallSet,
	registeredRepos *hooks.RegisterRepos,
	uiSettings *install.UISettings) {

	if len(registeredRepos.GitDirs) == 0 {
		return
	}

	dirsWithNoInstalls := strs.Filter(registeredRepos.GitDirs,
		func(s string) bool {
			return !installedRepos.Exists(s)
		})

	// Show prompt and run callback.
	install.PromptRegisteredRepos(
		log,
		dirsWithNoInstalls,
		nonInteractive,
		false,
		uiSettings.PromptCtx,
		func(gitDir string) {
			if install.InstallIntoRepo(
				log, gitDir, lfsHooksCache, nil,
				nonInteractive, dryRun,
				skipReadme, uiSettings) {

				registeredRepos.Insert(gitDir)
				installedRepos.Insert(gitDir)
			}
		})

}

func setupSharedRepositories(
	log cm.ILogContext,
	installDir string,
	dryRun bool,
	uiSettings *install.UISettings) {

	gitx := git.NewCtx()
	sharedRepos := gitx.GetConfigAll(hooks.GitCKShared, git.GlobalScope)

	var question string
	if len(sharedRepos) != 0 {
		question = "Looks like you already have shared hook\n" +
			"repositories setup, do you want to change them now?"
	} else {
		question = "You can set up shared hook repositories to avoid\n" +
			"duplicating common hooks across repositories you work on\n" +
			"See information on what are these in the project's documentation:\n" +
			strs.Fmt("'%s#shared-hook-repositories'\n\n", hooks.GithooksWebpage) +
			strs.Fmt("Note: you can also have a '%s' file listing the\n", hooks.GetRepoSharedFileRel()) +
			"repositories where you keep the shared hook files.\n\n" +
			"Would you like to set up shared hook repos now?"
	}

	answer, err := uiSettings.PromptCtx.ShowOptions(
		question,
		"(yes, No)", "y/N", "Yes", "No")
	log.AssertNoError(err, "Could not show prompt")

	if answer == "n" {
		return
	}

	log.Info("Let's input shared hook repository urls",
		"one-by-one and leave the input empty to stop.")

	entries, err := uiSettings.PromptCtx.ShowEntryMulti(
		"Enter the clone URL of a shared repository",
		"", // exit answer
		prompt.ValidatorAnswerNotEmpty)

	if err != nil {
		log.Error("Could not show prompt. Not settings shared hook repositories.")
		return // nolint: nlreturn
	}

	if dryRun {
		log.InfoF("Would have setup all shared hook repositories in global Git config.")

		return
	}

	// Unset all shared configs.
	err = gitx.UnsetConfig(hooks.GitCKShared, git.GlobalScope)
	log.AssertNoErrorF(err,
		"Could not unset Git config '%s'.\n"+
			"Failed to setup shared hook repositories.", hooks.GitCKShared)
	if err != nil {
		return
	}

	// Add all entries.
	for _, entry := range entries {
		err := gitx.AddConfig(hooks.GitCKShared, entry, git.GlobalScope)
		log.AssertNoError(err,
			"Could not add Git config '%s'.\n"+
				"Failed to setup shared hook repositories.", hooks.GitCKShared)
		if err != nil {
			return
		}
	}

	if len(entries) == 0 {
		log.InfoF(
			"Shared hook repositories are now unset.\n"+
				"If you want to set them up again in the future\n"+
				"run this script again, or change the '%s'\n"+
				"Git config variable manually.\n"+
				"Note: Shared hook repos listed in the '%s'\n",
			"file will still be executed", hooks.GitCKShared, hooks.GetRepoSharedFileRel())
	} else {

		updated, err := hooks.UpdateAllSharedHooks(log, gitx, installDir, "", nil)
		log.ErrorIf(err != nil, "Could not update shared hook repositories.")
		log.InfoF("Updated '%v' shared hook repositories.", updated)

		log.InfoF(
			"Shared hook repositories have been set up.\n"+
				"You can change them any time by running this script\n"+
				"again, or manually by changing the 'githooks.shared'\n"+
				"Git config variable.\n"+
				"Note: you can also list the shared hook repos per\n"+
				"project within the '%s' file", hooks.GetRepoSharedFileRel())
	}

}

func storeSettings(log cm.ILogContext, settings *Settings, uiSettings *install.UISettings) {
	// Store cached UI values back.

	if strs.IsNotEmpty(uiSettings.DeleteDetectedLFSHooks) {
		err := git.NewCtx().SetConfig(
			hooks.GitCKDeleteDetectedLFSHooksAnswer, uiSettings.DeleteDetectedLFSHooks, git.GlobalScope)
		log.AssertNoError(err, "Could not store config '%v'.", uiSettings.DeleteDetectedLFSHooks)
	}

	err := settings.RegisteredGitDirs.Store(settings.InstallDir)
	log.AssertNoError(err,
		"Could not store registered file in '%s'.",
		settings.InstallDir)

	if err != nil {
		for _, gitDir := range settings.InstalledGitDirs.ToList() {
			// For each installedGitDir entry, mark the repository as registered.
			err := hooks.MarkRepoRegistered(git.NewCtxAt(gitDir))
			log.AssertNoErrorF(err, "Could not mark Git directory '%s' as registered.", gitDir)
		}
	}

}

func updateClone(log cm.ILogContext, cloneDir string, updateToSHA string) {

	if strs.IsEmpty(updateToSHA) {
		return // We don't need to update the release clone.
	}

	commitSHA, err := updates.MergeUpdates(cloneDir, false)

	log.AssertNoErrorF(err,
		"Could not finalize by updating the local branch to the\n"+
			"remote branch in the release clone\n"+
			"'%s'.\n"+
			"This seems rather odd.\n"+
			"Either fix the problems or delete the clone\n"+
			"to trigger a new checkout.", cloneDir)

	cm.DebugAssert(err != nil ||
		commitSHA == updateToSHA, "Wrong updateToSHA.")
}

func thankYou(log cm.ILogContext) {
	log.InfoF("All done! Enjoy!\n"+
		"Please support the project by starring the project\n"+
		"at '%s', and report\n"+
		"bugs or missing features or improvements as issues.\n"+
		"Thanks!\n", hooks.GithooksWebpage)
}

func determineInstallMode(log cm.ILogContext, args *Arguments, gitx *git.Context) (bool, install.InstallModeType) {
	haveInstall, installedMode := install.GetInstallMode(gitx)

	var installMode install.InstallModeType

	if strs.IsNotEmpty(args.InternalUpdateFromVersion) {

		if !haveInstall {
			log.WarnF("Could not determine Githooks install mode.\n"+
				"Install seams corrupt?.\n"+
				"Taking default '%s'.", install.InstallModeTypeV.Manual.Name())
			installedMode = install.InstallModeTypeV.Manual
		}

		installMode = installedMode

	} else {

		installMode = install.MapInstallerArgsToInstallMode(
			args.Centralized)

		if haveInstall && installMode != installedMode {
			log.PanicF(
				"You seem to have already installed Githooks in mode '%s'\n"+
					"and we are going to reinstall it in mode '%s'.\n"+
					"Please uninstall Githooks first by running:\n"+
					"  $ git hooks uninstaller\n"+
					"for a proper cleanup.",
				installedMode.Name(),
				installMode.Name())
		}
	}

	return haveInstall, installMode
}

func runInstaller(
	log cm.ILogContext,
	gitx *git.Context,
	settings *Settings,
	uiSettings *install.UISettings,
	args *Arguments) {

	if strs.IsEmpty(args.InternalUpdateFromVersion) {
		log.InfoF("Running install at current version '%s' ...", build.BuildVersion)
	} else {
		log.InfoF("Running install from '%s' -> '%s' ...", args.InternalUpdateFromVersion, build.BuildVersion)
	}

	transformLegacyGitConfigSettings(log, gitx)

	err := settings.RegisteredGitDirs.Load(settings.InstallDir, true, true)
	log.AssertNoErrorPanicF(err, "Could not load register file in '%s'.", settings.InstallDir)

	haveInstall, installMode := determineInstallMode(log, args, gitx)
	log.InfoF("Githooks already installed: '%v' [mode: '%v'].", haveInstall, installMode.Name())

	settings.HookTemplateDir = setupInstallMode(
		log,
		gitx,
		settings.InstallDir,
		args.HooksDir,
		haveInstall,
		installMode,
		args.NonInteractive,
		args.DryRun,
		uiSettings.PromptCtx)

	if !cm.PackageManagerEnabled && len(args.InternalBinaries) != 0 {
		installBinaries(
			log,
			settings.TempDir,
			settings.InstallDir,
			args.InternalBinaries,
			args.DryRun)
	}

	setupHookTemplates(
		log,
		gitx,
		settings.HookTemplateDir,
		args.MaintainedHooks,
		settings.LFSHooksCache,
		args.NonInteractive,
		args.DryRun,
		uiSettings)

	if updates.Enabled {
		setupAutomaticUpdateChecks(log, gitx, args.NonInteractive, args.DryRun, uiSettings.PromptCtx)
	}

	if !args.SkipInstallIntoExisting &&
		installMode != install.InstallModeTypeV.UseGlobalCoreHooksPath {

		installIntoExistingRepos(
			log,
			gitx,
			settings.LFSHooksCache,
			args.NonInteractive,
			args.DryRun,
			false,
			settings.InstalledGitDirs,
			&settings.RegisteredGitDirs,
			uiSettings)

	}

	if installMode != install.InstallModeTypeV.UseGlobalCoreHooksPath {
		installIntoRegisteredRepos(
			log,
			settings.LFSHooksCache,
			args.NonInteractive,
			args.DryRun,
			false, // Do not skip readme setup.
			settings.InstalledGitDirs,
			&settings.RegisteredGitDirs,
			uiSettings)
	}

	if !args.NonInteractive {
		setupSharedRepositories(
			log,
			settings.InstallDir,
			args.DryRun,
			uiSettings)
	}

	if !args.DryRun {
		storeSettings(log, settings, uiSettings)
		updateClone(log, settings.CloneDir, args.InternalUpdateTo)
	}
}

func addInstallerLog(path string, log cm.ILogContext) (isOwned bool, resPath string) {
	var err error
	var file *os.File

	if strs.IsEmpty(path) {
		file, err = os.CreateTemp("", "githooks-installer-*.log")
		log.AssertNoErrorF(err, "Failed to create installer log at '%v'", path)
		isOwned = true
	} else {
		file, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, cm.DefaultFileModeFile)
		log.AssertNoErrorF(err, "Failed to append to installer log at '%v'", path)
	}

	if err != nil {
		return
	}

	log.AddFileWriter(file)
	resPath = file.Name()

	return
}

func assertOneInstallerRunning(log cm.ILogContext, interruptCtx *cm.InterruptContext) {
	lockFile := path.Join(os.TempDir(), "githooks-installer-lock")
	if exists, _ := cm.IsPathExisting(lockFile); exists {
		log.PanicF("Only one Githooks installer can run at the same time. "+
			"Maybe delete the lock file '%v", lockFile)
	}

	log.DebugF("Created lockfile '%s'.", lockFile)
	err := cm.TouchFile(lockFile, true)
	log.AssertNoErrorPanic(err, "Could not create lockfile '%v'.", lockFile)

	// Remove the lock on any exit.
	deleteLock := func() {
		log.DebugF("Remove lockfile '%s'.", lockFile)
		err := os.Remove(lockFile)
		log.AssertNoError(err, "Lockfile not removed?")
	}
	interruptCtx.AddHandler(deleteLock)
}

func setupTempDir(log cm.ILogContext, args *Arguments) (isOwned bool) {
	if strs.IsEmpty(args.InternalTempDir) {
		dir := os.TempDir()
		var err error
		args.InternalTempDir, err = os.MkdirTemp(dir, "githooks-installer-*")
		log.AssertNoErrorPanicF(err, "Could not create temp. dir in '%s'.", dir)
		isOwned = true
	}

	return
}

func runInstall(cmd *cobra.Command, ctx *ccm.CmdContext, vi *viper.Viper) error {

	args := Arguments{}
	log := ctx.Log
	logStats := ctx.LogStats

	initArgs(log, &args, vi)
	validateArgs(log, cmd, &args)

	isOwnedLog := false
	isOwnedLog, args.Log = addInstallerLog(args.Log, log)
	if RemoveInstallerLogOnSuccess && isOwnedLog {
		defer func() {
			if r := recover(); r != nil {
				panic(r)
			}
			// Only delete the log file if no panic, and no errors and
			// when it is owned.
			if RemoveInstallerLogOnSuccess && logStats.ErrorCount() == 0 {
				log.RemoveFileWriter()
				_ = os.Remove(args.Log)
			}
		}()
	}

	isOwnedTemp := setupTempDir(log, &args)
	if isOwnedTemp {
		// When we own the temp. directory we are cleaning it at the end.
		ctx.CleanupX.AddHandler(func() { _ = os.Remove(args.InternalTempDir) })
		defer os.Remove(args.InternalTempDir)
	}

	log.InfoF("Githooks Installer [version: %s]", build.BuildVersion)
	dt := time.Now()
	log.InfoF("Started at: %s", dt.String())

	log.InfoF("Log file: '%s'", args.Log)
	settings, uiSettings := setupSettings(log, ctx.GitX, &args)

	log.DebugF("Arguments: %+v", args)
	log.DebugF("Settings: %+v", settings)

	if !args.DryRun {
		setInstallDir(log, ctx.GitX, settings.InstallDir)
	}

	if !args.InternalPostDispatch {
		assertOneInstallerRunning(log, ctx.CleanupX)

		// Dispatch from an old installer to a new one.
		isDispatched, err := runInstallDispatched(log, ctx.GitX, &settings, args, ctx.CleanupX)
		log.MoveFileWriterToEnd() // We are logging to the same file. Move it to the end.
		if err != nil {
			return ctx.NewCmdExit(1, "%v", err)
		}

		if isDispatched {
			return nil
		}
		// intended fallthrough ... (only debug or on package manager enabled)
	}

	runInstaller(log, ctx.GitX, &settings, &uiSettings, &args)

	if logStats.ErrorCount() == 0 {
		thankYou(log)
	} else {
		log.ErrorF("Tried my best at installing, but\n"+
			" • %v errors\n"+
			" • %v warnings\n"+
			"occurred!", logStats.ErrorCount(), logStats.WarningCount())

		return ctx.NewCmdExit(1, "Install failed.")
	}

	return nil
}

func transformLegacyGitConfigSettings(log cm.ILogContext, gitx *git.Context) {
	// Server hooks.
	useOnlyServerHooks := gitx.GetConfig("githooks.maintainOnlyServerHooks", git.GlobalScope)
	if useOnlyServerHooks == git.GitCVTrue {
		err := hooks.SetMaintainedHooks(gitx, []string{"server"}, git.GlobalScope)
		log.AssertNoError(err, "Could not set maintained hooks to 'server'.")
	}

	_ = git.NewCtx().UnsetConfig("githooks.maintainOnlyServerHooks", git.GlobalScope)

	// Version 2.X to 3
	// ================

	// AutoUpdate to UpdateCheck
	autoUpdateEnabled := gitx.GetConfig("githooks.autoUpdateEnabled", git.GlobalScope)
	if strs.IsNotEmpty(autoUpdateEnabled) {
		err := updates.SetUpdateCheckSettings(autoUpdateEnabled == git.GitCVTrue, false)
		log.AssertNoError(err, "Could not set automatic update check.")
	}
	_ = git.NewCtx().UnsetConfig("githooks.autoUpdateEnabled", git.GlobalScope)

	// Transform old install modes from v2 to v3
	// =========================================
	hasInstallBeforeV3 := gitx.GetConfig(hooks.GitCKInstallMode, git.GlobalScope) == "" &&
		gitx.GetConfig(hooks.GitCKRunner, git.GlobalScope) != ""

	if hasInstallBeforeV3 {
		useCoreHooksPath := gitx.GetConfig("githooks.useCoreHooksPath", git.GlobalScope)
		useManual := gitx.GetConfig("githooks.useManual", git.GlobalScope)

		switch {
		case useCoreHooksPath == git.GitCVTrue:
			log.Warn("Transform old `--use-core-hookspath` install to new `--centralized` install.")
			err := gitx.SetConfig(
				hooks.GitCKInstallMode,
				install.InstallModeTypeV.UseGlobalCoreHooksPath.Name(),
				git.GlobalScope)
			log.AssertNoError(err, "Could not set install mode.")

		case useManual == git.GitCVTrue:
			log.Warn("Transform old manual install to new normal install.")
			err := gitx.SetConfig(
				hooks.GitCKInstallMode,
				install.InstallModeTypeV.Manual.Name(),
				git.GlobalScope)
			log.AssertNoError(err, "Could not set install mode.")

			manualTemplateDir := gitx.GetConfig("githooks.manualTemplateDir", git.GlobalScope)
			err = gitx.SetConfig(
				hooks.GitCKPathForUseCoreHooksPath,
				path.Join(manualTemplateDir, "hooks"),
				git.GlobalScope)
			log.AssertNoError(err, "Could not set path to use for core hooks path.")

		case useManual != git.GitCVTrue && useCoreHooksPath != git.GitCVTrue:
			log.Warn("Transform old normal (`init.templateDir`) install to new `--manual` install.")
			// That was template dir mode.
			// It gets transformed to `manual` mode.
			err := gitx.SetConfig(hooks.GitCKInstallMode, "manual", git.GlobalScope)
			log.AssertNoError(err, "Could not set install mode.")

			templateDir := gitx.GetConfig(git.GitCKInitTemplateDir, git.GlobalScope)
			if strs.IsNotEmpty(templateDir) {
				err = gitx.SetConfig(hooks.GitCKPathForUseCoreHooksPath, path.Join(templateDir, "hooks"), git.GlobalScope)
				log.AssertNoError(err, "Could not set path to use for core hooks path.")
			}
		}
	}

	// Delete all
	_ = git.NewCtx().UnsetConfig("githooks.useCoreHooksPath", git.GlobalScope)
	_ = git.NewCtx().UnsetConfig("githooks.useManual", git.GlobalScope)
	_ = git.NewCtx().UnsetConfig("githooks.manualTemplateDir", git.GlobalScope)
	// =========================

}
