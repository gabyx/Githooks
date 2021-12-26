package git

import (
	"bufio"
	"regexp"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

type ConfigEntry struct {
	name    string
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
		cm.PanicF("Wrong scope '%s'", scope)

		return nil
	}
}

func parseConfig(s string, filterFunc func(string) bool) (c ConfigCache, err error) {

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
		cm.DebugAssert(len(keyValue) == 2)                  // nolint: gomnd
		if len(keyValue) != 2 || !filterFunc(keyValue[0]) { // nolint: gomnd
			return
		}

		c.add(keyValue[0], keyValue[0], keyValue[1], scope, false)
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

func NewConfigCache(gitx Context, filterFunc func(string) bool) (ConfigCache, error) {

	conf, err := gitx.Get("config", "--includes", "--list", "--null", "--show-scope")
	if err != nil {
		return ConfigCache{}, nil
	}

	return parseConfig(conf, filterFunc)
}

func (c *ConfigCache) SyncChangedValues() {}

// Get all config values for key `key` in the cache.
func (c *ConfigCache) getAll(key string, scope ConfigScope) (val []string, exists bool) {

	if scope == Traverse {
		val, exists = c.GetAll(key, SystemScope) // This order is how Git reports it.
		if !exists {
			val, exists = c.GetAll(key, GlobalScope)
			if !exists {
				val, exists = c.GetAll(key, LocalScope)
			}
		}

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
		vals = append(vals, c.GetAllRegex(key, SystemScope)...)
		vals = append(vals, c.GetAllRegex(key, GlobalScope)...)
		vals = append(vals, c.GetAllRegex(key, LocalScope)...)

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
