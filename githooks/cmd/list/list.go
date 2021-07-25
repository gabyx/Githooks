package list

import (
	"io"
	"path"
	"strings"

	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/pkg/math"
	"github.com/spf13/cobra"
)

func runList(ctx *ccm.CmdContext,
	hookNames []string, warnNotFound bool,
	onlyListActiveHooks bool, withBatchName bool) {

	repoDir, gitDir, gitDirWorktree := ccm.AssertRepoRoot(ctx)

	repoHooksDir := hooks.GetGithooksDir(repoDir)
	state, shared, _ := PrepareListHookState(ctx, repoDir, repoHooksDir, gitDirWorktree, hookNames)

	total := 0
	for _, hookName := range hookNames {

		list, count := listHooksForName(
			ctx.Log,
			hookName,
			repoDir,
			gitDir,
			repoHooksDir,
			shared,
			state,
			onlyListActiveHooks,
			withBatchName)

		if count != 0 {
			ctx.Log.InfoF("Hook: '%s' [%v]:%s", hookName, count, list)
		}

		total += count
	}

	pendingShared := filterPendingSharedRepos(shared)
	printPendingShared(ctx, pendingShared)

	ctx.Log.InfoF("Total listed hooks: '%v'.", total)
}

type ignoresPerHooksDir = map[string]*hooks.HookPatterns

// PrepareListHookState prepares all
// state needed to list all hooks in the current repository.
func PrepareListHookState(
	ctx *ccm.CmdContext,
	repoDir string,
	repoHooksDir string,
	gitDirWorktree string,
	hookNames []string) (state *ListingState, shared hooks.SharedRepos, hookNamespace string) {

	// Load checksum store
	checksums, err := hooks.GetChecksumStorage(ctx.GitX, gitDirWorktree)
	ctx.Log.AssertNoErrorF(err, "Errors while loading checksum store.")
	ctx.Log.DebugF("%s", checksums.Summary())

	// Set this repository's hook namespace.
	hookNamespace, err = hooks.GetHooksNamespace(repoHooksDir)
	ctx.Log.AssertNoErrorF(err, "Errors while loading hook namespace.")
	if strs.IsEmpty(hookNamespace) {
		hookNamespace = hooks.NamespaceRepositoryHook
	}

	// Load ignore patterns
	ignores, err := hooks.GetIgnorePatterns(repoHooksDir, gitDirWorktree, hookNames, hookNamespace)
	ctx.Log.AssertNoErrorF(err, "Errors while loading ignore patterns.")
	ctx.Log.DebugF("User ignore patterns: '%+q'.", ignores.User)
	ctx.Log.DebugF("Accumuldated repository ignore patterns: '%q'.", ignores.HooksDir)

	// Load all shared hooks
	shared = hooks.NewSharedRepos(8) //nolint: gomnd

	shared[hooks.SharedHookTypeV.Repo], err = hooks.LoadRepoSharedHooks(ctx.InstallDir, repoDir)
	ctx.Log.AssertNoErrorF(err, "Could not load repository shared hooks.")

	shared[hooks.SharedHookTypeV.Local], err = hooks.LoadConfigSharedHooks(ctx.InstallDir, ctx.GitX, git.LocalScope)
	ctx.Log.AssertNoErrorF(err, "Could not load local shared hooks.")

	shared[hooks.SharedHookTypeV.Global], err = hooks.LoadConfigSharedHooks(ctx.InstallDir, ctx.GitX, git.GlobalScope)
	ctx.Log.AssertNoErrorF(err, "Could not load global shared hooks.")

	isTrusted, _ := hooks.IsRepoTrusted(ctx.GitX, repoDir)
	isDisabled := hooks.IsGithooksDisabled(ctx.GitX, true)

	state = &ListingState{
		Checksums:          &checksums,
		Ignores:            &ignores,
		isRepoTrusted:      isTrusted,
		isGithooksDisabled: isDisabled,
		sharedIgnores:      make(ignoresPerHooksDir, 10)}

	return
}

