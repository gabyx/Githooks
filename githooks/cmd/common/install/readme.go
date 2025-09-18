package install

import (
	"os"
	"path"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"
)

func setupReadme(
	log cm.ILogContext,
	repoGitDir string,
	dryRun bool,
	uiSettings *UISettings) {
	mainWorktree, err := git.NewCtxAt(repoGitDir).GetMainWorktree()
	if err != nil || !git.NewCtxAt(mainWorktree).IsGitRepo() {
		log.WarnF("Main worktree could not be determined in:\n'%s'\n"+
			"-> Skipping Readme setup.",
			repoGitDir)

		return
	}

	readme := hooks.GetReadmeFile(mainWorktree)
	hookDir := path.Dir(readme)

	if !cm.IsFile(readme) {
		createFile := false

		switch uiSettings.AnswerSetupIncludedReadme {
		case "s":
			// OK, we already said we want to skip all
			return
		case "a":
			createFile = true
		default:

			var msg string
			if cm.IsDirectory(hookDir) {
				msg = strs.Fmt(
					"Looks like you don't have a '%s' folder in repository\n"+
						"'%s' yet.\n"+
						"Would you like to create one with a 'README'\n"+
						"containing a brief overview of Githooks?", hookDir, mainWorktree)
			} else {
				msg = strs.Fmt(
					"Looks like you don't have a 'README.md' in repository\n"+
						"'%s' yet.\n"+
						"A 'README' file might help contributors\n"+
						"and other team members learn about what is this for.\n"+
						"Would you like to add one now containing a\n"+
						"brief overview of Githooks?", mainWorktree)
			}

			answer, e := uiSettings.PromptCtx.ShowOptions(
				msg, "(Skip all, no, yes, all)",
				"S/n/y/a",
				"Skip All", "No", "Yes", "All")
			log.AssertNoError(e, "Could not show prompt.")

			switch answer {
			case "s":
				uiSettings.AnswerSetupIncludedReadme = answer
			case "a":
				uiSettings.AnswerSetupIncludedReadme = answer

				fallthrough
			case "y":
				createFile = true
			}
		}

		if createFile {
			if dryRun {
				log.InfoF("[dry run] Readme file '%s' would have been written.", readme)

				return
			}

			e := os.MkdirAll(path.Dir(readme), cm.DefaultFileModeDirectory)

			if e != nil {
				log.WarnF("Could not create directory for '%s'.\n"+
					"-> Skipping Readme setup.", readme)

				return
			}

			e = hooks.WriteReadmeFile(readme)
			log.AssertNoErrorF(e, "Could not write README file '%s'.", readme)
			log.InfoF("Readme file has been written to '%s'.", readme)
		}
	}
}
