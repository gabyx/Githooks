package hooks

import (
	"strings"
	"testing"

	"github.com/gabyx/githooks/githooks/git"
	"github.com/stretchr/testify/assert"
)

func TestGitConfigPrefix(t *testing.T) {
	assert.True(t, strings.HasPrefix(GitCKInstallDir, git.GitConfigPrefix))
}
