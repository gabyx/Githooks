package git

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/hashicorp/go-version"
)

const (
	// NullRef is the null reference used by git during certain hook execution.
	NullRef = "0000000000000000000000000000000000000000"
)

var lfsVersionRegex = regexp.MustCompile(`\d+\.\d+\.\d+`)

// Get Git LFS version.
func GetGitLFSVersion() (ver *version.Version, err error) {
	res, err := NewCtx().Get("lfs", "--version")
	if err != nil {
		return nil, err
	}

	ver, err = version.NewVersion(lfsVersionRegex.FindString(res))

	return
}

// IsBareRepo returns `true` if `c.GetCwd()` is a bare repository.
func (c *Context) IsBareRepo() bool {
	out, _ := c.Get("rev-parse", "--is-bare-repository")
	return out == GitCVTrue //nolint:nlreturn
}

// IsGitRepo returns `true` if `path` is a git repository (bare or non-bare).
func (c *Context) IsGitRepo() bool {
	return c.Check("rev-parse") == nil
}

// IsGitDir returns `true` if `c.GetCwd()` is a git repository (bare or non-bare).
func (c *Context) IsGitDir() bool {
	s, err := c.Get("rev-parse", "--is-inside-git-dir")

	return err == nil && s == GitCVTrue
}

// GetMainWorktree returns the main worktree
// based on the current context's working directory.
func (c *Context) GetMainWorktree() (string, error) {
	// This feature is kind of buggy in earlier version of git < 2.28.0
	// it returns a git directory instead of the work tree
	// We strip "/.git" from the output.
	trees, err := c.Get("worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}

	list := strs.SplitLinesN(trees, 2) //nolint:mnd
	if len(list) == 0 {
		return "", cm.ErrorF("Could not get main worktree in '%s'", c.GetCwd())
	}

	tree := strings.TrimSuffix(
		strings.TrimSpace(
			strings.TrimPrefix(list[0], "worktree")),
		"/.git")

	if strs.IsEmpty(tree) {
		return "", cm.ErrorF("Could not get main worktree in '%s'", c.GetCwd())
	}

	return filepath.ToSlash(tree), nil
}

// GetGitDirCommon returns the common Git directory.
// For normal repos this points to the `.git` directory.
// For worktrees this points to the main worktrees git dir.
// The env. variable GIT_COMMON_DIR has especially
// be introduced for multiple worktrees, see:
// https://github.com/git/git/commit/c7b3a3d2fe2688a30ddb8d516ed000eeda13c24e
func (c *Context) GetGitDirCommon() (gitDir string, err error) {
	// Git 2.46.x apparently runs `reference-transaction` on `git init`
	// where `rev-parse` fails despite the documentation saying it reports `$GIT_COMMON_DIR` if set
	// (it is set!) -> Bug was reported.
	gitDir = os.Getenv("GIT_COMMON_DIR")
	if strs.IsEmpty(gitDir) {
		gitDir = os.Getenv("GIT_DIR")
		if strs.IsEmpty(gitDir) {
			gitDir, err = c.Get("rev-parse", "--git-common-dir")
			if err != nil {
				return
			}
		}
	}

	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(c.GetCwd(), gitDir)
	}

	gitDir, err = filepath.Abs(gitDir)
	if err != nil {
		return
	}

	gitDir = filepath.ToSlash(gitDir)

	return
}

// GetGitDirWorktree returns the Git directory.
// For normal repos this points to the `.git` directory.
// For worktrees this points to the actual worktrees git dir `.git/worktrees/<....>/`.
func (c *Context) GetGitDirWorktree() (gitDir string, err error) {
	// Git 2.46.x apparently runs `reference-transaction` on `git init`
	// where `rev-parse` fails despite the documentation saying it reports `$GIT_DIR` if set
	// (it is set!) -> Bug was reported.
	gitDir = os.Getenv("GIT_DIR")
	if strs.IsEmpty(gitDir) {
		gitDir, err = c.Get("rev-parse", "--absolute-git-dir")
		if err != nil {
			return
		}
	}

	gitDir = filepath.ToSlash(gitDir)

	return
}

