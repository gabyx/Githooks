package exec

import (
	"os"
	"path"
	"strings"

	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/spf13/cobra"
)

func execPath(
	ctx *ccm.CmdContext,
	res hooks.QueryResult,
	repoDir string,
	opts execCmdOptions,
	namespaceEnvs hooks.NamespaceEnvs) (err error) {

	ctx.Log.AssertNoError(err, "Could not load namespace environment variables.")

	containerized := opts.Containarized ||
		hooks.IsContainerizedHooksEnabled(ctx.GitX, true)

	hookCmds := make(hooks.HookPrioList, 1)
	path := path.Join(res.HooksDir, res.NamespacePath)

	cmd, err := hooks.GetHookRunCmd(
		git.NewCtxAt(repoDir),
		path,
		res.RepositoryRoot,
		res.HooksDir,
		true, containerized,
		res.Namespace,
		namespaceEnvs.Get(res.Namespace))

	if err != nil {
		return err
	}

	hook := hooks.Hook{
		IExecutable:   cmd,
		Path:          path,
		Namespace:     res.Namespace,
		NamespacePath: res.NamespacePath,
		Active:        true,
		Trusted:       true,
	}

	hookCmds[0] = append(hookCmds[0], hook)

	var execRes []hooks.HookResult
	execx := cm.ExecContext{Cwd: repoDir, Env: os.Environ()}

	execRes, e := hooks.ExecuteHooksParallel(
		nil,
		&execx,
		hookCmds,
		execRes,
		func(res ...hooks.HookResult) { logHookResults(ctx.Log, res...) },
		opts.Args...,
	)

	ctx.Log.AssertNoError(e, "Launching path '%s' failed", opts.NamespacePath)

	return ctx.NewCmdExit(execRes[0].ExitCode, "Execution failed.")
}

func runExec(ctx *ccm.CmdContext, opts execCmdOptions) (exitCode error) {
	ctx.WrapPanicExitCode()
	repoDir, _, _ := ccm.AssertRepoRoot(ctx)
	hooksDir := hooks.GetGithooksDir(repoDir)

	namespaceEnvs, err := hooks.LoadNamespaceEnvs(hooksDir)
	ctx.Log.AssertNoError(err, "Could not load namespace environment variables.")

	results, foundAll, err := hooks.ResolveNamespacePaths(
		ctx.Log,
		ctx.GitX,
		ctx.InstallDir,
		repoDir,
		[]string{opts.NamespacePath})

	ctx.Log.AssertNoErrorPanicF(err, "Could not resolve namespace paths.")
	ctx.Log.PanicIf(!foundAll,
		"Did not resolve namespace path '%s'", opts.NamespacePath)

	return execPath(ctx, results[0], repoDir, opts, namespaceEnvs)
}

func logHookResults(log cm.ILogContext, res ...hooks.HookResult) {
	var sb strings.Builder

	for _, r := range res {
		if r.Error == nil {
			if len(r.Output) != 0 {
				_, _ = log.GetInfoWriter().Write(r.Output)
			}
		} else {
			if len(r.Output) != 0 {
				_, _ = log.GetErrorWriter().Write(r.Output)
			}
			log.AssertNoErrorF(r.Error, "Execution '%s' failed!", r.Hook.Path)
			_, _ = strs.FmtW(&sb, "\n%s '%s'", cm.ListItemLiteral, r.Hook.NamespacePath)
		}
	}
}

type execCmdOptions struct {
	NamespacePath string
	Args          []string
	Containarized bool
	Parallel      bool
}

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) *cobra.Command {

	var opts execCmdOptions

	execCmd := &cobra.Command{
		Use:   "exec namespace-path [args...]",
		Short: "Execute namespace paths pointing to an executable.",
		Long: "Execute namespace paths\n" +
			"pointing to an executable or a run configuration\n" +
			"(e.g. `ns:xxx/mypath/a/b/c.sh` or `ns:xxx/mypath/a/b/c.yaml`).\n" +
			"The execution is run the same as Githooks performs during its execution.\n\n" +
			"Its not meant to execute hooks but rather add-on scripts\n" +
			"inside Githooks repositories.\n\n" +
			"If containerized hooks are enabled the execution always runs containerized.",
		PreRun: ccm.PanicIfNotRangeArgs(ctx.Log, 1, -1),
		RunE: func(c *cobra.Command, args []string) error {
			opts.NamespacePath = args[0]

			if len(args) > 1 {
				opts.Args = args[1:]
			}

			return runExec(ctx, opts)
		}}

	execCmd.Flags().BoolVar(&opts.Containarized,
		"containerized", false, "Force the execution to be containerized.")

	execCmd.Flags().BoolVar(&opts.Parallel,
		"parallel", false, "Execute all paths in parallel (beware of race conditions).")

	return ccm.SetCommandDefaults(ctx.Log, execCmd)
}
