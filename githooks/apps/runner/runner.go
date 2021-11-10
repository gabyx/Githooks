//go:generate go run -mod=vendor ../../tools/generate-version.go
package main

import (
	"github.com/gabyx/githooks/githooks/build"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	"github.com/gabyx/githooks/githooks/prompt"
	strs "github.com/gabyx/githooks/githooks/strings"
	"github.com/gabyx/githooks/githooks/updates"

	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/pbenner/threadpool"
	"github.com/pkg/math"
)

var log cm.ILogContext

func main() {
	os.Exit(mainRun())
}

func mainRun() (exitCode int) {

	createLog()

	log.DebugF("Githooks Runner [version: %s]", build.BuildVersion)

	startTime := cm.GetStartTime()

	defer func() {
		log.DebugF("Runner execution time: '%v'.",
			cm.GetDuration(startTime))
	}()

	cwd, err := os.Getwd()
	cm.AssertNoErrorPanic(err, "Could not get current working dir.")
	cwd = filepath.ToSlash(cwd)

	// Handle all panics and report the error
	defer func() {
		r := recover()
		if cm.HandleCLIErrors(r, cwd, log, hooks.GetBugReportingInfo) {
			exitCode = 1
		}
	}()

	settings, uiSettings := setMainVariables(cwd)
	assertRegistered(settings.GitX, settings.InstallDir)

	checksums, err := hooks.GetChecksumStorage(settings.GitX, settings.GitDirWorktree)
	log.AssertNoErrorF(err, "Errors while loading checksum store.")
	log.DebugF("%s", checksums.Summary())

	// Set this repositories hook namespace.
	ns, err := hooks.GetHooksNamespace(settings.RepositoryHooksDir)
	log.AssertNoErrorF(err, "Errors while loading hook namespace.")
	if strs.IsNotEmpty(ns) {
		settings.HookNamespace = ns
	}

	ignores, err := hooks.GetIgnorePatterns(
		settings.RepositoryHooksDir,
		settings.GitDirWorktree,
		[]string{settings.HookName},
		settings.HookNamespace)
	log.AssertNoErrorF(err, "Errors while loading ignore patterns.")
	log.DebugF("User ignore patterns: '%+q'.", ignores.User)
	log.DebugF("Accumuldated repository ignore patterns: '%q'.", ignores.HooksDir)

	defer storePendingData(&settings, &uiSettings, &ignores, &checksums)

	if hooks.IsGithooksDisabled(settings.GitX, true) {
		executeLFSHooks(&settings)
		executeOldHook(&settings, &uiSettings, &ignores, &checksums)

		return
	}

	exportGeneralEnvVars()
	exportStagedFiles(&settings)
	updateGithooks(&settings, &uiSettings)
	executeLFSHooks(&settings)
	executeOldHook(&settings, &uiSettings, &ignores, &checksums)

	hooks := collectHooks(&settings, &uiSettings, &ignores, &checksums)

	if cm.IsDebug {
		logBatches("Local Hooks", hooks.LocalHooks)
		logBatches("Repo Shared Hooks", hooks.RepoSharedHooks)
		logBatches("Local Shared Hooks", hooks.LocalSharedHooks)
		logBatches("Global Shared Hooks", hooks.GlobalSharedHooks)
	}

	executeHooks(&settings, &hooks)

	uiSettings.PromptCtx.Close()
	log.Debug("All done.\n")

	return
}

func createLog() {
	var err error
	// Its good to output everythin to stderr since git
	// might read stdin for certain hooks.
	// Either do redirection (which needs to be bombproof)
	// or just use stderr.
	log, err = cm.CreateLogContext(true)
	cm.AssertOrPanic(err == nil, "Could not create log")
}

func logInvocation(s *HookSettings) {
	if strs.IsNotEmpty(os.Getenv("GITHOOKS_RUNNER_TRACE")) {
		log.InfoF("Running hooks for: '%s' %q", s.HookName, s.Args)
	}
}

