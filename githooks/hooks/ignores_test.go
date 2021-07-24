package hooks

import (
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
