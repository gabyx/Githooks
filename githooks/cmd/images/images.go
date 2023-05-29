package images

import (
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"

	"github.com/spf13/cobra"
)

func runImagesUpdate(ctx *ccm.CmdContext) {
	repoDir, _, _, err := ctx.GitX.GetRepoRoot()

	if err != nil {
		repoDir = ""
		ctx.Log.WarnF("Not inside a bare or non-bare repository.\n" +
			"Updating shared and local shared hooks skipped.")
	}

	hooksDir := hooks.GetGithooksDir(repoDir)
	err = hooks.UpdateImages(ctx.Log, hooksDir, repoDir, hooksDir)
	ctx.Log.AssertNoErrorF(err, "Could not build images in '%s'.", hooksDir)

	// Cycle through all shared hooks an return the first with matching namespace.
	allRepos, err := hooks.LoadRepoSharedHooks(ctx.InstallDir, repoDir)
	ctx.Log.AssertNoErrorPanicF(err, "Could not load shared hook list '%s'.", hooks.GetRepoSharedFileRel())
	local, err := hooks.LoadConfigSharedHooks(ctx.InstallDir, ctx.GitX, git.LocalScope)
	ctx.Log.AssertNoErrorPanicF(err, "Could not load local shared hook list.")
	global, err := hooks.LoadConfigSharedHooks(ctx.InstallDir, ctx.GitX, git.GlobalScope)
	ctx.Log.AssertNoErrorPanicF(err, "Could not load local shared hook list.")

	allRepos = append(allRepos, local...)
	allRepos = append(allRepos, global...)

	for rI := range allRepos {
		hooksDir := hooks.GetSharedGithooksDir(allRepos[rI].RepositoryDir)

		err = hooks.UpdateImages(ctx.Log, allRepos[rI].OriginalURL, allRepos[rI].RepositoryDir, hooksDir)
		ctx.Log.AssertNoErrorF(err, "Could not build images in '%s'.", allRepos[rI].OriginalURL)
	}
}

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) *cobra.Command {

	sharedCmd := &cobra.Command{
		Use:   "images",
		Short: "Manage container images.",
		Long:  "Manages container images used by Githooks repositories in the current repository."}

	imagesUpdateCmd := &cobra.Command{
		Use:    "update",
		Short:  `Build/pull container images.`,
		Long:   "Build/pull container images in the current repository which as needed for Githooks.",
		PreRun: ccm.PanicIfNotExactArgs(ctx.Log, 0),
		Run: func(c *cobra.Command, args []string) {
			runImagesUpdate(ctx)
		}}

	sharedCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, imagesUpdateCmd))

	return ccm.SetCommandDefaults(ctx.Log, sharedCmd)
}