// GetRepoRoot returns the top level directory in a non-bare repository or the
// absolute Git directory in a bare repository for `topLevel`.
// This is the root level for Githooks.
// The `gitDir` is the common Git directory (main Git dir for worktrees).
func (c *Context) GetRepoRoot() (topLevel string, gitDir string, gitDirWorktree string, err error) {
	if gitDir, err = c.GetGitDirCommon(); err != nil {
		return
	}

	if gitDirWorktree, err = c.GetGitDirWorktree(); err != nil {
		return
	}

	if c.IsBareRepo() {
		topLevel = gitDir
	} else {
		if topLevel, err = c.Get("rev-parse", "--show-toplevel"); err != nil {
			return
		}
		topLevel = filepath.ToSlash(topLevel)
	}

	return
}

// GetCurrentBranch gets the current branch in repository.
func (c *Context) GetCurrentBranch() (string, error) {
	return c.Get("branch", "--show-current")
}

// FindGitDirs returns Git directories inside `searchDir`.
// Paths relative to `searchDir` containing `.dotfiles` (hidden files)
// will never be reported. Optionally the output can be sorted.
func FindGitDirs(searchDir string) (all []string, err error) {
	candidates, err := cm.Glob(path.Join(searchDir, "**/HEAD"), true)
	if err != nil {
		return all, err
	}

	// We obtain a list of HEAD files, e.g.
	// 	- ~/a/b/normal/.git/HEAD
	// 	- ~/a/b/normal/.git/.../HEAD  1) filter out
	// 	- ~/a/b/bare-repo/HEAD
	// 	- ~/a/b/bare-repo/.../HEAD

	repos := make(strs.StringSet, len(candidates))
	var dir, relPath string

	// Be consistent here , on windows we might get twice `
	// C:/a/.git` and also `c:/a/.git` in the loop below
	// because of the output of `GetGitDirCommon()``.
	// -> adjust the volume label to UpperCase always for storing
	// the result in the `StringMap`
	adjustVolumeNameCase := runtime.GOOS == cm.WindowsOsName

	// Filter wrong dirs out.
	for i := range candidates {
		dir = path.Dir(candidates[i])
		normalGitDir := path.Base(dir) == ".git"

		relPath, err = filepath.Rel(searchDir, dir) // filepath, because path.Rel is not available.
		if err != nil {
			return all, err
		}
		relPath = filepath.ToSlash(relPath)

		if normalGitDir && cm.ContainsDotFile(path.Dir(relPath)) ||
			!normalGitDir && cm.ContainsDotFile(relPath) {
			// With that we filter out matches which
			// contain .dotfiles in the relative path to the search dir.
			continue
		}

		// gitDir is always an absolute path
		gitDir, e := NewCtxAt(dir).GetGitDirCommon()

		if adjustVolumeNameCase && len(gitDir) >= 2 {
			gitDir = strings.ToUpper(gitDir[0:2]) + gitDir[2:]
		}

		// Check if its really a git directory.
		if e == nil &&
			NewCtxAt(gitDir).IsGitRepo() {
			repos.Insert(gitDir)
		}
	}

	all = repos.ToList()

	return all, err
}

// Initialize an empty repo at path `repoPath`.
func Init(repoPath string, bare bool) error {
	args := []string{"-c", "core.hooksPath=.git/hooks", "init", "--template="}
	if bare {
		args = append(args, "--bare")
	}

	// We must not execute this clone command inside a Git repo  (e.g. A)
	// due to `core.hooksPath=.git/hooks` which get applied to `A` -> Bug ?:
	// https://stackoverflow.com/questions/67273420/why-does-git-execute-hooks-from-an-other-repository
	ctx := NewCtxSanitizedAt(repoPath)
	if !cm.IsDirectory(ctx.GetCwd()) {
		if e := os.MkdirAll(ctx.GetCwd(), cm.DefaultFileModeDirectory); e != nil {
			return cm.ErrorF("Could not create working directory '%s'.", ctx.GetCwd())
		}
	}

	out, e := ctx.GetCombined(args...)

	if e != nil {
		return cm.ErrorF("Init Git repository in '%s' failed:\n%s", repoPath, out)
	}

	return nil
}