func setMainVariables(repoPath string) (HookSettings, UISettings) {

	cm.PanicIf(
		len(os.Args) <= 1,
		"No arguments given! -> Abort")

	// General execution context, in currenct working dir.
	execx := cm.ExecContext{}

	// Current git context, in current working dir.
	gitx := git.Ctx()

	gitDir, err := gitx.GetGitDirWorktree()
	cm.AssertNoErrorPanic(err, "Could not get git directory.")

	err = hooks.DeleteHookDirTemp(path.Join(gitDir, "hooks"))
	log.AssertNoErrorF(err, "Could not clean temporary directory in '%s/hooks'.", gitDir)

	hookPath, err := filepath.Abs(os.Args[1])
	cm.AssertNoErrorPanicF(err, "Could not abs. path from '%s'.",
		os.Args[1])
	hookPath = filepath.ToSlash(hookPath)

	installDir := getInstallDir(gitx)

	promptx, err := prompt.CreateContext(log, true, false)
	log.DebugIfF(err != nil, "Prompt setup failed -> using fallback.")

	nonInteractive := hooks.IsRunnerNonInteractive(gitx, git.Traverse)
	skipNonExistingSharedHooks, _ := hooks.SkipNonExistingSharedHooks(gitx, git.Traverse)
	skipUntrustedHooks, _ := hooks.SkipUntrustedHooks(gitx, git.Traverse)

	isTrusted, hasTrustFile := hooks.IsRepoTrusted(gitx, repoPath)
	if !isTrusted && hasTrustFile && !nonInteractive {
		isTrusted = showTrustRepoPrompt(gitx, promptx)
	}

	s := HookSettings{
		Args:               os.Args[2:],
		ExecX:              execx,
		GitX:               gitx,
		RepositoryDir:      repoPath,
		RepositoryHooksDir: path.Join(repoPath, hooks.HooksDirName),
		GitDirWorktree:     gitDir,
		InstallDir:         installDir,

		HookPath:      hookPath,
		HookName:      path.Base(hookPath),
		HookDir:       path.Dir(hookPath),
		HookNamespace: hooks.NamespaceRepositoryHook,
		IsRepoTrusted: isTrusted,

		SkipNonExistingSharedHooks: skipNonExistingSharedHooks,
		SkipUntrustedHooks:         skipUntrustedHooks,
		NonInteractive:             nonInteractive}

	log.DebugF(s.toString())
	logInvocation(&s)

	return s, UISettings{AcceptAllChanges: false, PromptCtx: promptx}
}

func getInstallDir(gitx *git.Context) string {
	installDir := hooks.GetInstallDir(gitx)

	setDefault := func() {
		usr, err := homedir.Dir()
		cm.AssertNoErrorPanic(err, "Could not get home directory.")
		usr = filepath.ToSlash(usr)
		installDir = path.Join(usr, hooks.HooksDirName)
	}

	if strs.IsEmpty(installDir) {
		setDefault()
	} else if exists, err := cm.IsPathExisting(installDir); !exists {

		log.AssertNoError(err,
			"Could not check path '%s'", installDir)
		log.WarnF(
			"Githooks installation is corrupt!\n"+
				"Install directory at '%s' is missing.",
			installDir)

		setDefault()

		log.WarnF(
			"Falling back to default directory at '%s'.\n"+
				"Please run the Githooks install script again to fix it.",
			installDir)
	}

	log.Debug(strs.Fmt("Install dir set to: '%s'.", installDir))

	return installDir
}

func assertRegistered(gitx *git.Context, installDir string) {

	if !gitx.IsConfigSet(hooks.GitCKRegistered, git.LocalScope) &&
		!gitx.IsConfigSet(git.GitCKCoreHooksPath, git.Traverse) {

		gitDir, err := gitx.GetGitDirCommon()
		log.AssertNoErrorPanicF(err, "Could not get Git common dir.")

		log.DebugF("Register repo '%s'", gitDir)
		err = hooks.RegisterRepo(gitDir, installDir, true, false)
		log.AssertNoErrorF(err, "Could not register repo '%s'.", gitDir)

		err = hooks.MarkRepoRegistered(gitx)
		log.AssertNoErrorF(err, "Could not set register flag in repo '%s'.", gitDir)

	} else {
		log.Debug(
			"Repository already registered or using 'core.hooksPath'.")
	}
}

