package updates

import (
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
)

// SetUpdateCheckSettings set the automatic update settings.
func SetUpdateCheckSettings(enable bool, reset bool) error {
	opt := hooks.GitCKUpdateCheckEnabled
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

// GetUpdateCheckSettings gets the automatic update settings.
func GetUpdateCheckSettings(gitx *git.Context) (enabled bool, isSet bool) {
	conf := gitx.GetConfig(hooks.GitCKUpdateCheckEnabled, git.GlobalScope)
	switch conf {
	case git.GitCVTrue:
		return true, true
	case git.GitCVFalse:
		return false, true
	default:
		return false, false
	}
}
