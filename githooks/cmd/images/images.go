package images

import (
	ccm "github.com/gabyx/githooks/githooks/cmd/common"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/spf13/cobra"
)

func runImagesUpdate(ctx *ccm.CmdContext, imagesFile string) {
	repoDir, _, _ := ccm.AssertRepoRoot(ctx)

	containerMgr, err := hooks.NewContainerManager(ctx.GitX, false)
	ctx.Log.AssertNoErrorPanicF(err, "Could not create container manager.")

	hooksDir := hooks.GetGithooksDir(repoDir)
	err = hooks.UpdateImages(ctx.Log, hooksDir, repoDir, hooksDir, imagesFile, containerMgr)
	ctx.Log.AssertNoErrorF(err, "Could not build images in '%s'.", imagesFile)

	if strs.IsNotEmpty(imagesFile) {
		return
	}

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
		repo := &allRepos[rI]

		if exists, _ := cm.IsPathExisting(repo.RepositoryDir); !exists {
			ctx.Log.WarnF(
				"Shared repository '%s' is not available yet.\n"+
					"Use 'git hooks shared update'.", repo.URL)

			continue
		}

		hooksDir := hooks.GetSharedGithooksDir(allRepos[rI].RepositoryDir)

		ctx.Log.InfoF("%s", hooksDir)
		err = hooks.UpdateImages(
			ctx.Log,
			allRepos[rI].OriginalURL,
			allRepos[rI].RepositoryDir,
			hooksDir,
			"",
			containerMgr)
		ctx.Log.AssertNoErrorF(err, "Could not build images in '%s'.", allRepos[rI].OriginalURL)
	}
}

// NewCmd creates this new command.
func NewCmd(ctx *ccm.CmdContext) *cobra.Command {

	sharedCmd := &cobra.Command{
		Use:   "images",
		Short: "Manage container images.",
		Long:  "Manages container images used by Githooks repositories in the current repository."}

	imagesFile := ""
	imagesUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: `Build/pull container images.`,
		Long: "Build/pull container images in the current\n" +
			"repository and shared repositories which are needed for Githooks.",
		PreRun: ccm.PanicIfNotExactArgs(ctx.Log, 0),
		Run: func(c *cobra.Command, args []string) {
			runImagesUpdate(ctx, imagesFile)
		}}

	imagesUpdateCmd.Flags().StringVar(&imagesFile,
		"config", "",
		"Use the given '.images.yaml' for the update.\n"+
			"Useful to build images in shared repositories\n"+
			"'githooks/.images.yaml' directory.\n"+
			"Namespace is read from the current repository.")

	sharedCmd.AddCommand(ccm.SetCommandDefaults(ctx.Log, imagesUpdateCmd))

	return ccm.SetCommandDefaults(ctx.Log, sharedCmd)
}