func showTrustRepoPrompt(gitx *git.Context, promptx prompt.IContext) (isTrusted bool) {
	question := "This repository wants you to trust all current and\n" +
		"future hooks without prompting.\n" +
		"Do you want to allow running every current and future hooks?"

	var answer string
	answer, err := promptx.ShowOptions(question, "(yes, no)", "y/n", "Yes", "No")
	log.AssertNoErrorF(err, "Could not get trust prompt answer.")
	if err != nil {
		return
	}

	if answer == "y" {
		err := hooks.SetTrustAllSetting(gitx, true, false)
		log.AssertNoErrorF(err, "Could not store trust setting.")
		isTrusted = true
	} else {
		err := hooks.SetTrustAllSetting(gitx, false, false)
		log.AssertNoErrorF(err, "Could not store trust setting.")
	}

	return
}

func exportGeneralEnvVars() {
	os.Setenv(hooks.EnvVariableOs, runtime.GOOS)
	os.Setenv(hooks.EnvVariableArch, runtime.GOARCH)
}

func exportStagedFiles(settings *HookSettings) {
	if strs.Includes(hooks.StagedFilesHookNames[:], settings.HookName) {

		files, err := hooks.GetStagedFiles(settings.GitX)

		if len(files) != 0 {
			log.DebugF("Exporting staged files:\n- %s",
				strings.ReplaceAll(files, "\n", "\n- "))
		}

		if err != nil {
			log.Warn("Could not export staged files.")
		} else {

			cm.DebugAssertF(
				func() bool {
					_, exists := os.LookupEnv(hooks.EnvVariableStagedFiles)
					return !exists // nolint:nlreturn
				}(),
				"Env. variable '%s' already defined.", hooks.EnvVariableStagedFiles)

			os.Setenv(hooks.EnvVariableStagedFiles, files)
		}

	}
}

func updateGithooks(settings *HookSettings, uiSettings *UISettings) {

	if !shouldRunUpdateCheck(settings) {
		return
	}

	opts := []string{"--internal-auto-update"}
	if settings.NonInteractive {
		opts = append(opts, "--non-interactive")
	}

	updateAvailable, accepted, err := updates.RunUpdate(
		settings.InstallDir,
		updates.DefaultAcceptUpdateCallback(log, uiSettings.PromptCtx, updates.AcceptNonInteractiveNone),
		func() error {
			return updates.RunUpdateOverExecutable(settings.InstallDir,
				&settings.ExecX,
				cm.UseStreams(nil, log.GetInfoWriter(), log.GetErrorWriter()),
				opts...)
		})

	log.AssertNoErrorPanic(err, "Running update failed.")

	switch {
	case updateAvailable:
		if accepted {
			log.Info("Update successfully dispatched.")
		} else {
			log.Info("Update declined.")
		}
	default:
		log.InfoF("Githooks is at the latest version '%s'",
			build.GetBuildVersion().String())
	}

	log.Info(
		"If you would like to disable auto-updates, run:",
		"  $ git hooks update disable")
}

func shouldRunUpdateCheck(settings *HookSettings) bool {
	if settings.HookName != "post-commit" {
		return false
	}

	enabled, _ := updates.GetAutomaticUpdateCheckSettings(settings.GitX)
	if !enabled {
		return false
	}

	lastUpdateCheck, _, err := updates.GetUpdateCheckTimestamp(settings.GitX)
	log.AssertNoErrorF(err, "Could get last update check time.")

	return time.Since(lastUpdateCheck).Hours() > 24.0 //nolint: gomnd
}

func executeLFSHooks(settings *HookSettings) {

	if !strs.Includes(hooks.LFSHookNames[:], settings.HookName) {
		return
	}

	lfsIsAvailable := git.IsLFSAvailable()
	lfsRequiredFile, lfsReqFileExists := hooks.GetLFSRequiredFile(settings.RepositoryDir)
	lfsConfFile, lfsConfExists := git.GetLFSConfigFile(settings.RepositoryDir)
	lfsIsRequired := lfsReqFileExists || lfsConfExists

	if lfsIsAvailable {
		log.Debug("Excuting LFS Hook")

		err := settings.GitX.CheckPiped(
			append(
				[]string{"lfs", settings.HookName},
				settings.Args...,
			)...)

		log.AssertNoErrorPanic(err, "Execution of LFS Hook failed.")

	} else if lfsIsRequired {
		log.PanicF("This repository requires Git LFS, but 'git-lfs' was\n"+
			"not found on your PATH.\n"+
			"Git LFS is required since one of the following is true:\n"+
			"  - file '%s' existing: '%v'\n"+
			"  - file `%s` existing: '%v'",
			lfsConfFile, lfsConfExists, lfsRequiredFile, lfsReqFileExists)
	}
}

