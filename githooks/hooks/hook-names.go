package hooks

import (
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"
)

const (
	hookNameAll    = "all"
	hookNameServer = "server"
)

// Check hook names supporting also 'all', 'server' and negation prefix '!'.
// Additionally sanitize the names.
func CheckHookNames(hookNames []string) ([]string, error) {
	hookNames = strs.Map(hookNames, strings.TrimSpace)

	for _, h := range hookNames {
		h := strings.TrimPrefix(h, "!")

		if h != hookNameAll && h != hookNameServer && !strs.Includes(ManagedHookNames, h) {
			return hookNames, cm.ErrorF(
				"Given value '%s' in '%q' is not a hook name supported by Githooks\n"+
					"nor its '%s' or '%s'. ", h, hookNames, hookNameAll, hookNameServer)
		}
	}

	return hookNames, nil
}

// UnwrapHookNames returns a unique list of hook names built from the input.
// Variable `hookNames` can contain hook names, 'server', 'all'
// and negation prefix '!'. This function alyways returns a non-nil list.
func UnwrapHookNames(hookNames []string) ([]string, error) {

	var err error

	if len(hookNames) == 0 {
		return ManagedHookNames, nil
	}

	// Start with a full set if not one is given.
	s := strs.NewStringSet(len(ManagedHookNames))
	if hookNames[0] != hookNameAll && hookNames[0] != hookNameServer {
		hookNames = append([]string{hookNameAll}, hookNames...)
	}

	for i := range hookNames {

		h := strings.TrimPrefix(hookNames[i], "!")
		subtract := len(hookNames[i]) != len(h)

		switch h {
		case hookNameAll:
			if subtract {
				for _, m := range ManagedHookNames {
					s.Remove(m)
				}
			} else {
				for _, m := range ManagedHookNames {
					s.Insert(m)
				}
			}
		case hookNameServer:
			if subtract {
				for _, m := range ManagedServerHookNames {
					s.Remove(m)
				}
			} else {
				for _, m := range ManagedServerHookNames {
					s.Insert(m)
				}
			}
		default:
			if !strs.Includes(ManagedHookNames, h) {
				err = cm.CombineErrors(err, cm.ErrorF("Given value '%s' is not a supported hook name.", h))

				continue
			}

			if subtract {
				s.Remove(h)
			} else {
				s.Insert(h)
			}
		}
	}

	l := s.ToList()
	if l == nil {
		return []string{}, err
	}

	return l, err
}

// Store the maintained hooks to the Git config at scope `scope`.
// An empty list is treated as all Githooks supported hooks.
func SetMaintainedHooks(
	gitx *git.Context,
	maintainedHooks []string,
	scope git.ConfigScope) (err error) {

	// Set maintained hooks into global config.
	if maintainedHooks == nil {
		maintainedHooks = append(maintainedHooks, "all")
	}

	err = git.NewCtx().SetConfig(GitCKMaintainedHooks, strings.Join(maintainedHooks, ", "), scope)

	if err != nil {
		err = cm.CombineErrors(err, cm.ErrorF("Could not set Git config '%s'.",
			GitCKMaintainedHooks))
	}

	return
}

// getMaintainedHooksFromString gets all maintained hooks.
// By default it returns `all`.
func getMaintainedHooksFromString(maintainedHooks string) (hookNamesUnwrapped []string,
	maintHooks []string, err error) {

	if strs.IsNotEmpty(maintainedHooks) {
		maintHooks = strings.Split(maintainedHooks, ",")
		maintHooks, err = CheckHookNames(maintHooks)

		var e error
		hookNamesUnwrapped, e = UnwrapHookNames(maintHooks)
		err = cm.CombineErrors(err, e)

		if err == nil {
			return
		}

		err = cm.CombineErrors(err,
			cm.ErrorF("Maintained hooks '%s' is not valid. Fallback to all hooks.", maintainedHooks))
	}

	return ManagedHookNames, []string{"all"}, err
}

// Get the maintained hooks from the Git config at scope `scope`.
// If an error occurs, all Githooks supported hooks are returned by default.
func GetMaintainedHooks(
	gitx *git.Context,
	scope git.ConfigScope) (hookNames []string, maintainedHooks []string, isSet bool, err error) {
	h := git.NewCtx().GetConfig(GitCKMaintainedHooks, scope)

	hookNames, maintainedHooks, err = getMaintainedHooksFromString(h)
	isSet = strs.IsNotEmpty(h)

	return
}

// Get all other hooks from `ManagedHookNames` which are not in `hookNames`.
func GetAllOtherHooks(hookNames []string) (other []string) {
	for i := range ManagedHookNames {
		if !strs.Includes(hookNames, ManagedHookNames[i]) {
			other = append(other, ManagedHookNames[i])
		}
	}

	return
}