// ListingState contains common state to successfully discover
// hooks.
type ListingState struct {
	Checksums *hooks.ChecksumStore
	Ignores   *hooks.RepoIgnorePatterns

	isRepoTrusted      bool
	isGithooksDisabled bool

	sharedIgnores ignoresPerHooksDir // sharedIgnores contains all ignores for the shared hooks
}

func filterPendingSharedRepos(shared hooks.SharedRepos) (pending hooks.SharedRepos) {

	pending = hooks.NewSharedRepos(0)

	// Filter out pending shared hooks.
	filter := func(shRepos []hooks.SharedRepo) (res []hooks.SharedRepo, pending []hooks.SharedRepo) {
		res = make([]hooks.SharedRepo, 0, len(shRepos))
		for idx := range shRepos {
			sh := &shRepos[idx]

			if cm.IsDirectory(sh.RepositoryDir) {
				res = append(res, *sh)
			} else {
				pending = append(pending, *sh)
			}
		}

		return
	}

	for idx := range shared {
		shared[idx], pending[idx] = filter(shared[idx])
	}

	return
}

func printPendingShared(ctx *ccm.CmdContext, shared hooks.SharedRepos) {

	count := shared.GetCount()
	if count == 0 {
		return
	}

	var sb strings.Builder

	listPending := func(shRepos []hooks.SharedRepo, indent string, category string) {
		for _, sh := range shRepos {
			_, err := strs.FmtW(&sb,
				"\n%s%s '%s' state: ['pending'], type: '%s'", indent, cm.ListItemLiteral, sh.OriginalURL, category)
			cm.AssertNoErrorPanic(err, "Could not write pending hooks.")
		}
	}

	indent := " "
	tagNames := hooks.GetSharedRepoTagNames()
	for i := range shared {
		idx := hooks.SharedHookType(i)
		listPending(shared[idx], indent, tagNames[idx])
	}

	ctx.Log.InfoF("Pending shared hooks [%v]:%s", count, sb.String())
}

func listHooksForName(
	log cm.ILogContext,
	hookName string,
	repoDir string,
	gitDir string,
	repoHooksDir string,
	shared hooks.SharedRepos,
	state *ListingState,
	onlyListActiveHooks bool,
	withBatchName bool) (string, int) {

	// List replaced hooks (normally only one)
	replacedHooks := GetAllHooksIn(
		log, repoDir, path.Join(gitDir, "hooks"), hookName,
		hooks.NamespaceReplacedHook, state, false, true)

	// List repository hooks
	repoHooks := GetAllHooksIn(
		log, repoDir, repoHooksDir, hookName,
		hooks.NamespaceRepositoryHook, state, false, false)

	// List all shared hooks
	sharedCount := 0
	all := make([]SharedHooks, 0, shared.GetCount())
	for idx, sharedRepos := range shared {
		coll, count := GetAllHooksInShared(log, hookName, state, sharedRepos, hooks.SharedHookType(idx))
		sharedCount += count
		all = append(all, coll...)
	}

	var sb strings.Builder
	paddingMax := 60
	printHooks := func(hooks []hooks.Hook, title string, category string) {
		if len(hooks) == 0 {
			return
		}

		padding := findPaddingListHooks(hooks, paddingMax)
		_, err := strs.FmtW(&sb, "\n %s", title)
		cm.AssertNoErrorPanicF(err, "Could not write hook state.")

		for i := range hooks {

			if onlyListActiveHooks && !hooks[i].Active {
				continue
			}

			sb.WriteString("\n")
			formatHookState(
				&sb, &hooks[i],
				category, withBatchName,
				state.isGithooksDisabled, padding, "  ")
		}
	}

	printHooks(replacedHooks, "Replaced:", "replaced")
	printHooks(repoHooks, "Repository:", "repo")

	tagNames := hooks.GetSharedRepoTagNames()
	for i := range all {
		printHooks(
			all[i].Hooks,
			strs.Fmt("Shared '%s':", all[i].Repo.OriginalURL),
			tagNames[all[i].Category])
	}

	return sb.String(), len(replacedHooks) + len(repoHooks) + sharedCount
}

