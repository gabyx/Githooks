package git

import (
	"bufio"
	"regexp"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// ConfigScope Defines the scope of a config file, such as local, global or system.
type ConfigScope int

// Available ConfigScope's.
const (
	commandScope  ConfigScope = 0
	worktreeScope ConfigScope = 1

	LocalScope  ConfigScope = 2
	GlobalScope ConfigScope = 3
	SystemScope ConfigScope = 4

	Traverse ConfigScope = -1
)

type ConfigEntry struct {
	name    string
	values  []string
	changed bool
}

// ConfigMap holds all configs Git reads.
type ConfigMap map[string]*ConfigEntry

// GitConfigCache for faster read access.
type ConfigCache struct {
	scopes [5]ConfigMap
}

func (c *ConfigCache) getScopeMap(scope ConfigScope) ConfigMap {
	cm.DebugAssertF(int(scope) < len(c.scopes), "Wrong scope '%s'", scope)

	return c.scopes[int(scope)]
}

func toMapIdx(scope string) ConfigScope {

	switch scope {
	case "system":
		return SystemScope
	case "global":
		return GlobalScope
	case "local":
		return LocalScope
	case "worktree":
		return worktreeScope
	case "command":
		return commandScope
	default:
		return -1
	}
}

func toConfigArg(scope ConfigScope) string {
	if scope == Traverse {
		return ""
	}

	return "--" + ToConfigName(scope)
}

func ToConfigName(scope ConfigScope) string {

	switch scope {
	case SystemScope:
		return "system"
	case GlobalScope:
		return "global"
	case LocalScope:
		return "local"
	case worktreeScope:
		return "worktree"
	case commandScope:
		return "command"
	default:
		cm.PanicF("Wrong scope '%v'", scope)

		return ""
	}
}

func parseConfig(s string, filterFunc func(string) bool) (c ConfigCache, err error) {

	c.scopes = [5]ConfigMap{
		make(ConfigMap),
		make(ConfigMap),
		make(ConfigMap),
		make(ConfigMap),
		make(ConfigMap)}

	// Define a split function that separates on null-terminators.
	onNullTerminator := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data); i++ {
			if data[i] == '\x00' {
				return i + 1, data[:i], nil
			}
		}

		if !atEOF {
			return 0, nil, nil
		}
		// There is one final token to be delivered, which may be the empty string.
		// Returning bufio.ErrFinalToken here tells Scan there are no more tokens after this
		// but does not trigger an error to be returned from Scan itself.
		return 0, data, bufio.ErrFinalToken
	}

	addEntry := func(scope ConfigScope, keyValue []string) {
		cm.DebugAssert(len(keyValue) == 2) // nolint: gomnd
		if len(keyValue) != 2 ||
			(filterFunc != nil && !filterFunc(keyValue[0])) { // nolint: gomnd
			return
		}

		c.add(keyValue[0], keyValue[0], keyValue[1], scope, false)
	}

	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(onNullTerminator)

	// Scan.
	i := -1
	var txt string
	var scope string
	for scanner.Scan() {
		i++

		txt = scanner.Text()

		if i%2 == 0 {
			scope = txt
		} else {
			if strs.IsEmpty(scope) { // Can happen, but shouldn't...
				continue
			}

			idx := toMapIdx(scope)
			if idx < 0 {
				err = cm.ErrorF("Wrong Git config scope '%v' for value '%s'", scope, txt)

				return
			}

			addEntry(
				idx,
				strings.SplitN(txt, "\n", 2)) // nolint: gomnd
		}
	}

	return
}

func NewConfigCache(gitx Context, filterFunc func(string) bool) (ConfigCache, error) {

	conf, err := gitx.Get("config", "--includes", "--list", "--null", "--show-scope")
	if err != nil {
		return ConfigCache{}, err
	}

	return parseConfig(conf, filterFunc)
}

func (c *ConfigCache) SyncChangedValues() {}

