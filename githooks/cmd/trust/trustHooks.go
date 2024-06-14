package trust

import (
	"path"

	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	"github.com/gabyx/githooks/githooks/cmd/ignore"
	"github.com/gabyx/githooks/githooks/cmd/list"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"

	"github.com/spf13/cobra"
)

func getAllHooks(
	log cm.ILogContext,
	hookNames []string,
	repoDir string,
	gitDir string,
	repoHooksDir string,
	shared hooks.SharedRepos,
	state *list.ListingState) (allHooks []hooks.Hook) {

	allHooks = make([]hooks.Hook, 0, 10+2*shared.GetCount()) // nolint: mnd

	gitx := git.NewCtxAt(repoDir)

	for _, hookName := range hookNames {

		// List replaced hooks (normally only one)
		replacedHooks := list.GetAllHooksIn(
			log, gitx, repoDir, path.Join(gitDir, "hooks"), hookName,
			hooks.NamespaceReplacedHook, state, false, true)
		allHooks = append(allHooks, replacedHooks...)

		// List repository hooks
		repoHooks := list.GetAllHooksIn(log, gitx, repoDir, repoHooksDir, hookName,
			hooks.NamespaceRepositoryHook, state, false, false)
		allHooks = append(allHooks, repoHooks...)

		// List all shared hooks
		sharedCount := 0
		for idx, sharedRepos := range shared {
			coll, count := list.GetAllHooksInShared(log, gitx,
				hookName, state, sharedRepos, hooks.SharedHookType(idx))
			sharedCount += count

			for i := range coll {
				allHooks = append(allHooks, coll[i].Hooks...)
			}
		}
	}

	return
}

func apply(log cm.ILogContext, hook *hooks.Hook, checksums *hooks.ChecksumStore, reset bool) {

	err := hook.AssertSHA1()
	log.AssertNoErrorPanicF(err, "Could not compute SHA1 hash for hook '%s'.", hook.Path)

	if reset {

		removed, err := checksums.SyncChecksumRemove(hook.SHA1)
		log.AssertNoErrorPanicF(err, "Could not sync checksum for hook '%s'.", hook.Path)

		if removed != 0 {
			log.InfoF("Removed trust checksum for hook '%s'.", hook.NamespacePath)
		} else {
			log.InfoF("No trust checksum for hook '%s'.", hook.NamespacePath)
		}

	} else {

		err = checksums.SyncChecksumAdd(
			hooks.ChecksumResult{
				SHA1:          hook.SHA1,
				Path:          hook.Path,
				NamespacePath: hook.NamespacePath})

		log.AssertNoErrorPanicF(err, "Could not sync checksum for hook '%s'.", hook.Path)

		log.InfoF("Set trust checksum for hook '%s'.", hook.NamespacePath)
	}
}

func runTrustPatterns(ctx *ccm.CmdContext, reset bool, all bool, patterns *hooks.HookPatterns) {
	repoDir, gitDir, gitDirWorktree := ccm.AssertRepoRoot(ctx)

	repoHooksDir := hooks.GetGithooksDir(repoDir)
	hookNames := hooks.ManagedHookNames

	state, shared, hookNamespace := list.PrepareListHookState(ctx, repoDir, repoHooksDir, gitDirWorktree, hookNames)
	allHooks := getAllHooks(ctx.Log, hookNames, repoDir, gitDir, repoHooksDir, shared, state)

	patterns.MakeRelativePatternsAbsolute(hookNamespace, "")

	countMatches := 0

	for i := range allHooks {
		hook := &allHooks[i]

		if all || patterns.Matches(hook.NamespacePath) {
			countMatches++
			apply(ctx.Log, hook, state.Checksums, reset)
		}
	}

	ctx.Log.PanicIfF(countMatches == 0,
		"Given pattern or paths did not match any hooks '%v'.",
		patterns)

}

// NewTrustHooksCmd creates this new command.
func NewTrustHooksCmd(ctx *ccm.CmdContext) *cobra.Command {

	reset := false
	all := false
	patterns := hooks.HookPatterns{}

	trustHooks := &cobra.Command{
		Use:   "hooks [flags]",
		Short: "Trust all hooks which match the glob patterns or namespace paths.",
		Long: `Trust all hooks which match the glob patterns or namespace paths given
by '--patterns' or '--paths'.` + "\n\n" +
			ignore.SeeHookListHelpText + "\n\n" +
			ignore.NamespaceHelpText + "\n\n" +
			ignore.PatternsHelpText,

		PreRun: func(cmd *cobra.Command, args []string) {
			ccm.PanicIfAnyArgs(ctx.Log)(cmd, args)

			count := len(patterns.NamespacePaths) + len(patterns.Patterns)
			if all {
				count++
			}

			ctx.Log.PanicIfF(count == 0, "You need to provide at least one pattern or namespace path.")
		},

		Run: func(cmd *cobra.Command, args []string) {

			runTrustPatterns(ctx, reset, all, &patterns)
		},
	}

	trustHooks.Flags().StringArrayVar(&patterns.Patterns, "pattern", nil,
		"Specified glob pattern matching hook namespace paths.")

	trustHooks.Flags().StringArrayVar(&patterns.NamespacePaths, "path", nil,
		"Specified path fully matching a hook namespace path.")

	trustHooks.Flags().BoolVar(&all, "all", false,
		`If the action applies to all found hooks.
(ignoring '--patterns', '--paths')`)

	trustHooks.Flags().BoolVar(&reset, "reset", false,
		"If the matched hooks are set 'untrusted'.")

	trustHooks.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		ccm.CheckGithooksSetup(ctx.Log, ctx.GitX)
	}

	return ccm.SetCommandDefaults(ctx.Log, trustHooks)
}
