package git

import (
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
		"\x00global\x00a.a\na3"

	c, err := parseConfig(s)

	local := c.scopes[0]
	global := c.scopes[1]
	system := c.scopes[2]

	assert.Nil(t, err)
	assert.Equal(t, 1, len(system))
	assert.Equal(t, 2, len(global))
	assert.Equal(t, 3, len(local))

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

	c.Add("a.aa", "upsi", GlobalScope)
	val, _ = c.Get("a.aa", GlobalScope)
	assert.Equal(t, "upsi", val)

	c.Add("a.aa", "upsi2", GlobalScope)
	val, _ = c.Get("a.aa", GlobalScope)
	assert.Equal(t, "upsi2", val)

	e := c.Set("a.aa", "upsi2", GlobalScope)
	assert.NotNil(t, e)

}