func executeOldHook(
	settings *HookSettings,
	uiSettings *UISettings,
	ignores *hooks.RepoIgnorePatterns,
	checksums *hooks.ChecksumStore) {

	// e.g. 'hooks/pre-commit.replaced.githook's
	hookName := hooks.GetHookReplacementFileName(settings.HookName)
	hookNamespace := hooks.NamespaceReplacedHook

	// Old hook can only be ignored by user ignores...
	isIgnored := func(namespacePath string) bool {
		ignored, byUser := ignores.IsIgnored(namespacePath)

		return ignored && byUser
	}

	isTrusted := func(hookPath string) (bool, string) {
		if settings.IsRepoTrusted {
			return true, ""
		}

		trusted, sha, e := checksums.IsTrusted(hookPath)
		log.AssertNoErrorPanicF(e, "Could not check trust status '%s'.", hookPath)

		return trusted, sha
	}

	hooks, _, err := hooks.GetAllHooksIn(
		settings.RepositoryDir,
		settings.HookDir, hookName, hookNamespace,
		isIgnored, isTrusted, true, false)
	log.AssertNoErrorPanicF(err, "Errors while collecting hooks in '%s'.", settings.HookDir)

	if len(hooks) == 0 {
		log.DebugF("Old hook '%s' does not exist. -> Skip!", hookName)

		return
	}

	hook := hooks[0]

	if hook.Active && !hook.Trusted {
		if !settings.NonInteractive {
			// Active hook, but not trusted:
			// Show trust prompt to let user trust it or disable it.
			showTrustPrompt(uiSettings, checksums, &hook)
		}

		log.PanicIfF(!settings.SkipUntrustedHooks && hook.Active && !hook.Trusted,
			"Hook '%s' is active and needs to be trusted first.\n"+
				"Either trust the hook or disable it, or skip active,\n"+
				"untrusted hooks by running:\n"+
				"  $ git hooks config skip-untrusted-hooks --enable",
			hook.NamespacePath)
	}

	if !hook.Active || !hook.Trusted {
		log.DebugF("Hook '%s' is skipped [active: '%v', trusted: '%v']",
			hook.Path, hook.Active, hook.Trusted)

		return
	}

	log.DebugF("Executing hook: '%s'.", hook.Path)
	err = cm.RunExecutable(&settings.ExecX, &hook, cm.UseStdStreams(true, true, true))

	log.AssertNoErrorPanicF(err, "Hook launch failed: '%q'.", hook)
}

func collectHooks(
	settings *HookSettings,
	uiSettings *UISettings,
	ignores *hooks.RepoIgnorePatterns,
	checksums *hooks.ChecksumStore) (h hooks.Hooks) {

	// Local hooks in repository
	// No parsing of local includes because already happened.
	h.LocalHooks = getHooksIn(
		settings, uiSettings, settings.RepositoryDir, settings.RepositoryHooksDir,
		false, settings.HookNamespace, false, ignores, checksums)

	// All shared hooks
	var allAddedShared = make([]string, 0)
	h.RepoSharedHooks = getRepoSharedHooks(settings, uiSettings, ignores, checksums, &allAddedShared)

	h.LocalSharedHooks = getConfigSharedHooks(
		settings,
		uiSettings,
		ignores,
		checksums,
		&allAddedShared,
		hooks.SharedHookTypeV.Local)

	h.GlobalSharedHooks = getConfigSharedHooks(
		settings,
		uiSettings,
		ignores,
		checksums,
		&allAddedShared,
		hooks.SharedHookTypeV.Global)

	return
}

