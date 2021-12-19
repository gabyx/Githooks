package hooks

import (
	"testing"

	strs "github.com/gabyx/githooks/githooks/strings"
	"github.com/stretchr/testify/assert"
)

func isSame(t *testing.T, a []string, b []string) {
	for _, b := range b {
		assert.Contains(t, a, b)
	}
	assert.Equal(t, len(a), len(b))
}

func TestCheckHookNames(t *testing.T) {

	h, err := getMaintainedHooksFromString("!all")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(h))

	h, err = getMaintainedHooksFromString("!all, pre-commit")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(h))

	h, err = getMaintainedHooksFromString("!all, p-commit")
	assert.Error(t, err)
	assert.Equal(t, len(ManagedHookNames), len(h))

	h, err = getMaintainedHooksFromString("!all,\npre-commit,    post-commit")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(h))
}

func TestHookNameUnwrap(t *testing.T) {

	// All hooks minus 1.
	res := strs.NewStringSetFromList(ManagedHookNames)
	res.Remove("post-merge")
	h, err := UnwrapHookNames([]string{"!post-merge"})
	isSame(t, res.ToList(), h)
	assert.NoError(t, err)

	h, err = UnwrapHookNames([]string{"!all"})
	assert.NoError(t, err)
	assert.NotNil(t, h)
	assert.Equal(t, 0, len(h))

	// All hooks minus 1.
	h, err = UnwrapHookNames([]string{"all", "!post-merge"})
	isSame(t, res.ToList(), h)
	assert.NoError(t, err)

	// All server hooks minus one.
	res = strs.NewStringSetFromList(ManagedServerHookNames)
	res.Remove("pre-push")
	h, err = UnwrapHookNames([]string{"server", "!pre-push"})
	isSame(t, res.ToList(), h)
	assert.NoError(t, err)

	// Only pre-commit, and post-commit.
	h, err = UnwrapHookNames([]string{"!all", "pre-commit", "post-commit"})
	isSame(t, []string{"pre-commit", "post-commit"}, h)
	assert.NoError(t, err)

	// Remove all hooks.
	h, err = UnwrapHookNames([]string{"server", "!all"})
	isSame(t, []string{}, h)
	assert.NoError(t, err)

	// Remove only server hooks.
	res = strs.NewStringSetFromList(ManagedHookNames)
	for _, s := range ManagedServerHookNames {
		res.Remove(s)
	}
	h, err = UnwrapHookNames([]string{"all", "!server"})
	isSame(t, res.ToList(), h)
	assert.NoError(t, err)

	// LFS hooks stays in the set.
	h, err = UnwrapHookNames([]string{"server", "!all", "all", "!all"})
	isSame(t, []string{}, h)
	assert.NoError(t, err)

	// Test invalid hook.
	res = strs.NewStringSetFromList(ManagedHookNames)
	h, err = UnwrapHookNames([]string{"all", "!post-gaga"})
	isSame(t, res.ToList(), h)
	assert.Error(t, err)

}
