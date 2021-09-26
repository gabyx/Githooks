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
func CheckHookNames(hookNames []string) error {
	for _, h := range hookNames {
		h := strings.TrimPrefix(h, "!")

		if h != hookNameAll && h != hookNameServer && !strs.Includes(ManagedHookNames, h) {
			return cm.ErrorF(
				"Given value '%s' is not a hook name supported by Githooks\n"+
					"nor 'all' or 'server'.", h)
		}
	}

	return nil
}

// UnwrapHookNames returns a unique list of hook names built from the input.
// Variable `hookNames` can contain hook names, 'server', 'all'
// and negation prefix '!'.
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

	return s.ToList(), err
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

	// Deprecation
	if scope == git.GlobalScope {
		_ = git.Ctx().UnsetConfig(GitCKMaintainOnlyServerHooks, scope)
	}

	err = git.Ctx().SetConfig(GitCKMaintainedHooks, strings.Join(maintainedHooks, ", "), scope)

	if err != nil {
		err = cm.CombineErrors(err, cm.ErrorF("Could not set Git config '%s'.",
			GitCKMaintainedHooks))
	}

	return
}

// Get the maintained hooks from the Git config at scope `scope`.
// If an error occurs, all Githooks supported hooks are returned by default.
func GetMaintainedHooks(
	gitx *git.Context,
	scope git.ConfigScope) (hookNames []string, err error) {

	h := git.Ctx().GetConfig(GitCKMaintainedHooks, scope)

	if err != nil {
		err = cm.CombineErrors(err, cm.ErrorF("Could not read maintained hooks from Git config '%s'.\n"+
			"  cwd: '%s'\n"+
			"-> Fallback to all supported hooks.",
			GitCKMaintainedHooks, gitx.GetCwd()))

		return ManagedHookNames, err
	}

	maintainedHooks := strings.Split(h, ",")

	for i := range maintainedHooks {
		maintainedHooks[i] = strings.TrimSpace(maintainedHooks[i])
	}

	err = CheckHookNames(maintainedHooks)
	hookNames, e := UnwrapHookNames(maintainedHooks)
	err = cm.CombineErrors(err, e)

	if err != nil {
		err = cm.CombineErrors(err, cm.ErrorF("Maintained hooks in Git config '%s' is not valid.\n"+
			"  cwd: '%s'\n"+
			"-> Fallback to all supported hooks.",
			GitCKMaintainedHooks, gitx.GetCwd()))

		return ManagedHookNames, err
	}

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