func updateSharedHooks(settings *HookSettings, sharedHooks []hooks.SharedRepo, sharedType hooks.SharedHookType) {

	disableUpdate, _ := hooks.IsSharedHooksUpdateDisabled(settings.GitX, git.Traverse)
	updateTriggers := settings.GitX.GetConfigAll(hooks.GitCKSharedUpdateTriggers, git.Traverse)

	triggered := settings.HookName == "post-merge" ||
		(settings.HookName == "post-checkout" &&
			settings.Args[0] == git.NullRef) ||
		strs.Includes(updateTriggers, settings.HookName)

	if disableUpdate || !triggered {
		log.Debug("Shared hooks not updated.")

		return
	}

	log.Debug("Updating all shared hooks.")
	_, err := hooks.UpdateSharedHooks(log, sharedHooks, sharedType)
	log.AssertNoError(err, "Errors while updating shared hooks repositories.")
}

func getRepoSharedHooks(
	settings *HookSettings,
	uiSettings *UISettings,
	ignores *hooks.RepoIgnorePatterns,
	checksums *hooks.ChecksumStore,
	allAddedHooks *[]string) (hs hooks.HookPrioList) {

	shared, err := hooks.LoadRepoSharedHooks(settings.InstallDir, settings.RepositoryDir)

	if err != nil {
		log.ErrorOrPanicF(!settings.SkipNonExistingSharedHooks, err,
			"Repository shared hooks are demanded but failed "+
				"to parse the file:\n'%s'",
			hooks.GetRepoSharedFile(settings.RepositoryDir))
	}

	updateSharedHooks(settings, shared, hooks.SharedHookTypeV.Repo)

	for i := range shared {
		shRepo := &shared[i]

		if checkSharedHook(settings, shRepo, allAddedHooks, hooks.SharedHookTypeV.Repo) {
			hs = append(hs, getHooksInShared(settings, uiSettings, shRepo, ignores, checksums)...)
			*allAddedHooks = append(*allAddedHooks, shRepo.RepositoryDir)
		}
	}

	return
}

func getConfigSharedHooks(
	settings *HookSettings,
	uiSettings *UISettings,
	ignores *hooks.RepoIgnorePatterns,
	checksums *hooks.ChecksumStore,
	allAddedHooks *[]string,
	sharedType hooks.SharedHookType) (hs hooks.HookPrioList) {

	var shared []hooks.SharedRepo
	var err error

	switch sharedType {
	case hooks.SharedHookTypeV.Local:
		shared, err = hooks.LoadConfigSharedHooks(settings.InstallDir, settings.GitX, git.LocalScope)
	case hooks.SharedHookTypeV.Global:
		shared, err = hooks.LoadConfigSharedHooks(settings.InstallDir, settings.GitX, git.GlobalScope)
	default:
		cm.DebugAssertF(false, "Wrong shared type '%v'", sharedType)
	}

	if err != nil {
		log.ErrorOrPanicF(!settings.SkipNonExistingSharedHooks,
			err,
			"Shared hooks are demanded but failed "+
				"to parse the %s config:\n'%s'",
			hooks.GetSharedHookTypeString(sharedType),
			hooks.GitCKShared)
	}

	for i := range shared {
		shRepo := &shared[i]

		if checkSharedHook(settings, shRepo, allAddedHooks, sharedType) {
			hs = append(hs, getHooksInShared(settings, uiSettings, shRepo, ignores, checksums)...)
			*allAddedHooks = append(*allAddedHooks, shRepo.RepositoryDir)
		}
	}

	return
}

