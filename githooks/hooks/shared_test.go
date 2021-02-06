package hooks

import (
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
