package git

import (
	"os"
	"sort"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// ConfigScope Defines the scope of a config file, such as local, global or system.
type ConfigScope string

// Available ConfigScope's.
const (
	LocalScope  ConfigScope = "--local"
	GlobalScope ConfigScope = "--global"
	System      ConfigScope = "--system"
	Traverse    ConfigScope = ""
	HEAD        string      = "HEAD"
)

// Context defines the context to execute it commands.
type Context struct {
	cm.CmdContext
}

// CtxC creates a git command execution context with
// working dir `cwd`.
func CtxC(cwd string) *Context {
	return &Context{cm.CmdContext{BaseCmd: "git", Cwd: cwd}}
}

// CtxCSanitized creates a git command execution context with
// working dir `cwd` and sanitized environement.
func CtxCSanitized(cwd string) *Context {
	return (&Context{cm.CmdContext{BaseCmd: "git", Cwd: cwd, Env: SanitizeEnv(os.Environ())}})
}

// Ctx creates a git command execution context
// with current working dir.
func Ctx() *Context {
	return CtxC("")
}

// CtxSanitized creates a git command execution context
// with current working dir and sanitized environement.
func CtxSanitized() *Context {
	return CtxCSanitized("")
}

// GetConfig gets a Git configuration value for key `key`.
func (c *Context) GetConfig(key string, scope ConfigScope) string {
	var out string
	var err error

	if scope != Traverse {
		out, err = c.Get("config", "--includes", string(scope), key)
	} else {
		out, err = c.Get("config", "--includes", key)
	}

	if err == nil {
		return out
	}

	return ""
}

// LookupConfig gets a Git configuration value and
// reports if it exists or not.
func (c *Context) LookupConfig(key string, scope ConfigScope) (string, bool) {
	var out string
	var err error

	if scope != Traverse {
		out, err = c.Get("config", "--includes", string(scope), key)
	} else {
		out, err = c.Get("config", "--includes", key)
	}

	if err == nil {
		return out, true
	}

	return "", false
}

// getConfigWithArgs gets a Git configuration values for key `key`.
func (c *Context) getConfigWithArgs(key string, scope ConfigScope, args ...string) string {
	var out string
	var err error

	if scope != Traverse {
		out, err = c.Get(append(append([]string{"config", "--includes"}, args...), string(scope), key)...)
	} else {
		out, err = c.Get(append(append([]string{"config", "--includes"}, args...), key)...)
	}

	if err != nil {
		return ""
	}

	return out
}

// GetConfigAll gets a all Git configuration values for key `key`.
func (c *Context) GetConfigAll(key string, scope ConfigScope) []string {
	return strs.Filter(
		strs.SplitLines(c.getConfigWithArgs(key, scope, "--get-all")),
		strs.IsNotEmpty)
}

// GetConfigAllU gets a all Git configuration values unsplitted for key `key`.
func (c *Context) GetConfigAllU(key string, scope ConfigScope) string {
	return c.getConfigWithArgs(key, scope, "--get-all")
}

// GetConfigRegex gets all Git configuration values for regex `regex`.
// Returns a list of pairs.
func (c *Context) GetConfigRegex(regex string, scope ConfigScope) (res [][]string) {
	configs, err := c.Get("config", "--includes", string(scope), "--get-regexp", regex)

	if err != nil {
		return
	}

	list := strs.SplitLines(configs)
	sort.Strings(list)

	res = make([][]string, 0, len(list))

	for i := range list {
		if strs.IsEmpty(list[i]) {
			continue
		}

		keyValue := strings.SplitN(list[i], " ", 2)
		cm.PanicIf(len(keyValue) == 0, "Wrong split.")
		// Handle unset values
		if len(keyValue) == 1 {
			keyValue = append(keyValue, "")
		}

		res = append(res, keyValue)
	}

	return
}

// SetConfig sets a Git configuration values with key `key`.
func (c *Context) SetConfig(key string, value interface{}, scope ConfigScope) error {
	cm.DebugAssert(scope != Traverse, "Wrong input.")

	return c.Check("config", string(scope), key, strs.Fmt("%v", value))
}

// AddConfig adds a Git configuration values with key `key`.
func (c *Context) AddConfig(key string, value interface{}, scope ConfigScope) error {
	cm.DebugAssert(scope != Traverse, "Wrong input.")

	return c.Check("config", "--add", string(scope), key, strs.Fmt("%v", value))
}

// UnsetConfig unsets all Git configuration values with key `key`.
func (c *Context) UnsetConfig(key string, scope ConfigScope) (err error) {
	var exitC int

	if scope != Traverse {
		exitC, err = c.GetExitCode("config", "--unset-all", string(scope), key)
	} else {
		exitC, err = c.GetExitCode("config", "--unset-all", key)
	}

	if exitC == 5 || exitC == 0 { // nolint: gomnd
		//See: https: //git-scm.com/docs/git-config#_description
		return nil
	}

	return cm.CombineErrors(err, cm.ErrorF("Exit code: '%v'", exitC))
}

// IsConfigSet tells if a git config is set.
func (c *Context) IsConfigSet(key string, scope ConfigScope) bool {
	var err error
	if scope != Traverse {
		err = c.Check("config", string(scope), key)
	} else {
		err = c.Check("config", key)
	}

	return err == nil
}

// SanitizeEnv santizes the environement from unwanted Git (possibly leaking)
// Git variables which might interfere with certain buggy Git commands.
func SanitizeEnv(env []string) []string {
	return strs.Filter(env, func(s string) bool {
		return !strings.Contains(s, "GIT_DIR") &&
			!strings.Contains(s, "GIT_WORK_TREE")
	})
}
