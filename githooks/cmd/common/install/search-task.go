package install

import (
	cm "gabyx/githooks/common"
	"gabyx/githooks/git"
	"path"
	"sort"
)

// PreCommitSearchTask is a task to search for pre-commit files.
type PreCommitSearchTask struct {
	Dir     string
	Matches []string
}

func (t *PreCommitSearchTask) Run(exitCh chan bool) (err error) {
	t.Matches, err = cm.Glob(path.Join(t.Dir,
		"**/templates/hooks/pre-commit.sample"),
		true)

	if SortSearchResults {
		sort.Strings(t.Matches)
	}

	return err
}

func (t *PreCommitSearchTask) Clone() cm.ITask {
	c := *t                    // Copy the struct.
	copy(t.Matches, c.Matches) // Create a new slice.

	return &c
}

type GitDirsSearchTask struct {
	Dir     string
	Matches []string
}

func (t *GitDirsSearchTask) Run(exitCh chan bool) (err error) {
	t.Matches, err = git.FindGitDirs(t.Dir)

	if SortSearchResults {
		sort.Strings(t.Matches)
	}

	return
}

func (t *GitDirsSearchTask) Clone() cm.ITask {
	c := *t                    // Copy the struct.
	copy(t.Matches, c.Matches) // Create a new slice.

	return &c
}
