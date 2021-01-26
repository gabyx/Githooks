//go:generate go run -mod=vendor ../../tools/embed-files.go
package main

import (
	"gabyx/githooks/apps/install"
	"gabyx/githooks/build"
	"gabyx/githooks/builder"
	ccm "gabyx/githooks/cmd/common"
	cm "gabyx/githooks/common"
	"gabyx/githooks/git"
	"gabyx/githooks/hooks"
	"gabyx/githooks/prompt"
	strs "gabyx/githooks/strings"
	"gabyx/githooks/updates"
	"gabyx/githooks/updates/download"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var log cm.ILogContext
var args = Arguments{}

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "installer [flags]",
	Short: "Githooks installer application",
	Long: "Githooks installer application\n" +
		"See further information at https://github.com/gabyx/githooks/blob/master/README.md",
	PreRun: ccm.PanicIfAnyArgs(log),
	Run:    runInstall}

// Run adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Run() {
	cobra.OnInitialize(initArgs)

	rootCmd.SetOut(cm.ToInfoWriter(log))
	rootCmd.SetErr(cm.ToErrorWriter(log))
	rootCmd.Version = build.BuildVersion

	defineArguments(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initArgs() {

	config := viper.GetString("config")
	if strs.IsNotEmpty(config) {
		viper.SetConfigFile(config)
		err := viper.ReadInConfig()
		log.AssertNoErrorPanicF(err, "Could not read config file '%s'.", config)
	}

	err := viper.Unmarshal(&args)
	log.AssertNoErrorPanicF(err, "Could not unmarshal parameters.")
}

func writeArgs(file string, args *Arguments) {
	err := cm.StoreJSON(file, args)
	log.AssertNoErrorPanicF(err, "Could not write arguments to '%s'.", file)
}

func defineArguments(rootCmd *cobra.Command) {
	// Internal commands
	rootCmd.PersistentFlags().String("config", "",
		"JSON config according to the 'Arguments' struct.")
	cm.AssertNoErrorPanic(rootCmd.MarkPersistentFlagDirname("config"))
	cm.AssertNoErrorPanic(rootCmd.PersistentFlags().MarkHidden("config"))

	rootCmd.PersistentFlags().Bool("internal-auto-update", false,
		"Internal argument, do not use!") // @todo Remove this...
	cm.AssertNoErrorPanic(rootCmd.PersistentFlags().MarkHidden("internal-auto-update"))

	// User commands
	rootCmd.PersistentFlags().Bool("dry-run", false,
		"Dry run the installation showing whats being done.")
	rootCmd.PersistentFlags().Bool(
		"non-interactive", false,
		"Run the installation non-interactively\n"+
			"without showing prompts.")
	rootCmd.PersistentFlags().Bool(
		"skip-install-into-existing", false,
		"Skip installation into existing repositories\n"+
			"defined by a search path.")
	rootCmd.PersistentFlags().String(
		"prefix", "",
		"Githooks installation prefix such that\n"+
			"'<prefix>/.githooks' will be the installation directory.")
	cm.AssertNoErrorPanic(rootCmd.MarkPersistentFlagDirname("prefix"))

	rootCmd.PersistentFlags().String(
		"template-dir", "",
		"The preferred template directory to use.")
	rootCmd.PersistentFlags().Bool(
		"only-server-hooks", false,
		"Only install and maintain server hooks.")
	rootCmd.PersistentFlags().Bool(
		"use-core-hookspath", false,
		"If the install mode 'core.hooksPath' should be used.")

	rootCmd.PersistentFlags().String(
		"clone-url", "",
		"The clone url from which Githooks should clone\n"+
			"and install/update itself. Githooks tries to\n"+
			"auto-detect the deploy setting for downloading binaries.\n"+
			"You can however provide a deploy settings file yourself if\n"+
			"the auto-detection does not work (see '--deploy-settings').")
	rootCmd.PersistentFlags().String(
		"clone-branch", "",
		"The clone branch from which Githooks should\n"+
			"clone and install/update itself.")
	rootCmd.PersistentFlags().String(
		"deploy-api", "",
		"The deploy api type (e.g. ['gitea', 'github']) to use for updates\n"+
			"of the specified 'clone-url' for helping the deploy settings\n"+
			"auto-detection. For Github urls, this is not needed.")
	rootCmd.PersistentFlags().String(
		"deploy-settings", "",
		"The deploy settings YAML file to use for updates of the specified\n"+
			"'--clone-url'. See the documentation for further details.")

	rootCmd.PersistentFlags().Bool(
		"build-from-source", false,
		"If the binaries are built from source instead of\n"+
			"downloaded from the deploy url.")
	rootCmd.PersistentFlags().StringArray(
		"build-tags", nil,
		"Build tags for building from source (get extended with defaults).")

	rootCmd.Args = cobra.NoArgs

	cm.AssertNoErrorPanic(
		viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config")))
	// @todo Remove this internalAutoUpdate...
	cm.AssertNoErrorPanic(
		viper.BindPFlag("internalAutoUpdate", rootCmd.PersistentFlags().Lookup("internal-auto-update")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("dryRun", rootCmd.PersistentFlags().Lookup("dry-run")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("nonInteractive", rootCmd.PersistentFlags().Lookup("non-interactive")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("skipInstallIntoExisting", rootCmd.PersistentFlags().Lookup("skip-install-into-existing")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("onlyServerHooks", rootCmd.PersistentFlags().Lookup("only-server-hooks")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("useCoreHooksPath", rootCmd.PersistentFlags().Lookup("use-core-hookspath")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("cloneURL", rootCmd.PersistentFlags().Lookup("clone-url")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("cloneBranch", rootCmd.PersistentFlags().Lookup("clone-branch")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("deploySettings", rootCmd.PersistentFlags().Lookup("deploy-settings")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("deployAPI", rootCmd.PersistentFlags().Lookup("deploy-api")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("buildFromSource", rootCmd.PersistentFlags().Lookup("build-from-source")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("buildTags", rootCmd.PersistentFlags().Lookup("build-tags")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("installPrefix", rootCmd.PersistentFlags().Lookup("prefix")))
	cm.AssertNoErrorPanic(
		viper.BindPFlag("templateDir", rootCmd.PersistentFlags().Lookup("template-dir")))

	setupMockFlags(rootCmd)
}

func validateArgs(cmd *cobra.Command, args *Arguments) {

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

}

func setMainVariables(args *Arguments) (Settings, install.UISettings) {

	var promptCtx prompt.IContext
	var err error

	cwd, err := os.Getwd()
	log.AssertNoErrorPanic(err, "Could not get current working directory.")

	if !args.NonInteractive {
		promptCtx, err = prompt.CreateContext(log, &cm.ExecContext{}, nil, false, args.UseStdin)
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
		installDir = install.LoadInstallDir(log)
	}

	// Remove temporary directory if existing
	tempDir, err := hooks.CleanTemporaryDir(installDir)
	log.AssertNoErrorPanicF(err,
		"Could not clean temporary directory in '%s'", installDir)

	return Settings{
			Cwd:              cwd,
			InstallDir:       installDir,
			CloneDir:         hooks.GetReleaseCloneDir(installDir),
			TempDir:          tempDir,
			InstalledGitDirs: make(InstallSet, 10)},
		install.UISettings{PromptCtx: promptCtx}
}

func setInstallDir(installDir string) {
	log.AssertNoErrorPanic(hooks.SetInstallDir(installDir),
		"Could not set install dir '%s'", installDir)
}

func buildFromSource(
	buildTags []string,
	tempDir string,
	url string,
	branch string,
	commitSHA string) updates.Binaries {

	log.Info("Building binaries from source ...")

	// Clone another copy of the release clone into temporary directory
	log.InfoF("Clone to temporary build directory '%s'", tempDir)
	err := git.Clone(tempDir, url, branch, -1)
	log.AssertNoErrorPanicF(err, "Could not clone release branch into '%s'.", tempDir)

	// Checkout the remote commit sha
	log.InfoF("Checkout out commit '%s'", commitSHA[0:6])
	gitx := git.CtxC(tempDir)
	err = gitx.Check("checkout",
		"-b", "update-to-"+commitSHA[0:6],
		commitSHA)

	log.AssertNoErrorPanicF(err,
		"Could not checkout update commit '%s' in '%s'.",
		commitSHA, tempDir)

	tag, _ := gitx.Get("describe", "--tags", "--abbrev=6")
	log.InfoF("Building binaries at '%s'", tag)

	// Build the binaries.
	binPath, err := builder.Build(tempDir, buildTags)
	log.AssertNoErrorPanicF(err, "Could not build release branch in '%s'.", tempDir)

	bins, err := cm.GetAllFiles(binPath)
	log.AssertNoErrorPanicF(err, "Could not get files in path '%s'.", binPath)

	binaries := updates.Binaries{BinDir: binPath}
	strs.Map(bins, func(s string) string {
		if cm.IsExecutable(s) {
			if strings.HasPrefix(path.Base(s), "installer") {
				binaries.Installer = s
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
			strs.IsEmpty(binaries.Installer),
		"No binaries or installer found in '%s'", binPath)

	// Remember to build from source
	err = gitx.SetConfig(hooks.GitCK_BuildFromSource, true, git.GlobalScope)
	log.AssertNoErrorF(err, "Could not store Git config '%s'.", hooks.GitCK_BuildFromSource)

	return binaries
}

func getDeploySettings(
	installDir string,
	cloneUrl string,
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
		deploySettings, err = detectDeploySettings(cloneUrl, args.DeployAPI)
		log.AssertNoErrorF(err, "Could not auto-detect deploy settings.")
	}

	if deploySettings != nil {
		err := download.StoreDeploySettings(installDeploySettings, deploySettings)
		log.AssertNoErrorPanicF(err, "Could not store deploy settings '%s'.", installDeploySettings)
	}

	return deploySettings
}

func prepareDispatch(settings *Settings, args *Arguments) bool {

	skipPrerelease := !(git.Ctx().GetConfig(hooks.GitCK_AutoUpdateUsePrerelease, git.GlobalScope) == "true")

	var status updates.ReleaseStatus
	var err error

	if args.InternalAutoUpdate {
		log.Info("Executing auto update...")

		status, err = updates.GetStatus(settings.CloneDir, true, skipPrerelease)
		log.AssertNoErrorPanic(err,
			"Could not get status of release clone '%s'",
			settings.CloneDir)

	} else {
		log.Info("Fetching update in Githooks clone...")
		status, err = updates.FetchUpdates(
			settings.CloneDir,
			args.CloneURL,
			args.CloneBranch,
			true,
			updates.RecloneOnWrongRemote,
			skipPrerelease)

		log.AssertNoErrorPanicF(err,
			"Could not assert release clone '%s' existing",
			settings.CloneDir)
	}

	updateAvailable := status.LocalCommitSHA != status.RemoteCommitSHA

	cm.PanicIfF(args.InternalAutoUpdate && !updateAvailable,
		"An autoupdate should only be triggered when and update is found.")

	installer := hooks.GetInstallerExecutable(settings.InstallDir)
	haveInstaller := cm.IsFile(installer)

	// We download/build the binaries if an update is available
	// or the installer is missing.
	binaries := updates.Binaries{}

	if updateAvailable || !haveInstaller {

		log.Info("Getting Githooks binaries...")

		tempDir, err := ioutil.TempDir(os.TempDir(), "*-githooks-update")
		log.AssertNoErrorPanic(err, "Can not create temporary update dir in '%s'", os.TempDir())
		defer os.RemoveAll(tempDir)

		if args.BuildFromSource {

			log.Info("Building from clone...")
			binaries = buildFromSource(
				args.BuildTags,
				tempDir,
				status.RemoteURL,
				status.Branch,
				status.RemoteCommitSHA)

		} else {
			tag := status.UpdateTag
			if strs.IsEmpty(tag) {
				tag = status.LocalTag
			}

			log.InfoF("Download '%s' from deploy source...", tag)

			deploySettings := getDeploySettings(settings.InstallDir, status.RemoteURL, args)
			binaries = downloadBinaries(log, deploySettings, tempDir, tag)
		}

		installer = binaries.Installer
	}

	// Set variables for further update procedure...
	args.InternalPostDispatch = true
	args.InternalBinaries = binaries.All
	if updateAvailable {
		args.InternalUpdateFromVersion = build.BuildVersion
		args.InternalUpdateTo = status.UpdateCommitSHA
	}

	if DevIsDispatchSkipped {
		return false
	}

	runInstaller(installer, args)

	return true
}

func runInstaller(installer string, args *Arguments) {

	log.Info("Dispatching to new installer ...")
	log.PanicIfF(!cm.IsFile(installer), "Installer '%s' is not existing.", installer)

	file, err := ioutil.TempFile("", "*install-config.json")
	log.AssertNoErrorPanicF(err, "Could not create temporary file in '%s'.")
	defer os.Remove(file.Name())

	// Write the config to
	// make the installer gettings all settings
	writeArgs(file.Name(), args)

	// Run the installer binary
	err = cm.RunExecutable(
		&cm.ExecContext{},
		&cm.Executable{Cmd: installer},
		cm.UseStreams(os.Stdin, log.GetInfoWriter(), log.GetErrorWriter()),
		"--config", file.Name())

	log.AssertNoErrorPanic(err, "Running installer failed.")
}

// findGitHookTemplates returns the Git hook template directory
// and optional a Git template dir which gets only set in case of
// not using the core.hooksPath method.
func findGitHookTemplates(
	installDir string,
	useCoreHooksPath bool,
	nonInteractive bool,
	promptCtx prompt.IContext) (string, string) {

	installUsesCoreHooksPath := git.Ctx().GetConfig(hooks.GitCK_UseCoreHooksPath, git.GlobalScope)
	haveInstall := strs.IsNotEmpty(installUsesCoreHooksPath)

	hookTemplateDir, err := install.FindHookTemplateDir(useCoreHooksPath || installUsesCoreHooksPath == "true")
	log.AssertNoErrorF(err, "Error while determining default hook template directory.")
	if err == nil && strs.IsNotEmpty(hookTemplateDir) {
		return hookTemplateDir, ""
	}

	// If we have an installation, and have not found
	// the template folder by now...
	log.PanicIfF(haveInstall,
		"Your installation is corrupt.\n"+
			"The global Git config 'githooks.useCoreHooksPath = %v'\n"+
			"is set but the corresponding hook templates directory\n"+
			"is not found. Is 'core.hooksPath' unset?", installUsesCoreHooksPath)

	// 4. Try setup new folder if running non-interactively
	// and no folder is found by now
	if nonInteractive {
		templateDir := setupNewTemplateDir(installDir, nil)
		return path.Join(templateDir, "hooks"), templateDir // nolint:nlreturn
	}

	// 5. Try to search for it on disk
	answer, err := promptCtx.ShowPromptOptions(
		"Could not find the Git hook template directory.\n"+
			"Do you want to search for it?",
		"(yes, No)",
		"y/N",
		"Yes", "No")
	log.AssertNoErrorF(err, "Could not show prompt.")

	if answer == "y" {

		templateDir := searchTemplateDirOnDisk(promptCtx)

		if strs.IsNotEmpty(templateDir) {

			if useCoreHooksPath {
				return path.Join(templateDir, "hooks"), ""
			}

			// If we dont use core.hooksPath, we ask
			// if the user wants to continue setting this as
			// 'init.templateDir'.
			answer, err := promptCtx.ShowPromptOptions(
				"Do you want to set this up as the Git template\n"+
					"directory (e.g setting 'init.templateDir')\n"+
					"for future use?",
				"(yes, No (abort))",
				"y/N",
				"Yes", "No (abort)")
			log.AssertNoErrorF(err, "Could not show prompt.")

			log.PanicIf(answer != "y",
				"Could not determine Git hook",
				"templates directory. -> Abort.")

			return path.Join(templateDir, "hooks"), templateDir
		}
	}

	// 6. Set up as new
	answer, err = promptCtx.ShowPromptOptions(
		"Do you want to set up a new Git templates folder?",
		"(yes, No)",
		"y/N",
		"Yes", "No")
	log.AssertNoErrorF(err, "Could not show prompt.")

	if answer == "y" {
		templateDir := setupNewTemplateDir(installDir, promptCtx)
		return path.Join(templateDir, "hooks"), templateDir // nolint:nlreturn
	}

	return "", ""
}

func searchPreCommitFile(startDirs []string, promptCtx prompt.IContext) (result string) {

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

			answer, err := promptCtx.ShowPromptOptions(
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

func searchTemplateDirOnDisk(promptCtx prompt.IContext) string {

	first, second := GetDefaultTemplateSearchDir()

	templateDir := searchPreCommitFile(first, promptCtx)

	if strs.IsEmpty(templateDir) {

		answer, err := promptCtx.ShowPromptOptions(
			"Git hook template directory not found\n"+
				"Do you want to keep searching?",
			"(yes, No)",
			"y/N",
			"Yes", "No")

		log.AssertNoErrorF(err, "Could not show prompt.")

		if answer == "y" {
			templateDir = searchPreCommitFile(second, promptCtx)
		}
	}

	return templateDir
}

func setupNewTemplateDir(installDir string, promptCtx prompt.IContext) string {
	templateDir := path.Join(installDir, "templates")

	homeDir, err := homedir.Dir()
	cm.AssertNoErrorPanic(err, "Could not get home directory.")

	if promptCtx != nil {
		var err error
		templateDir, err = promptCtx.ShowPrompt(
			"Enter the target folder",
			templateDir,
			nil)
		log.AssertNoErrorF(err, "Could not show prompt.")
	}

	templateDir = cm.ReplaceTildeWith(templateDir, homeDir)
	log.AssertNoErrorPanicF(err, "Could not replace tilde '~' in '%s'.", templateDir)

	return templateDir
}

func getTargetTemplateDir(
	installDir string,
	templateDir string,
	useCoreHooksPath bool,
	nonInteractive bool,
	dryRun bool,
	promptCtx prompt.IContext) (hookTemplateDir string) {

	if strs.IsEmpty(templateDir) {
		// Automatically find a template directory.
		hookTemplateDir, templateDir = findGitHookTemplates(
			installDir,
			useCoreHooksPath,
			nonInteractive,
			promptCtx)

		log.PanicIfF(strs.IsEmpty(hookTemplateDir),
			"Could not determine Git hook template directory.")
	} else {
		// The user provided a template directory, check it and
		// add `hooks` which is needed.
		log.PanicIfF(!cm.IsDirectory(templateDir),
			"Given template dir '%s' does not exist.", templateDir)
		hookTemplateDir = path.Join(templateDir, "hooks")
	}

	log.InfoF("Hook template dir set to '%s'.", hookTemplateDir)

	err := os.MkdirAll(hookTemplateDir, cm.DefaultFileModeDirectory)
	log.AssertNoErrorPanicF(err,
		"Could not assert directory '%s' exists",
		hookTemplateDir)

	// Set the global Git configuration
	if useCoreHooksPath {
		setGithooksDirectory(true, hookTemplateDir, dryRun)
	} else {
		setGithooksDirectory(false, templateDir, dryRun)
	}

	return
}

func setGithooksDirectory(useCoreHooksPath bool, directory string, dryRun bool) {
	gitx := git.Ctx()

	prefix := "Setting"
	if dryRun {
		prefix = "[dry run] Would set"
	}

	if useCoreHooksPath {

		log.InfoF("%s 'core.hooksPath' to '%s'.", prefix, directory)

		if !dryRun {
			err := gitx.SetConfig(hooks.GitCK_UseCoreHooksPath, true, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")

			err = gitx.SetConfig(hooks.GitCK_PathForUseCoreHooksPath, directory, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")

			err = gitx.SetConfig(git.GitCK_CoreHooksPath, directory, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")
		}

		// Warnings:
		// Check if hooks might not run...
		tD := gitx.GetConfig(git.GitCK_InitTemplateDir, git.GlobalScope)
		msg := ""
		if strs.IsNotEmpty(tD) && cm.IsDirectory(path.Join(tD, "hooks")) {
			d := path.Join(tD, "hooks")
			files, err := cm.GetAllFiles(d)
			log.AssertNoErrorPanicF(err, "Could not get files in '%s'.", d)

			if len(files) > 0 {
				msg = strs.Fmt(
					"The 'init.templateDir' setting is currently set to\n"+
						"'%s'\n"+
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

		log.WarnIf(strs.IsNotEmpty(msg),
			msg+
				"These hooks might get installed but\n"+
				"ignored because 'core.hooksPath' is also set.\n"+
				"It is recommended to either remove the files or run\n"+
				"the Githooks installation without the '--use-core-hookspath'\n"+
				"parameter.")

	} else {

		if !dryRun {
			err := gitx.SetConfig(hooks.GitCK_UseCoreHooksPath, false, git.GlobalScope)
			log.AssertNoErrorPanic(err, "Could not set Git config value.")
		}

		if strs.IsNotEmpty(directory) {
			log.InfoF("%s 'init.templateDir' to '%s'.", prefix, directory)

			if !dryRun {
				err := gitx.SetConfig(git.GitCK_InitTemplateDir, directory, git.GlobalScope)
				log.AssertNoErrorPanic(err, "Could not set Git config value.")
			}
		}

		// Warnings:
		// Check if hooks might not run..
		hP := gitx.GetConfig(git.GitCK_CoreHooksPath, git.GlobalScope)
		log.WarnIfF(strs.IsNotEmpty(hP),
			"The 'core.hooksPath' setting is currently set to\n"+
				"'%s'\n"+
				"This could mean that Githooks hooks will be ignored\n"+
				"Either unset 'core.hooksPath' or run the Githooks\n"+
				"installation with the '--use-core-hookspath' parameter.",
			hP)

	}
}

func setupHookTemplates(
	hookTemplateDir string,
	cloneDir string,
	tempDir string,
	onlyServerHooks bool,
	nonInteractive bool,
	dryRun bool,
	uiSettings *install.UISettings) {

	if dryRun {
		log.InfoF("[dry run] Would install Git hook templates into '%s'.",
			hookTemplateDir)
		return // nolint:nlreturn
	}

	log.InfoF("Installing Git hook templates into '%s'.",
		hookTemplateDir)

	var hookNames []string
	if onlyServerHooks {
		hookNames = hooks.ManagedServerHookNames
	} else {
		hookNames = hooks.ManagedHookNames
	}

	log.InfoF("Saving Githooks run wrapper to '%s' :", hookTemplateDir)

	err := hooks.InstallRunWrappers(
		hookTemplateDir,
		hookNames,
		func(dest string) {
			log.InfoF(" %s '%s'", cm.ListItemLiteral, path.Base(dest))
		},
		install.GetHookDisableCallback(log, nonInteractive, uiSettings),
		log)

	log.AssertNoErrorPanicF(err, "Could not install run wrappers into '%s'.", hookTemplateDir)

	if onlyServerHooks {
		err := git.Ctx().SetConfig(hooks.GitCK_MaintainOnlyServerHooks, true, git.GlobalScope)
		log.AssertNoErrorPanic(err, "Could not set Git config 'githooks.maintainOnlyServerHooks'.")
	}
}

func installBinaries(
	installDir string,
	cloneDir string,
	tempDir string,
	binaries []string,
	dryRun bool) {

	if len(args.InternalBinaries) == 0 {
		return
	}

	binDir := hooks.GetBinaryDir(installDir)
	err := os.MkdirAll(binDir, cm.DefaultFileModeDirectory)
	log.AssertNoErrorPanicF(err, "Could not create binary dir '%s'.", binDir)

	msg := strs.Map(binaries, func(s string) string { return strs.Fmt(" • '%s'", path.Base(s)) })
	if dryRun {
		log.InfoF("[dry run] Would install binaries:\n%s\n"+"to '%s'.", msg, binDir)
		return // nolint:nlreturn
	}

	log.InfoF("Installing binaries:\n%s\n"+"to '%s'.", strings.Join(msg, "\n"), binDir)

	for _, binary := range binaries {
		dest := path.Join(binDir, path.Base(binary))
		err := cm.CopyFileWithBackup(binary, dest, tempDir, false)
		log.AssertNoErrorPanicF(err,
			"Could not move file '%s' to '%s'.", binary, dest)
	}

	// Set CLI executable alias.
	cliTool := hooks.GetCLIExecutable(installDir)
	err = hooks.SetCLIExecutableAlias(cliTool)
	log.AssertNoErrorPanicF(err,
		"Could not set Git config 'alias.hooks' to '%s'.", cliTool)

	// Set runner executable alias.
	runner := hooks.GetRunnerExecutable(installDir)
	err = hooks.SetRunnerExecutableAlias(runner)
	log.AssertNoErrorPanic(err,
		"Could not set runner executable alias '%s'.", runner)
}

func setupAutomaticUpdate(nonInteractive bool, dryRun bool, promptCtx prompt.IContext) {

	enabled, isSet := updates.GetAutomaticUpdateCheckSettings()
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

	activate := false

	if nonInteractive {
		activate = true
	} else {
		answer, err := promptCtx.ShowPromptOptions(
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

			err := updates.SetAutomaticUpdateCheckSettings(true, false)
			if log.AssertNoErrorF(err, "Failed to enable automatic update checks.") {
				log.Info("Automatic update checks are now enabled.")
			}
		}
	} else {
		log.Info(
			"If you change your mind in the future, you can enable it by running:",
			"  $ git hooks update enable")
	}
}

func installIntoExistingRepos(
	tempDir string,
	nonInteractive bool,
	dryRun bool,
	installedRepos InstallSet,
	registeredRepos *hooks.RegisterRepos,
	uiSettings *install.UISettings) {

	// Show prompt and run callback.
	install.PromptExistingRepos(
		log,
		nonInteractive,
		false,
		uiSettings.PromptCtx,

		func(gitDir string) {

			if install.InstallIntoRepo(
				log, gitDir,
				nonInteractive, dryRun, uiSettings) {

				registeredRepos.Insert(gitDir)
				installedRepos.Insert(gitDir)
			}
		})

}

func installIntoRegisteredRepos(
	tempDir string,
	nonInteractive bool,
	dryRun bool,
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
				log, gitDir,
				nonInteractive, dryRun, uiSettings) {

				registeredRepos.Insert(gitDir)
				installedRepos.Insert(gitDir)
			}
		})

}

func setupSharedRepositories(installDir string, cliExectuable string, dryRun bool, uiSettings *install.UISettings) {

	gitx := git.Ctx()
	sharedRepos := gitx.GetConfigAll(hooks.GitCK_Shared, git.GlobalScope)

	var question string
	if len(sharedRepos) != 0 {
		question = "Looks like you already have shared hook\n" +
			"repositories setup, do you want to change them now?"
	} else {
		question = "You can set up shared hook repositories to avoid\n" +
			"duplicating common hooks across repositories you work on\n" +
			"See information on what are these in the project's documentation:\n" +
			strs.Fmt("'%s#shared-hook-repositories'\n", hooks.GithooksWebpage) +
			strs.Fmt("Note: you can also have a '%s' file listing the\n", hooks.GetRepoSharedFileRel()) +
			"      repositories where you keep the shared hook files.\n" +
			"Would you like to set up shared hook repos now?"
	}

	answer, err := uiSettings.PromptCtx.ShowPromptOptions(
		question,
		"(yes, No)", "y/N", "Yes", "No")
	log.AssertNoError(err, "Could not show prompt")

	if answer == "n" {
		return
	}

	log.Info("Let's input shared hook repository urls",
		"one-by-one and leave the input empty to stop.")

	entries, err := uiSettings.PromptCtx.ShowPromptMulti(
		"Enter the clone URL of a shared repository",
		prompt.ValidatorAnswerNotEmpty)

	if err != nil {
		log.Error("Could not show prompt. Not settings shared hook repositories.")
		return // nolint: nlreturn
	}

	// Unset all shared configs.
	err = gitx.UnsetConfig(hooks.GitCK_Shared, git.GlobalScope)
	log.AssertNoErrorF(err,
		"Could not unset Git config '%s'.\n"+
			"Failed to setup shared hook repositories.", hooks.GitCK_Shared)
	if err != nil {
		return
	}

	// Add all entries.
	for _, entry := range entries {
		err := gitx.AddConfig(hooks.GitCK_Shared, entry, git.GlobalScope)
		log.AssertNoError(err,
			"Could not add Git config '%s'.\n"+
				"Failed to setup shared hook repositories.", hooks.GitCK_Shared)
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
			"file will still be executed", hooks.GitCK_Shared, hooks.GetRepoSharedFileRel())
	} else {

		updated, err := hooks.UpdateAllSharedHooks(log, gitx, installDir, "")
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

func storeSettings(settings *Settings, uiSettings *install.UISettings) {
	// Store cached UI values back.
	err := git.Ctx().SetConfig(
		hooks.GitCK_DeleteDetectedLFSHooksAnswer, uiSettings.DeleteDetectedLFSHooks, git.GlobalScope)
	log.AssertNoError(err, "Could not store config 'githooks.deleteDetectedLFSHooks'.")

	err = settings.RegisteredGitDirs.Store(settings.InstallDir)
	log.AssertNoError(err,
		"Could not store registered file in '%s'.",
		settings.InstallDir)

	if err != nil {
		for _, gitDir := range settings.InstalledGitDirs.ToList() {
			// For each installedGitDir entry, mark the repository as registered.
			err := hooks.MarkRepoRegistered(git.CtxC(gitDir))
			log.AssertNoErrorF(err, "Could not mark Git directory '%s' as registered.", gitDir)
		}
	}

}

func updateClone(cloneDir string, updateToSHA string) {

	if strs.IsEmpty(updateToSHA) {
		return // We don't need to update the relase clone.
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

func thankYou() {
	log.InfoF("All done! Enjoy!\n"+
		"Please support the project by starring the project\n"+
		"at '%s', and report\n"+
		"bugs or missing features or improvements as issues.\n"+
		"Thanks!\n", hooks.GithooksWebpage)
}

func runUpdate(
	settings *Settings,
	uiSettings *install.UISettings,
	args *Arguments) {

	log.InfoF("Running install to version '%s' ...", build.BuildVersion)

	// Read registered file if existing.
	// We ensured during load, that only existing Git directories are listed.
	err := settings.RegisteredGitDirs.Load(settings.InstallDir, true, true)
	log.AssertNoErrorPanicF(err, "Could not load register file in '%s'.", settings.InstallDir)

	settings.HookTemplateDir = getTargetTemplateDir(
		settings.InstallDir,
		args.TemplateDir,
		args.UseCoreHooksPath,
		args.NonInteractive,
		args.DryRun,
		uiSettings.PromptCtx)

	installBinaries(
		settings.InstallDir,
		settings.CloneDir,
		settings.TempDir,
		args.InternalBinaries,
		args.DryRun)

	setupHookTemplates(
		settings.HookTemplateDir,
		settings.CloneDir,
		settings.TempDir,
		args.OnlyServerHooks,
		args.NonInteractive,
		args.DryRun,
		uiSettings)

	if !args.InternalAutoUpdate {
		setupAutomaticUpdate(args.NonInteractive, args.DryRun, uiSettings.PromptCtx)
	}

	if !args.SkipInstallIntoExisting && !args.UseCoreHooksPath {

		if !args.InternalAutoUpdate {
			installIntoExistingRepos(
				settings.TempDir,
				args.NonInteractive,
				args.DryRun,
				settings.InstalledGitDirs,
				&settings.RegisteredGitDirs,
				uiSettings)
		}
	}

	if !args.UseCoreHooksPath {
		installIntoRegisteredRepos(
			settings.TempDir,
			args.NonInteractive,
			args.DryRun,
			settings.InstalledGitDirs,
			&settings.RegisteredGitDirs,
			uiSettings)
	}

	if !args.InternalAutoUpdate && !args.NonInteractive {
		setupSharedRepositories(
			settings.InstallDir,
			hooks.GetCLIExecutable(settings.InstallDir),
			args.DryRun,
			uiSettings)
	}

	if !args.DryRun {
		storeSettings(settings, uiSettings)
		updateClone(settings.CloneDir, args.InternalUpdateTo)
	}

	thankYou()
}

func runInstall(cmd *cobra.Command, auxArgs []string) {

	log.InfoF("Githooks Installer [version: %s]", build.BuildVersion)

	log.DebugF("Arguments: %+v", args)
	validateArgs(cmd, &args)

	settings, uiSettings := setMainVariables(&args)

	if !args.DryRun {
		setInstallDir(settings.InstallDir)
	}

	if !args.InternalPostDispatch {
		if isDispatched := prepareDispatch(&settings, &args); isDispatched {
			return
		}
	}

	runUpdate(&settings, &uiSettings, &args)
}

func main() {

	cwd, err := os.Getwd()
	cm.AssertNoErrorPanic(err, "Could not get current working dir.")
	cwd = filepath.ToSlash(cwd)

	log, err = cm.CreateLogContext(cm.IsRunInDocker)
	cm.AssertOrPanic(err == nil, "Could not create log")

	exitCode := 0
	defer func() { os.Exit(exitCode) }()

	// Handle all panics and report the error
	defer func() {
		r := recover()
		if hooks.HandleCLIErrors(r, cwd, log) {
			exitCode = 1
		}
	}()

	Run()
}