func checkSharedHook(
	settings *HookSettings,
	hook *hooks.SharedRepo,
	allAddedHooks *[]string,
	sharedType hooks.SharedHookType) bool {

	if strs.Includes(*allAddedHooks, hook.RepositoryDir) {
		log.WarnF(
			"Shared hooks entry:\n'%s'\n"+
				"is already listed and will be skipped.", hook.OriginalURL)

		return false
	}

	// Check that no local paths are in repository configured
	// shared hooks
	log.PanicIfF(!hooks.AllowLocalURLInRepoSharedHooks() &&
		sharedType == hooks.SharedHookTypeV.Repo && hook.IsLocal,
		"Shared hooks in '%[1]s' contain a local path\n"+
			"'%[2]s'\n"+
			"which is forbidden.\n"+
			"\n"+
			"You can only have local paths in shared hooks defined\n"+
			"in the local or global Git configuration.\n"+
			"\n"+
			"You need to fix this by running\n"+
			"  $ git hooks shared add [--local|--global] '%[2]s'\n"+
			"and deleting it from the '.shared' file by\n"+
			"  $ git hooks shared remove --shared '%[2]s'",
		hooks.GetRepoSharedFileRel(), hook.OriginalURL)

	// Check if existing otherwise skip or fail...
	exists, err := cm.IsPathExisting(hook.RepositoryDir)

	if !exists {

		mess := "Failed to execute shared hooks in:\n" +
			"'%s'\n"

		if hook.IsCloned {
			mess += "It is not available. To fix, run:\n" +
				"$ git hooks shared update"
		} else {
			mess += "It does not exist."
		}

		if settings.SkipNonExistingSharedHooks {
			mess += "\nContinuing..."
		}

		log.ErrorOrPanicF(!settings.SkipNonExistingSharedHooks, err, mess, hook.OriginalURL)

		return false
	}

	// If cloned check that the remote url
	// is the same as the specified
	// Note: GIT_DIR might be set (?bug?) (actually the case for post-checkout hook)
	if hook.IsCloned {
		url := git.CtxCSanitized(hook.RepositoryDir).GetConfig(
			"remote.origin.url", git.LocalScope)

		if url != hook.URL {
			mess := "Failed to execute shared hooks in '%s'\n" +
				"The remote URL '%s' is different.\n" +
				"To fix it, run:\n" +
				"  $ git hooks shared purge\n" +
				"  $ git hooks shared update"

			if settings.SkipNonExistingSharedHooks {
				mess += "\nContinuing..."
			}

			log.ErrorOrPanicF(!settings.SkipNonExistingSharedHooks,
				nil, mess, hook.OriginalURL, url)

			return false
		}
	}

	return true
}

func getHooksIn(
	settings *HookSettings,
	uiSettings *UISettings,
	rootDir string,
	hooksDir string,
	addInternalIgnores bool,
	hookNamespace string,
	readNamespace bool,
	ignores *hooks.RepoIgnorePatterns,
	checksums *hooks.ChecksumStore) (batches hooks.HookPrioList) {

	log.DebugF("Getting hooks in '%s'", hooksDir)

	isTrusted := func(hookPath string) (bool, string) {
		if settings.IsRepoTrusted {
			return true, ""
		}

		trusted, sha, e := checksums.IsTrusted(hookPath)
		log.AssertNoErrorPanicF(e, "Could not check trust status '%s'.", hookPath)

		return trusted, sha
	}

	// Determine namespace
	if readNamespace {
		ns, err := hooks.GetHooksNamespace(hooksDir)
		log.AssertNoErrorPanicF(err, "Could not get hook namespace in '%s'", hooksDir)
		if strs.IsNotEmpty(ns) {
			hookNamespace = ns
		}
	}

	var internalIgnores hooks.HookPatterns
	if addInternalIgnores {
		var e error
		internalIgnores, e = hooks.GetHookPatternsHooksDir(hooksDir, []string{settings.HookName}, hookNamespace)
		log.AssertNoErrorPanicF(e, "Could not get worktree ignores in '%s'.", hooksDir)
	}

	isIgnored := func(namespacePath string) bool {
		ignored, _ := ignores.IsIgnored(namespacePath)

		return ignored || internalIgnores.Matches(namespacePath)
	}

	allHooks, maxBatches, err := hooks.GetAllHooksIn(
		rootDir,
		hooksDir, settings.HookName, hookNamespace,
		isIgnored, isTrusted, true, true)
	log.AssertNoErrorPanicF(err, "Errors while collecting hooks in '%s'.", hooksDir)

	if len(allHooks) == 0 {
		return
	}

	// Sort allHooks by the given batchName
	if len(allHooks) > 1 {
		sort.Slice(allHooks, func(i, j int) bool {
			return allHooks[i].BatchName < allHooks[j].BatchName
		})
	}

	// Split all hooks (sorted by the batch names)
	// into batches.
	batches = make(hooks.HookPrioList, 1, maxBatches)

	curBatchIdx := 0
	curBatchName := &allHooks[0].BatchName

	for i := range allHooks {

		hook := &allHooks[i]

		if hook.Active && !hook.Trusted {

			if !settings.NonInteractive {
				// Active hook, but not trusted:
				// Show trust prompt to let user trust it or disable it.
				showTrustPrompt(uiSettings, checksums, hook)
			}

			log.PanicIfF(!settings.SkipUntrustedHooks && hook.Active && !hook.Trusted,
				"Hook '%s' is active and needs to be trusted first.\n"+
					"Either trust the hook or disable it, or skip active,\n"+
					"untrusted hooks by running:\n"+
					"  $ git hooks config skip-untrusted-hooks --enable",
				hook.NamespacePath)
		}

		if !hook.Active || !hook.Trusted {
			log.DebugF("Hook '%s' is skipped [active: '%v', trusted: '%v']",
				hook.Path, hook.Active, hook.Trusted)

			continue
		}

		if *curBatchName != hook.BatchName {
			// Batch name changed, add another batch...
			batches = append(batches, []hooks.Hook{})
			curBatchIdx++
			curBatchName = &hook.BatchName
		}

		batches[curBatchIdx] = append(batches[curBatchIdx], *hook)
	}

	return
}

