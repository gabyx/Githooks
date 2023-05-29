package hooks

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBranchSyntaxStripping(t *testing.T) {

	// Test branch syntax '...@(.*)' stripping
	url := "file://C:/what/here/then@branch"
	prefix, branch, err := parseSharedURLBranch(url)
	assert.Nil(t, err)
	assert.Equal(t, "file://C:/what/here/then", prefix)
	assert.Equal(t, "branch", branch)

	url = "file:///what/here/then@branch"
	prefix, branch, err = parseSharedURLBranch(url)
	assert.Nil(t, err)
	assert.Equal(t, "file:///what/here/then", prefix)
	assert.Equal(t, "branch", branch)

	url = "protocol://github.com/githooks.git@branch"
	prefix, branch, err = parseSharedURLBranch(url)
	assert.Nil(t, err)
	assert.Equal(t, "protocol://github.com/githooks.git", prefix)
	assert.Equal(t, "branch", branch)

	url = "protocol://user:pass@github.com/githooks.git@branch"
	prefix, branch, err = parseSharedURLBranch(url)
	assert.Nil(t, err)
	assert.Equal(t, "protocol://user:pass@github.com/githooks.git", prefix)
	assert.Equal(t, "branch", branch)

	url = "github.com:/githooks.git@branch"
	prefix, branch, err = parseSharedURLBranch(url)
	assert.Nil(t, err)
	assert.Equal(t, "github.com:/githooks.git", prefix)
	assert.Equal(t, "branch", branch)

	url = "user@github.com:/githooks.git@branch"
	prefix, branch, err = parseSharedURLBranch(url)
	assert.Nil(t, err)
	assert.Equal(t, "user@github.com:/githooks.git", prefix)
	assert.Equal(t, "branch", branch)

	// Local paths
	url = "C:/a/b/c/repo.git@branch"
	prefix, branch, err = parseSharedURLBranch(url)
	assert.Nil(t, err)
	assert.Equal(t, "C:/a/b/c/repo.git", prefix)
	assert.Equal(t, "branch", branch)

	url = "C:/@/b/c/repo.git@branch"
	prefix, branch, err = parseSharedURLBranch(url)
	assert.Nil(t, err)
	assert.Equal(t, "C:/@/b/c/repo.git", prefix)
	assert.Equal(t, "branch", branch)

}

func TestSharedConfigVersion(t *testing.T) {
	f, e := os.CreateTemp("", "")
	assert.Nil(t, e)

	defer os.Remove(f.Name())
	_, e = io.WriteString(f,
		`
version: 999999
	  `)
	assert.Nil(t, e)

	_, e = loadRepoSharedHooks(f.Name())
	assert.Error(t, e)
	if e != nil {
		assert.Contains(t, e.Error(), "Githooks only supports version >= 1")
	}
}