// Clone an URL to a path `repoPath`.
func Clone(repoPath string, url string, branch string, depth int) error {
	// Its important to not use any template directory here to not
	// install accidentally Githooks run-wrappers.
	// We set the `core.hooksPath` explicitly to its internal
	// hooks directory to not interfere
	// with global settings.
	// Also this installs LFS hooks, which comes handy for certain shared hook repos
	// with prebuilt binaries.
	args := []string{"clone", "-c", "core.hooksPath=.git/hooks", "--template=", "--single-branch"}

	if branch != "" {
		args = append(args, "--branch", branch)
	}

	if depth > 0 {
		args = append(args, strs.Fmt("--depth=%v", depth))
	}

	args = append(args, []string{url, repoPath}...)

	// We must not execute this clone command inside a Git repo  (e.g. A)
	// due to `core.hooksPath=.git/hooks` which get applied to `A` -> Bug ?:
	// https://stackoverflow.com/questions/67273420/why-does-git-execute-hooks-from-an-other-repository
	ctx := NewCtxSanitizedAt(path.Dir(repoPath))
	if !cm.IsDirectory(ctx.GetCwd()) {
		if e := os.MkdirAll(ctx.GetCwd(), cm.DefaultFileModeDirectory); e != nil {
			return cm.ErrorF("Could not create working directory '%s'.", ctx.GetCwd())
		}
	}

	out, e := ctx.GetCombined(args...)

	if e != nil {
		return cm.ErrorF(
			"Cloning of '%s' [branch: '%s']\ninto '%s' failed:\n%s",
			url,
			branch,
			repoPath,
			out,
		)
	}

	return nil
}

// Pull executes a pull in `repoPath`.
func (c *Context) Pull(remote string) error {
	out, e := c.GetCombined("pull", remote)
	if e != nil {
		return cm.ErrorF("Pulling '%s' in '%s' failed:\n%s", remote, c.GetCwd(), out)
	}

	return nil
}

// FetchBranch executes a fetch of a `branch` from the `remote` in `repoPath`.
// This command sadly does not automatically (git 2.30) fetch the tags on this branch
// automatically. Use `tagPattern` to specify explicitly which tags to fetch.
func (c *Context) FetchBranch(remote string, branch string, tagPattern string) error {
	cmd := []string{"fetch", "--force", "--prune", "--prune-tags", "--no-tags", remote, branch}

	if strs.IsNotEmpty(tagPattern) {
		cmd = append(cmd, strs.Fmt("refs/tags/%s:refs/tags/%s", tagPattern, tagPattern))
	}

	out, e := c.GetCombined(cmd...)
	if e != nil {
		return cm.ErrorF("Fetching of '%s' from '%s'\nin '%s' failed:\n%s",
			branch, remote, c.GetCwd(), out)
	}

	return nil
}

// GetCommits gets all commits in the ancestry path starting from `firstSHA` (excluded in the result)
// up to and including `lastSHA`.
func (c *Context) GetCommits(firstSHA string, lastSHA string) ([]string, error) {
	return c.GetSplit("rev-list", "--ancestry-path", strs.Fmt("%s..%s", firstSHA, lastSHA))
}

// GetCommitLog gets all commits in the ancestry path starting from `firstSHA` (excluded in the result)
// up to and including `lastSHA`.
func (c *Context) GetCommitLog(commitSHA string, format string) (string, error) {
	return c.Get("log", strs.Fmt("--format=%s", format), commitSHA)
}

// GetRemoteURLAndBranch reports the `remote`s `url` and
// the current `branch` of HEAD.
func (c *Context) GetRemoteURLAndBranch(
	remote string,
) (currentURL string, currentBranch string, err error) {
	currentURL = c.GetConfig("remote."+remote+".url", LocalScope)
	currentBranch, err = c.Get("symbolic-ref", "-q", "--short", HEAD)

	return
}

// PullOrClone either executes a pull in `repoPath` or if not
// existing, clones to this path.
func PullOrClone(
	repoPath string,
	url string,
	branch string,
	depth int,
	repoCheck func(*Context) error) (isNewClone bool, err error) {
	gitx := NewCtxSanitizedAt(repoPath)
	if gitx.IsGitRepo() {
		isNewClone = false

		if repoCheck != nil {
			if err = repoCheck(gitx); err != nil {
				return
			}
		}

		err = gitx.Pull("origin")
	} else {
		isNewClone = true

		if err = os.RemoveAll(repoPath); err != nil {
			err = cm.ErrorF("Could not remove directory '%s'.", repoPath)
			return //nolint:nlreturn
		}

		err = Clone(repoPath, url, branch, depth)
	}

	return
}

