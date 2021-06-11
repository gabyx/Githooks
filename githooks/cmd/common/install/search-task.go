package install

import (
	"path"
	"sort"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
)

// PreCommitSearchTask is a task to search for pre-commit files.
type PreCommitSearchTask struct {
	Dir     string
	Matches []string
}

// Run runs the search task of finding `pre-commit.sample` files.
func (t *PreCommitSearchTask) Run(exitCh chan bool) (err error) {
	t.Matches, err = cm.Glob(path.Join(t.Dir,
		"**/templates/hooks/pre-commit.sample"),
		true)

	if SortSearchResults {
		sort.Strings(t.Matches)
	}

	return err
}

// Clone clones the task. Necessary for safe Go routine execution.
func (t *PreCommitSearchTask) Clone() cm.ITask {
	c := *t                    // Copy the struct.
	copy(t.Matches, c.Matches) // Create a new slice.

	return &c
}

// GitDirsSearchTask holds data for searching Git directories.
type GitDirsSearchTask struct {
	Dir     string
	Matches []string
}

// Run searches Git directories.
func (t *GitDirsSearchTask) Run(exitCh chan bool) (err error) {
	t.Matches, err = git.FindGitDirs(t.Dir)

	if SortSearchResults {
		sort.Strings(t.Matches)
	}

	return
}

// Clone clones the task. Necessary for safe Go routine execution.
func (t *GitDirsSearchTask) Clone() cm.ITask {
	c := *t                    // Copy the struct.
	copy(t.Matches, c.Matches) // Create a new slice.

	return &c
}