func findPaddingListHooks(hooks []hooks.Hook, maxPadding int) int {
	const addChars = 3
	max := 0
	for i := range hooks {
		max = math.MaxInt(len(path.Base(hooks[i].Path))+addChars, max)
	}

	return math.MinInt(max, maxPadding)
}

// SharedHooks contains data about a shared hook repository.
type SharedHooks struct {
	Repo     *hooks.SharedRepo
	Category hooks.SharedHookType
	Hooks    []hooks.Hook
}

// GetAllHooksInShared gets all hooks in shared repositories.
func GetAllHooksInShared(
	log cm.ILogContext,
	hookName string,
	state *ListingState,
	sharedRepos []hooks.SharedRepo,
	category hooks.SharedHookType) (coll []SharedHooks, count int) {

	coll = make([]SharedHooks, 0, len(sharedRepos))

	for i := range sharedRepos {
		shRepo := &sharedRepos[i]

		hookNamespace := hooks.GetDefaultHooksNamespaceShared(shRepo)

		var allHooks []hooks.Hook

		if dir := hooks.GetSharedGithooksDir(shRepo.RepositoryDir); cm.IsDirectory(dir) {
			allHooks = GetAllHooksIn(log, shRepo.RepositoryDir,
				dir, hookName, hookNamespace, state, true, false)

		} else if dir := hooks.GetGithooksDir(shRepo.RepositoryDir); cm.IsDirectory(dir) {
			allHooks = GetAllHooksIn(log, shRepo.RepositoryDir,
				dir, hookName, hookNamespace, state, true, false)

		} else {
			allHooks = GetAllHooksIn(log, shRepo.RepositoryDir,
				shRepo.RepositoryDir, hookName, hookNamespace, state, true, false)
		}

		if len(allHooks) != 0 {
			count += len(allHooks)
			coll = append(coll,
				SharedHooks{
					Hooks:    allHooks,
					Repo:     shRepo,
					Category: category})
		}
	}

	return
}

// GetAllHooksIn gets all hooks in a hooks directory.
func GetAllHooksIn(
	log cm.ILogContext,
	rootDir string,
	hooksDir string,
	hookName string,
	hookNamespace string,
	state *ListingState,
	addInternalIgnores bool,
	isReplacedHook bool) []hooks.Hook {

	isTrusted := func(hookPath string) (bool, string) {
		if state.isRepoTrusted {
			return true, ""
		}

		trusted, sha, e := state.Checksums.IsTrusted(hookPath)
		log.AssertNoErrorF(e, "Could not check trust status '%s'.", hookPath)

		return trusted, sha
	}

	// Overwrite namespace/name.
	if isReplacedHook {
		hookName = hooks.GetHookReplacementFileName(hookName)
		cm.DebugAssert(strs.IsNotEmpty(hookNamespace), "Wrong namespace")

	} else {
		ns, err := hooks.GetHooksNamespace(hooksDir)
		log.AssertNoErrorPanicF(err, "Could not get hook namespace in '%s'", hooksDir)

		if strs.IsNotEmpty(ns) {
			hookNamespace = ns
		}
	}

	// Cache shared repository ignores
	hookDirIgnores := state.sharedIgnores[hooksDir]
	if hookDirIgnores == nil && addInternalIgnores {
		var e error

		igns, e := hooks.GetHookPatternsHooksDir(hooksDir, []string{hookName}, hookNamespace)
		log.AssertNoErrorF(e, "Could not get worktree ignores in '%s'.", hooksDir)
		state.sharedIgnores[hooksDir] = &igns
		hookDirIgnores = &igns
	}

	isIgnored := func(namespacePath string) bool {
		ignored, byUser := state.Ignores.IsIgnored(namespacePath)

		if isReplacedHook {
			return ignored && byUser // Replaced hooks can only be ignored by the user.
		} else if hookDirIgnores != nil {
			return ignored || hookDirIgnores.Matches(namespacePath)
		}

		return ignored
	}

	allHooks, _, err := hooks.GetAllHooksIn(
		rootDir, hooksDir,
		hookName, hookNamespace,
		isIgnored, isTrusted, false,
		!isReplacedHook)
	log.AssertNoErrorPanicF(err, "Errors while collecting hooks in '%s'.", hooksDir)

	return allHooks
}

