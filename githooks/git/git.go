package git

import (
	"os"
	"regexp"
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
	SystemScope ConfigScope = "--system"
	Traverse    ConfigScope = ""
	HEAD        string      = "HEAD"
)

// Context defines the context to execute it commands.
type Context struct {
	cm.CmdContext

	cache *ConfigCache
}

// NewCtxAt creates a git command execution context with
// working dir `cwd`.
func NewCtxAt(cwd string) *Context {
	return &Context{cm.NewCommandCtx("git", cwd, nil), nil}
}

// NewCtxSanitizedAt creates a git command execution context with
// working dir `cwd` and sanitized environement.
func NewCtxSanitizedAt(cwd string) *Context {
	return &Context{cm.NewCommandCtx("git", cwd, SanitizeEnv(os.Environ())), nil}
}

// NewCtx creates a git command execution context
// with current working dir.
func NewCtx() *Context {
	return NewCtxAt("")
}

// NewCtxSanitized creates a git command execution context
// with current working dir and sanitized environement.
func NewCtxSanitized() *Context {
	return NewCtxSanitizedAt("")
}

// SetConfigCache sets the Git config cache to use.
func (c *Context) InitConfigCache(filter func(string) bool) (err error) {
	cache, err := NewConfigCache(*c, filter)

	c.cache = &cache

	return
}

// GetConfig gets a Git configuration value for key `key`.
func (c *Context) GetConfig(key string, scope ConfigScope) (val string) {
	var err error

	if c.cache != nil {
		val, _ = c.cache.Get(key, scope)

		return
	}

	if scope != Traverse {
		val, err = c.Get("config", "--includes", string(scope), key)
	} else {
		val, err = c.Get("config", "--includes", key)
	}

	if err != nil {
		return ""
	}

	return
}

// LookupConfig gets a Git configuration value and
// reports if it exists or not.
func (c *Context) LookupConfig(key string, scope ConfigScope) (val string, exists bool) {
	var err error

	if c.cache != nil {
		return c.cache.Get(key, scope)
	}

	if scope != Traverse {
		val, err = c.Get("config", "--includes", string(scope), key)
	} else {
		val, err = c.Get("config", "--includes", key)
	}

	if err != nil {
		return "", false
	}

	return val, true
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
	if c.cache != nil {
		vals, _ := c.cache.GetAll(key, scope)

		return vals
	}

	s := c.getConfigWithArgs(key, scope, "--get-all")
	if strs.IsEmpty(s) {
		return nil
	}

	return strs.Filter(
		strs.SplitLines(s),
		strs.IsNotEmpty)
}

type KeyValue struct {
	Key   string
	Value string
}

// GetConfigRegex gets all Git configuration values for regex `regex`.
// Returns a list of pairs.
func (c *Context) GetConfigRegex(regex string, scope ConfigScope) (res []KeyValue) {
	if c.cache != nil {
		return c.cache.GetAllRegex(regexp.MustCompile(regex), scope)
	}

	configs, err := c.Get("config", "--includes", string(scope), "--get-regexp", regex)

	if err != nil {
		return
	}

	list := strs.SplitLines(configs)
	sort.Strings(list)

	res = make([]KeyValue, 0, len(list))

	for i := range list {
		if strs.IsEmpty(list[i]) {
			continue
		}

		keyValue := strings.SplitN(list[i], " ", 2) // nolint: gomnd
		cm.PanicIf(len(keyValue) == 0, "Wrong split.")
		// Handle unset values
		if len(keyValue) == 1 {
			keyValue = append(keyValue, "")
		}

		res = append(res, KeyValue{keyValue[0], keyValue[1]})
	}

	return
}

// SetConfig sets a Git configuration values with key `key`.
func (c *Context) SetConfig(key string, value interface{}, scope ConfigScope) error {
	cm.DebugAssert(scope != Traverse, "Wrong scope.")

	s := strs.Fmt("%v", value)
	if c.cache != nil {
		c.cache.Set(key, s, scope)
	}

	return c.Check("config", string(scope), key, s)
}

// AddConfig adds a Git configuration values with key `key`.
func (c *Context) AddConfig(key string, value interface{}, scope ConfigScope) error {
	cm.DebugAssert(scope != Traverse, "Wrong scope.")

	s := strs.Fmt("%v", value)
	if c.cache != nil {
		c.cache.Add(key, s, scope)
	}

	return c.Check("config", "--add", string(scope), key, s)
}

// UnsetConfig unsets all Git configuration values with key `key`.
func (c *Context) UnsetConfig(key string, scope ConfigScope) (err error) {
	if c.cache != nil {
		c.cache.Unset(key, scope)
	}

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

// IsConfigSet tells if a git config with `key` is set.
func (c *Context) IsConfigSet(key string, scope ConfigScope) bool {
	if c.cache != nil {
		return c.cache.IsSet(key, scope)
	}

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

// SanitizeOsEnv santizes the process environement from unwanted Git (possibly leaking)
// Git variables which might interfere with certain buggy Git commands.
func SanitizeOsEnv() error {
	err := os.Unsetenv("GIT_DIR")
	err = cm.CombineErrors(err, os.Unsetenv("GIT_WORK_TREE"))

	return err
}
