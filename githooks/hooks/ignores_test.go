package hooks

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreWildcardA(t *testing.T) {
	pattern := HookPatterns{Patterns: []string{"**/pre-commit*/pre-commit*"}}

	assert.Equal(t, true, pattern.Matches("ns:gh-self/pre-commit/pre-commit"))

	pattern.MakeRelativePatternsAbsolute("my-hooks", "")
	assert.Equal(t, false, pattern.Matches("ns:gh-self/pre-commit/pre-commit"))
	assert.Equal(t, true, pattern.Matches("ns:my-hooks/pre-commitA/pre-commitC"))
	assert.Equal(t, true, pattern.Matches("ns:my-hooks/pre-commitB/pre-commitD"))

	pattern.AddPatterns("!ns:my-*/*/pre-commit*", "ns:gh-*/**/**/**/*pre-commit*")
	assert.Equal(t, true, pattern.Matches("ns:gh-self/pre-commit/pre-commit"))
	assert.Equal(t, false, pattern.Matches("ns:my-hooks/pre-commitA/pre-commitC"))
	assert.Equal(t, false, pattern.Matches("ns:my-hooks/pre-commitB/pre-commitD"))
}

func TestIgnoreWildcardB(t *testing.T) {
	pattern := HookPatterns{Patterns: []string{"**/pre-commit*/pre-commit*", "!**/pre-commit*/*"}}

	assert.Equal(t, false, pattern.Matches("pre-commit/pre-commit"))

	pattern.MakeRelativePatternsAbsolute("my-hooks", "")
	assert.Equal(t, false, pattern.Matches("ns:gh-self/pre-commit/pre-commit"))
	assert.Equal(t, false, pattern.Matches("ns:my-hooks/pre-commitA/pre-commitC"))
	assert.Equal(t, false, pattern.Matches("ns:my-hooks/pre-commitB/pre-commitD"))

	pattern.AddPatterns("**/pre-commit*")
	pattern.MakeRelativePatternsAbsolute("my-hooks", "")
	assert.Equal(t, false, pattern.Matches("ns:gh-self/pre-commit/pre-commit"))
	assert.Equal(t, true, pattern.Matches("ns:my-hooks/pre-commitA/pre-commitC"))
	assert.Equal(t, true, pattern.Matches("ns:my-hooks/pre-commitB/pre-commitD"))
}

func TestIgnoreConfigVersion(t *testing.T) {
	f, e := os.CreateTemp("", "")
	assert.Nil(t, e)

	defer os.Remove(f.Name())
	_, e = io.WriteString(f,
		`
version: 999999
	  `)
	assert.Nil(t, e)

	_, e = LoadIgnorePatterns(f.Name())
	assert.Error(t, e)
	if e != nil {
		assert.Contains(t, e.Error(), "Githooks only supports version >= 1")
	}
}