func formatHookState(
	w io.Writer,
	hook *hooks.Hook,
	categeory string,
	withBatchName bool,
	isGithooksDisabled bool,
	padding int,
	indent string) {

	hooksFmt := strs.Fmt("%s%s %%-%vs : ", indent, cm.ListItemLiteral, padding)
	const stateFmt = " state: ['%[2]s', '%[3]s']"
	const disabledStateFmt = " state: ['disabled']"
	const categeoryFmt = ", type: '%[4]s'"
	const namespaceFmt = ", ns-path: '%[5]s'"
	const batchIDFmt = ", batch: '%[6]s'"

	hookPath := strs.Fmt("'%s'", path.Base(hook.Path))
	if isGithooksDisabled {
		fmt := hooksFmt + disabledStateFmt + categeoryFmt + namespaceFmt
		if withBatchName {
			fmt += batchIDFmt
		}
		_, err := strs.FmtW(w, fmt,
			hookPath, "", "", categeory, hook.NamespacePath, hook.BatchName)

		cm.AssertNoErrorPanicF(err, "Could not write hook state.")

		return
	}

	active := "active" // nolint: goconst
	trusted := "trusted"

	if !hook.Active {
		active = "ignored"
	}

	if !hook.Trusted {
		trusted = "untrusted"
	}

	fmt := hooksFmt + stateFmt + categeoryFmt + namespaceFmt
	if withBatchName {
		fmt += batchIDFmt
	}

	_, err := strs.FmtW(w, fmt,
		hookPath, active, trusted, categeory, hook.NamespacePath, hook.BatchName)

	cm.AssertNoErrorPanicF(err, "Could not write hook state.")
}

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) *cobra.Command {

	onlyListActiveHooks := false
	withBatchName := false

	listCmd := &cobra.Command{
		Use:   "list [type]...",
		Short: "Lists the active hooks in the current repository.",
		Long: "Lists the active hooks in the current repository along with their state.\n" +
			"This command needs to be run at the root of a repository.\n\n" +
			"If 'type' is given, then it only lists the hooks for that trigger event.\n" +
			"The supported hooks are:\n\n" +
			ccm.GetFormattedHookList("") + "\n\n" +
			"The value 'ns-path' is the namespaced path which is used for the ignore patterns.",

		PreRun: ccm.PanicIfNotRangeArgs(ctx.Log, 0, -1),

		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				args = strs.MakeUnique(args)

				for _, h := range args {
					ctx.Log.PanicIfF(!strs.Includes(hooks.ManagedHookNames, h),
						"Hook type '%s' is not managed by Githooks.", h)
				}

				runList(ctx, args, true, onlyListActiveHooks, withBatchName)

			} else {
				runList(ctx, hooks.ManagedHookNames, false, onlyListActiveHooks, withBatchName)
			}
		}}

	listCmd.Flags().BoolVar(&onlyListActiveHooks, "active", false, "Only list hooks with state 'active'.")
	listCmd.Flags().BoolVar(&withBatchName, "batch-name", false, "Also show the parallel batch name.")

	return ccm.SetCommandDefaults(ctx.Log, listCmd)
}