// RepoCheck is the function which is executed before a fetch.
// Arguments 1 and 2 are `url`, `branch`.
// Return an error to abort the action.
// Return `true` to trigger a complete reclone.
// Available ConfigScope's.
type RepoCheck = func(Context, string, string) (bool, error)

// FetchOrClone either executes a fetch in `repoPath` or if not
// existing, clones to this path.
// The callback `repoCheck` before a fetch can trigger a reclone.
func FetchOrClone(
	repoPath string,
	url string, branch string,
	depth int,
	tagPattern string,
	repoCheck RepoCheck) (isNewClone bool, gitx *Context, err error) {
	gitx = NewCtxSanitizedAt(repoPath)

	if gitx.IsGitRepo() {
		isNewClone = false

		if repoCheck != nil {
			var reclone bool
			if reclone, err = repoCheck(*gitx, url, branch); err != nil {
				return
			}

			isNewClone = reclone
		}
	} else {
		isNewClone = true
	}

	if isNewClone {
		if err = os.RemoveAll(repoPath); err != nil {
			return
		}
		err = Clone(repoPath, url, branch, depth)
	}

	if err != nil {
		return
	}

	err = gitx.FetchBranch("origin", branch, tagPattern)

	return
}

// IsRefReachable reports if `ref` (can be branch/tag/commit) is contained starting
// from `startRef`.
func IsRefReachable(gitx *Context, startRef string, ref string) (bool, error) {
	i, err := gitx.GetExitCode("merge-base", "--is-ancestor", ref, startRef)

	return i == 0, err
}

// GetTags gets the tags  at `commitSHA`.
func GetTags(gitx *Context, commitSHA string) ([]string, error) {
	if strs.IsEmpty(commitSHA) {
		commitSHA = HEAD
	}

	return gitx.GetSplit("tag", "--points-at", commitSHA)
}

// GetVersionAt gets the version & tag from the tags at `commitSHA`.
func GetVersionAt(gitx *Context, commitSHA string) (*version.Version, string, error) {
	tags, err := GetTags(gitx, commitSHA)
	if err != nil {
		return nil, "", err
	}

	for _, tag := range tags {
		ver, e := version.NewVersion(tag)
		if e == nil && ver != nil {
			return ver, tag, nil
		}
	}

	return nil, "", nil
}

// GetVersion gets the semantic version and its tag.
func GetVersion(
	gitx *Context,
	commitSHA string,
	matchPattern string,
) (v *version.Version, tag string, err error) {
	if commitSHA == HEAD {
		commitSHA, err = GetCommitSHA(gitx, HEAD)
		if err != nil {
			return
		}
	}

	tag, err = gitx.Get("describe", "--tags", "--abbrev=0", "--match", matchPattern)
	if err != nil {
		return
	}
	ver := tag

	// Get number of commits ahead.
	commitsAhead, err := gitx.Get("rev-list", "--count", strs.Fmt("%s..%s", ver, commitSHA))
	if err != nil {
		return
	}

	if commitsAhead != "0" {
		ver = strs.Fmt("%s+%s.%s", ver, commitsAhead, commitSHA[:7])
	}

	ver = strings.TrimPrefix(ver, "v")
	v, err = version.NewVersion(ver)

	return v, tag, err
}

// GetCommitSHA gets the commit SHA1 of the ref.
func GetCommitSHA(gitx *Context, ref string) (string, error) {
	if strs.IsEmpty(ref) {
		ref = HEAD
	}

	return gitx.Get("rev-parse", ref)
}

// GetLFSConfigFile gets the LFS config file inside the repository and
// `true` if existing.
func GetLFSConfigFile(repoDir string) (string, bool) {
	s := path.Join(repoDir, ".lfsconfig")

	return s, cm.IsFile(s)
}

// IsLFSAvailable tells if git-lfs is available in the path.
func IsLFSAvailable() bool {
	_, err := exec.LookPath("git-lfs")

	return err == nil
}