// Get all config values for key `key` in the cache.
func (c *ConfigCache) getAll(key string, scope ConfigScope) (val []string, exists bool) {

	if scope == Traverse {
		for i := len(c.scopes) - 1; i >= 0; i-- {
			res, _ := c.getAll(key, ConfigScope(i)) // This order is how Git config reports it.
			val = append(val, res...)
		}
		exists = len(val) != 0

		return
	}

	m := c.getScopeMap(scope)
	v, inMap := m[key]
	if inMap && v.values != nil {
		val = append(val, v.values...) // dont return reference to internal slice.
		exists = true
	}

	return
}

// Get all config values for key `key` in the cache.
func (c *ConfigCache) GetAll(key string, scope ConfigScope) (val []string, exists bool) {
	return c.getAll(strings.ToLower(key), scope)
}

// Get all config values for regex key `key` in the cache.
func (c *ConfigCache) GetAllRegex(key *regexp.Regexp, scope ConfigScope) (vals []KeyValue) {
	if scope == Traverse {
		for i := len(c.scopes) - 1; i >= 0; i-- {
			vals = append(vals, c.GetAllRegex(key, ConfigScope(i))...)
		}

		return
	}

	m := c.getScopeMap(scope)
	for k, v := range m {
		if key.MatchString(k) {
			for i := range v.values {
				vals = append(vals, KeyValue{k, v.values[i]})
			}
		}
	}

	return
}

// Get a config value for key `key` in the cache.
func (c *ConfigCache) get(key string, scope ConfigScope) (val string, exists bool) {
	if scope == Traverse {
		for i := 0; i < len(c.scopes); i++ {
			val, exists = c.get(key, ConfigScope(i)) // This order is how Git config takes precedence over others.
			if exists {
				break
			}
		}

		return
	}

	m := c.getScopeMap(scope)
	v, inMap := m[key]
	if inMap && v.values != nil {
		// Get always the last value defined.
		// Git config behavior for multiple values for one key.
		val = v.values[len(v.values)-1]
		exists = true
	}

	return
}

// Get a config value for key `key` in the cache.
func (c *ConfigCache) Get(key string, scope ConfigScope) (val string, exists bool) {
	return c.get(strings.ToLower(key), scope)
}

func (c *ConfigCache) add(key string, name string, value string, scope ConfigScope, changed bool) {
	m := c.getScopeMap(scope)

	val, exists := m[key]
	if !exists {
		val = &ConfigEntry{name: name}
		m[key] = val
	}

	val.values = append(val.values, value)
	val.changed = changed
}

// Set sets a config value `value` for `key` in the cache.
func (c *ConfigCache) Set(key string, value string, scope ConfigScope) (added bool) {
	m := c.getScopeMap(scope)

	k := strings.ToLower(key)
	val, inMap := m[k]
	cm.PanicIfF(inMap && len(val.values) > 1,
		"Cannot overwrite multiple values in '%v'.", key)

	if !inMap || len(val.values) == 0 {
		c.add(k, key, value, scope, true)
		added = true
	} else if val.values[0] != value {
		val.values[0] = value
		// Try to set the name to the upper case
		// version for better readibility
		val.name = key
		val.changed = true
	}

	return
}

// IsSet tells if a config value for `key` is set in the cache.
func (c *ConfigCache) IsSet(key string, scope ConfigScope) (exists bool) {
	_, exists = c.Get(key, scope)

	return
}

// Add a config value `value` to a `key` in the cache.
func (c *ConfigCache) Add(key string, value string, scope ConfigScope) {
	c.add(strings.ToLower(key), key, value, scope, true)
}

// Unset unsets all config values for `key` is set in the cache.
func (c *ConfigCache) Unset(key string, scope ConfigScope) bool {
	m := c.getScopeMap(scope)

	val, exists := m[strings.ToLower(key)]
	if !exists || len(val.values) == 0 {
		return false
	}

	val.changed = true
	val.values = nil

	return true
}
