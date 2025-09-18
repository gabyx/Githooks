package install

import (
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// GetHookDisableCallback returns the callback for prompting about hook disabling
// during install procedure.
func GetHookDisableCallback(
	log cm.ILogContext,
	gitx *git.Context,
	nonInteractive bool,
	uiSettings *UISettings) func(file string) hooks.HookDisableOption {
	if strs.IsEmpty(uiSettings.DeleteDetectedLFSHooks) {
		// Load default UI value from config.
		uiSettings.DeleteDetectedLFSHooks = gitx.GetConfig(hooks.GitCKDeleteDetectedLFSHooksAnswer, git.GlobalScope)
	}

	return func(file string) (answer hooks.HookDisableOption) {
		userAnswer := "n"
		if strs.IsNotEmpty(uiSettings.DeleteDetectedLFSHooks) {
			userAnswer = uiSettings.DeleteDetectedLFSHooks
		} else if !nonInteractive {
			var err error
			userAnswer, err = uiSettings.PromptCtx.ShowOptions(
				"There is an LFS command statement in hook:\n"+
					strs.Fmt("'%s'\n", file)+
					"Githooks will call LFS hooks internally and LFS\n"+
					"should not be called twice.\n"+
					"Do you want to delete this hook instead of\n"+
					"being disabled/backed up?", "(No, yes, all, skip all)",
				"N/y/a/s",
				"No", "Yes", "All", "Skip All")

			log.AssertNoError(err, "Could not show prompt.")

			switch userAnswer {
			case "s":
				uiSettings.DeleteDetectedLFSHooks = "n" // Store the decision.
			case "a":
				uiSettings.DeleteDetectedLFSHooks = "y" // Store the decision.
			}
		}

		switch userAnswer {
		case "a":
			fallthrough // yes delete all...
		case "y":
			log.WarnF("Previous hook '%s' will be disabled (deleted)", file)

			return hooks.DeleteHook
		default:
			log.WarnF("Previous hook '%s' will be disabled (backed up)", file)

			return hooks.BackupHook
		}
	}
}
