package hooks

import (
	"strings"

	"github.com/gabyx/githooks/githooks/git"
)

// GetStagedFiles gets all currently staged files.
// Delimited by `\x00`.
func GetStagedFiles(gitx *git.Context) (string, error) {
	changed, err := gitx.Get("diff", "--cached", "--diff-filter=ACMR", "--name-only", "-z")
	if err != nil {
		return "", err
	}

	changed = strings.TrimRight(changed, "\x00")

	return changed, nil
}
