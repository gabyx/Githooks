package updates

import (
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
)

// SetAutomaticUpdateCheckSettings set the automatic update settings.
func SetAutomaticUpdateCheckSettings(enable bool, reset bool) error {
	opt := hooks.GitCKAutoUpdateEnabled
	gitx := git.NewCtx()

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
func GetAutomaticUpdateCheckSettings(gitx *git.Context) (enabled bool, isSet bool) {
	conf := gitx.GetConfig(hooks.GitCKAutoUpdateEnabled, git.GlobalScope)
	switch {
	case conf == git.GitCVTrue:
		return true, true
	case conf == git.GitCVFalse:
		return false, true
	default:
		return false, false
	}
}
