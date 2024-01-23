package git

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitConfigCache(t *testing.T) {

	s := "system\x00a.a\nbla" +
		"\x00global\x00a.a\na1" +
		"\x00global\x00a.a\na2" +
		"\x00global\x00a.b\nvalueE" +
		"\x00local\x00a.a\na1" +
		"\x00local\x00a.a\na2" +
		"\x00local\x00a.b\na3.1\na3.2\na3.3" +
		"\x00local\x00a.c\nc1\x00\x00\x00\x00" +
		"\x00global\x00a.a\na3" +
		"\x00worktree\x00t.t\na2" +
		"\x00command\x00t.t\na3"

	c, err := parseConfig(s, func(string) bool { return true })

	command := c.scopes[0]
	worktree := c.scopes[1]
	local := c.scopes[2]
	global := c.scopes[3]
	system := c.scopes[4]

	assert.Nil(t, err)
	assert.Equal(t, 1, len(command))
	assert.Equal(t, 1, len(system))
	assert.Equal(t, 2, len(global))
	assert.Equal(t, 3, len(local))
	assert.Equal(t, 1, len(worktree))

	assert.Equal(t, "bla", system["a.a"].values[0])

	assert.Equal(t, "a1", global["a.a"].values[0])
	assert.Equal(t, "a2", global["a.a"].values[1])
	assert.Equal(t, "a3", global["a.a"].values[2])
	assert.Equal(t, "valueE", global["a.b"].values[0])

	assert.Equal(t, "a1", local["a.a"].values[0])
	assert.Equal(t, "a2", local["a.a"].values[1])
	assert.Equal(t, "a3.1\na3.2\na3.3", local["a.b"].values[0])
	assert.Equal(t, "c1", local["a.c"].values[0])

	val, _ := c.Get("a.a", GlobalScope)
	assert.Equal(t, "a3", val)

	_, exists := c.Get("a.aa", GlobalScope)
	assert.False(t, exists)

	// Set
	c.Set("s.s", "upsi", LocalScope)
	c.Set("s.s", "upsi", LocalScope)
	v, exists := c.GetAll("s.s", LocalScope)
	assert.True(t, exists)
	assert.Equal(t, 1, len(v))

	// Add
	c.Add("a.aa", "upsi", GlobalScope)
	val, _ = c.Get("a.aa", GlobalScope)
	assert.Equal(t, "upsi", val)

	c.Add("a.aa", "upsi2", GlobalScope)
	val, _ = c.Get("a.aa", GlobalScope)
	assert.Equal(t, "upsi2", val)

	assert.Panics(t, func() { c.Set("a.aa", "upsi2", GlobalScope) })
	assert.False(t, c.IsSet("a.aa", SystemScope))
	assert.False(t, c.IsSet("a.aa", LocalScope))
	assert.True(t, c.IsSet("a.aa", GlobalScope))
	assert.True(t, c.IsSet("a.aa", Traverse))

	assert.Panics(t, func() { c.Unset("a.aa", Traverse) })
	assert.False(t, c.Unset("a.aa", LocalScope))
	c.Add("a.aa", "upsi2", LocalScope)
	assert.True(t, c.Unset("a.aa", GlobalScope))
	assert.False(t, c.IsSet("a.aa", GlobalScope))
	assert.True(t, c.IsSet("a.aa", Traverse))

	v, exists = c.GetAll("t.t", Traverse)
	assert.True(t, exists)
	assert.Equal(t, 2, len(v))
	assert.Equal(t, []string{"a2", "a3"}, v)

	kv := c.GetAllRegex(regexp.MustCompile("t.*"), Traverse)
	assert.True(t, exists)
	assert.Equal(t, 2, len(v))
	assert.Equal(t, []KeyValue{
		{Key: "t.t", Value: "a2"},
		{Key: "t.t", Value: "a3"}}, kv)
}

func TestGitConfigCacheEnv(t *testing.T) {

	s := "system\x00githooks.a\n${MONKEY}-a" +
		"\x00global\x00githooks.b\n$MONKEY-b" +
		"\x00global\x00a.c\n$MONKEY-b"

	os.Setenv("MONKEY", "banana")
	assert.Equal(t, "banana", os.Getenv("MONKEY"))

	c, err := parseConfig(s, func(string) bool { return true })

	command := c.scopes[0]
	worktree := c.scopes[1]
	local := c.scopes[2]
	global := c.scopes[3]
	system := c.scopes[4]

	assert.Nil(t, err)
	assert.Equal(t, 0, len(command))
	assert.Equal(t, 1, len(system))
	assert.Equal(t, 2, len(global))
	assert.Equal(t, 0, len(local))
	assert.Equal(t, 0, len(worktree))

	assert.Equal(t, "banana-a", system["githooks.a"].values[0])
	assert.Equal(t, "banana-b", global["githooks.b"].values[0])
	assert.Equal(t, "$MONKEY-b", global["a.c"].values[0])
}