func getHooksInShared(settings *HookSettings,
	uiSettings *UISettings,
	shRepo *hooks.SharedRepo,
	ignores *hooks.RepoIgnorePatterns,
	checksums *hooks.ChecksumStore) hooks.HookPrioList {

	hookNamespace := hooks.GetDefaultHooksNamespaceShared(shRepo)

	dir := hooks.GetSharedGithooksDir(shRepo.RepositoryDir)

	return getHooksIn(settings, uiSettings,
		shRepo.RepositoryDir, dir, true, hookNamespace, true, ignores, checksums)
}

func logBatches(title string, hooks hooks.HookPrioList) {
	var l string

	if hooks == nil {
		log.DebugF("%s: none", title)
	} else {
		for bIdx, batch := range hooks {
			l += strs.Fmt(" Batch: %v\n", bIdx)
			for i := range batch {
				l += strs.Fmt("  - '%s' %+q\n", batch[i].GetCommand(), batch[i].GetArgs())
			}
		}
		log.DebugF("%s :\n%s", title, l)
	}
}

func showTrustPrompt(
	uiSettings *UISettings,
	checksums *hooks.ChecksumStore,
	hook *hooks.Hook) {

	if hook.Trusted {
		return
	}

	mess := strs.Fmt("New or changed hook found:\n'%s'", hook.Path)

	acceptHook := uiSettings.AcceptAllChanges
	disableHook := false

	if !acceptHook {

		question := mess + "\nDo you accept the changes?"

		answer, err := uiSettings.PromptCtx.ShowOptions(question,
			"(yes, all, no, disable)",
			"y/a/n/d",
			"Yes", "All", "No", "Disable")
		log.AssertNoErrorPanic(err, "Could not get trust prompt answer.")

		switch answer {
		case "a":
			uiSettings.AcceptAllChanges = true
			fallthrough // nolint:nlreturn
		case "y":
			acceptHook = true
		case "d":
			disableHook = true
		default:
			// Don't run hook ...
		}
	} else {
		log.Info("-> Already accepted.")
	}

	if acceptHook || disableHook {
		err := hook.AssertSHA1()
		log.AssertNoError(err, "Could not compute SHA1 hash of '%s'.", hook.Path)
	}

	if acceptHook {
		hook.Trusted = true

		uiSettings.AppendTrustedHook(
			hooks.ChecksumResult{
				SHA1:          hook.SHA1,
				Path:          hook.Path,
				NamespacePath: hook.NamespacePath})

		checksums.AddChecksum(hook.SHA1, hook.Path)

	} else if disableHook {
		log.InfoF("-> Adding hook\n'%s'\nto disabled list.", hook.Path)

		hook.Active = false

		uiSettings.AppendDisabledHook(
			hooks.ChecksumResult{
				SHA1:          hook.SHA1,
				Path:          hook.Path,
				NamespacePath: hook.NamespacePath})
	}
}

