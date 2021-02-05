package updates

import (
	"gabyx/githooks/git"
	"gabyx/githooks/hooks"
)

// SetAutomaticUpdateCheckSettings set the automatic update settings.
func SetAutomaticUpdateCheckSettings(enable bool, reset bool) error {
	opt := hooks.GitCKAutoUpdateEnabled
	gitx := git.Ctx()

	switch {
	case reset:
		return gitx.UnsetConfig(opt, git.GlobalScope)
	case enable:
		return gitx.SetConfig(opt, true, git.GlobalScope)
	default:
		return gitx.SetConfig(opt, false, git.GlobalScope)
	}
}

// GetAutomaticUpdateCheckSettings gets the automatic update settings.
func GetAutomaticUpdateCheckSettings() (enabled bool, isSet bool) {
	conf := git.Ctx().GetConfig(hooks.GitCKAutoUpdateEnabled, git.GlobalScope)
	switch {
	case conf == "true":
		return true, true
	case conf == "false":
		return false, true
	default:
		return false, false
	}
}
