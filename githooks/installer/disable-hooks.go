package main

import (
	cm "rycus86/githooks/common"
	"rycus86/githooks/git"
	"rycus86/githooks/hooks"
	strs "rycus86/githooks/strings"
)

func getHookDisableCallback(
	log cm.ILogContext,
	nonInteractive bool,
	dryRun bool,
	uiSettings *UISettings) func(file string) hooks.HookDisableOption {

	gitx := git.Ctx()

	if strs.IsEmpty(uiSettings.DeleteDetectedLFSHooks) {
		// Load default UI value from config.
		uiSettings.DeleteDetectedLFSHooks = gitx.GetConfig("githooks.deleteDetectedLFSHooks", git.GlobalScope)
	}

	return func(file string) (answer hooks.HookDisableOption) {

		userAnswer := "n"
		if strs.IsNotEmpty(uiSettings.DeleteDetectedLFSHooks) {
			userAnswer = uiSettings.DeleteDetectedLFSHooks
		}

		if !nonInteractive {
			var err error
			userAnswer, err = uiSettings.PromptCtx.ShowPromptOptions(
				"There is an LFS command statement in hook:\n"+
					strs.Fmt("'%s'\n", file)+
					"Githooks will call LFS hooks internally and LFS\n"+
					"should not be called twice.\n"+
					"Do you want to delete this hook instead of\n"+
					"being disabled/backed up?", "(No, yes, all, skip all)",
				"N/y/a/s",
				"No", "Yes", "All", "Skip All")

			log.AssertNoError(err, "Could not show prompt.")

			if (userAnswer == "s" || userAnswer == "a") && !dryRun {
				uiSettings.DeleteDetectedLFSHooks = userAnswer // Store the decision.
			}

		}

		switch userAnswer {
		case "a":
			fallthrough // yes delete all...
		case "y":
			return hooks.DeleteHook
		default:
			return hooks.BackupHook
		}
	}
}