func executeHooks(settings *HookSettings, hs *hooks.Hooks) {

	var nThreads = runtime.NumCPU()
	nThSetting := settings.GitX.GetConfig(hooks.GitCKNumThreads, git.Traverse)
	if n, err := strconv.Atoi(nThSetting); err == nil {
		nThreads = n
	}

	// Minimal 1 thread.
	nThreads = math.MaxInt(nThreads, 1)

	var pool *threadpool.ThreadPool
	if hooks.UseThreadPool && hs.GetHooksCount() > 1 {
		log.Debug("Launching with thread pool")
		p := threadpool.New(nThreads, 15) // nolint: gomnd
		pool = &p
	}

	var results []hooks.HookResult
	var err error

	// Dump execution sequence.
	if cm.IsDebug {
		file, err := os.CreateTemp("", strs.Fmt("*-githooks-prio-list-%s.json", settings.HookName))
		log.AssertNoErrorPanic(err, "Failed to create execution log.")
		defer file.Close()
		err = hs.StoreJSON(file)
		log.AssertNoErrorPanic(err, "Failed to create execution log.")
		log.DebugF("Hooks priority list written to '%s'.", file.Name())
	}

	log.DebugIf(len(hs.LocalHooks) != 0, "Launching local hooks ...")
	results, err = hooks.ExecuteHooksParallel(
		pool, &settings.ExecX, &hs.LocalHooks,
		results, logHookResults,
		settings.Args...)
	log.AssertNoErrorPanic(err, "Local hook execution failed.")

	log.DebugIf(len(hs.RepoSharedHooks) != 0, "Launching repository shared hooks ...")
	results, err = hooks.ExecuteHooksParallel(
		pool, &settings.ExecX, &hs.RepoSharedHooks,
		results, logHookResults,
		settings.Args...)
	log.AssertNoErrorPanic(err, "Shared repository hook execution failed.")

	log.DebugIf(len(hs.LocalSharedHooks) != 0, "Launching local shared hooks ...")
	results, err = hooks.ExecuteHooksParallel(
		pool, &settings.ExecX, &hs.LocalSharedHooks,
		results, logHookResults,
		settings.Args...)
	log.AssertNoErrorPanic(err, "Local shared hook execution failed.")

	log.DebugIf(len(hs.GlobalSharedHooks) != 0, "Launching global shared hooks ...")
	_, err = hooks.ExecuteHooksParallel(
		pool, &settings.ExecX, &hs.GlobalSharedHooks,
		results, logHookResults,
		settings.Args...)
	log.AssertNoErrorPanic(err, "Gobal shared hook execution failed.")
}

func logHookResults(res ...hooks.HookResult) {
	for _, r := range res {
		if r.Error == nil {
			if len(r.Output) != 0 {
				_, _ = log.GetInfoWriter().Write(r.Output)
			}
		} else {
			if len(r.Output) != 0 {
				_, _ = log.GetErrorWriter().Write(r.Output)
			}
			log.AssertNoErrorPanicF(r.Error, "Hook '%s' failed!",
				r.Hook.GetString())
		}
	}
}

func storePendingData(
	settings *HookSettings,
	uiSettings *UISettings,
	ignores *hooks.RepoIgnorePatterns,
	checksums *hooks.ChecksumStore) {

	// Store all ignore user patterns if there are new ones.
	if len(uiSettings.DisabledHooks) != 0 {

		// Add all back to the list ...
		for i := range uiSettings.DisabledHooks {
			ignores.User.AddNamespacePaths(uiSettings.DisabledHooks[i].NamespacePath)
		}

		// ... and store them
		err := hooks.StoreHookPatternsGitDir(ignores.User, settings.GitDirWorktree)
		log.AssertNoErrorF(err, "Could not store disabled hooks.")
	}

	// Store all checksums if there are any new ones.
	if len(uiSettings.TrustedHooks) != 0 {
		err := checksums.SyncChecksumAdd(uiSettings.TrustedHooks...)
		log.AssertNoErrorF(err, "Could not store checksum for hook")
	}
}
