package git

import (
	"bufio"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

type ConfigEntry struct {
	key     string
	values  []string
	changed bool
}

type ConfigMap map[string]*ConfigEntry

// GitConfigCache for faster read access.
type ConfigCache struct {
	scopes [3]ConfigMap
}

func (c *ConfigCache) getScopeMap(scope ConfigScope) ConfigMap {

	switch scope {
	case SystemScope:
		return c.scopes[2]
	case GlobalScope:
		return c.scopes[1]
	case LocalScope:
		return c.scopes[0]
	default:
		cm.DebugAssertF(false, "Wrong scope '%s'", scope)

		return nil
	}
}

func parseConfig(s string) (c ConfigCache, err error) {

	c.scopes = [3]ConfigMap{
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
		if len(keyValue) != 2 {            // nolint: gomnd
			return
		}

		c.add(keyValue[0], keyValue[1], scope, false)
	}

	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(onNullTerminator)

	// Scan.
	var scope string
	i := 0
	for scanner.Scan() {
		if i%2 == 0 {
			scope = scanner.Text()
			if strs.IsNotEmpty(s) {
				scope = "--" + scope
			}
		} else {
			addEntry(ConfigScope(scope), strings.SplitN(scanner.Text(), "\n", 2)) // nolint: gomnd
		}
		i++
	}

	return
}

func NewConfigCache(gitx Context) (ConfigCache, error) {

	conf, err := gitx.Get("config", "--list", "--null", "--show-scope")
	if err != nil {
		return ConfigCache{}, nil
	}

	return parseConfig(conf)
}

func (c *ConfigCache) add(key string, value string, scope ConfigScope, changed bool) {
	m := c.getScopeMap(scope)

	val, exists := m[key]
	if !exists {
		val = &ConfigEntry{key: key}
		m[key] = val
	}

	val.values = append(val.values, value)
	val.changed = changed
}

func (c *ConfigCache) set(key string, value string, scope ConfigScope, changed bool) error {
	m := c.getScopeMap(scope)

	val, exists := m[key]
	if exists {
		if len(val.values) > 1 {
			return cm.ErrorF("Cannot overwrite multiple values in '%v'.", key)
		}
	}

	c.add(key, value, scope, changed)

	return nil
}

// Get a config value in the cache.
func (c *ConfigCache) Get(key string, scope ConfigScope) (val string, exists bool) {
	if scope == Traverse {
		val, exists = c.Get(key, LocalScope)
		if !exists {
			val, exists = c.Get(key, GlobalScope)
			if !exists {
				val, exists = c.Get(key, SystemScope)
			}
		}

		return
	}

	m := c.getScopeMap(scope)
	v, exists := m[key]
	if exists {
		// Get always the last value defined.
		// Git config behavior for multiple values for one key.
		val = v.values[len(v.values)-1]
		exists = true
	}

	return
}

// Set a config value `value` for `key` in the cache.
func (c *ConfigCache) Set(key string, value string, scope ConfigScope) error {
	return c.set(key, value, scope, true)
}

// ADd a config value `value` to a `key` in the cache.
func (c *ConfigCache) Add(key string, value string, scope ConfigScope) {
	c.add(key, value, scope, true)
}
