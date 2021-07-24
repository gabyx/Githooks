package hooks

import (
	"path"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// hookIgnoreFile is the format of the ignore patterns file.
// A path is ignored if matched by `Patterns` or `NamespacePaths`.
type hookIgnoreFile struct {
	// Git ignores patterns matching hook namespace paths.
	Patterns []string `yaml:"patterns"`
	// Specific hook namespace paths (uses full match).
	NamespacePaths []string `yaml:"namespace-paths"`

	// The version of the file.
	Version int `yaml:"version"`
}

// hookIngoreFileVersion is the ignore file version.
var hookIngoreFileVersion = 1

// createHookIgnoreFile creates the data for the hook ignore file.
func createHookIgnoreFile() hookIgnoreFile {
	return hookIgnoreFile{Version: hookIngoreFileVersion}
}

// HookPatterns for matching the namespace path of hooks.
type HookPatterns struct {
	Patterns       []string
	NamespacePaths []string
}

// RepoIgnorePatterns is the list of possible ignore patterns in a repository.
type RepoIgnorePatterns struct {
	HooksDir HookPatterns // Ignores set by `.ignore.yaml` file in the hooks directory of the repository.
	User     HookPatterns // Ignores set by the `.ignore.yaml` file in the Git directory of the repository.
}

// CombineIgnorePatterns combines two ignore patterns.
func CombineIgnorePatterns(patterns ...*HookPatterns) HookPatterns {
	var p HookPatterns
	for _, pat := range patterns {
		p.Add(pat)
	}

	return p
}

// GetCount gets the count of all patterns.
func (h *HookPatterns) GetCount() int {
	return len(h.Patterns) + len(h.NamespacePaths)
}

// AddPatterns adds pattern to the patterns.
func (h *HookPatterns) AddPatterns(pattern ...string) {
	h.Patterns = append(h.Patterns, pattern...)
}

// AddPatternsUnique adds a namespace path to the patterns.
func (h *HookPatterns) AddPatternsUnique(pattern ...string) (added int) {
	h.Patterns, added = strs.AppendUnique(h.Patterns, pattern...)

	return
}

// AddNamespacePaths adds a namespace path to the patterns.
func (h *HookPatterns) AddNamespacePaths(namespacePath ...string) {
	h.NamespacePaths = append(h.NamespacePaths, namespacePath...)
}

// AddNamespacePathsUnique adds a namespace path to the patterns.
func (h *HookPatterns) AddNamespacePathsUnique(namespacePath ...string) (added int) {
	h.NamespacePaths, added = strs.AppendUnique(h.NamespacePaths, namespacePath...)

	return
}

// RemovePatterns removes patterns from the list.
func (h *HookPatterns) RemovePatterns(pattern ...string) (removed int) {
	c := 0

	for _, p := range pattern {
		h.Patterns, c = strs.Remove(h.Patterns, p)
		removed += c
	}

	return
}

// RemoveNamespacePaths adds a namespace path to the patterns.
func (h *HookPatterns) RemoveNamespacePaths(namespacePath ...string) (removed int) {
	c := 0
	for _, p := range namespacePath {
		h.NamespacePaths, c = strs.Remove(h.NamespacePaths, p)
		removed += c
	}

	return
}

// Add adds pattern from patterns `p` to itself.
func (h *HookPatterns) Add(p *HookPatterns) {
	h.AddPatterns(p.Patterns...)
	h.AddNamespacePaths(p.NamespacePaths...)
}

// AddUnique adds pattern uniquely from patterns `p` to itself.
func (h *HookPatterns) AddUnique(p *HookPatterns) (added int) {
	added = h.AddPatternsUnique(p.Patterns...)
	added += h.AddNamespacePathsUnique(p.NamespacePaths...)

	return
}

// Remove removes pattern from patterns `p` to itself.
func (h *HookPatterns) Remove(p *HookPatterns) (removed int) {
	removed = h.RemovePatterns(p.Patterns...)
	removed += h.RemoveNamespacePaths(p.NamespacePaths...)

	return
}

// RemoveAll removes all patterns.
func (h *HookPatterns) RemoveAll() (removed int) {
	removed = len(h.Patterns) + len(h.NamespacePaths)
	h.Patterns = nil
	h.NamespacePaths = nil

	return
}

// MakeRelativePatternsAbsolute makes all relative patterns (not starting with `ns:`) absolute
// by prepending `ns:<hookNamespace>/<rootPath>/`.
func (h *HookPatterns) MakeRelativePatternsAbsolute(hookNamespace string, rootPath string) {

	cm.DebugAssertF(strs.IsNotEmpty(hookNamespace), "Namespace must not be empty in any case!")

	replace := func(strs []string) {
		var invertPref string
		for i := range strs {

			startIdx, inverted := checkPatternInversion(strs[i])
			if inverted {
				invertPref = patternInversionPrefix
			} else {
				invertPref = ""
			}

			if !strings.HasPrefix(strs[i][startIdx:], NamespacePrefix) {
				// patterns like "**/*/pre-commit/*"
				strs[i] = invertPref + path.Clean(path.Join(NamespacePrefix+hookNamespace, rootPath, strs[i][startIdx:]))
			} else if strings.HasPrefix(strs[i][startIdx:], NamespacePrefix+NamespaceRepositoryHook) {
				// "ns:gh-self..." prefixes are directly replaced by the current namespace.
				strs[i] = invertPref + NamespacePrefix +
					hookNamespace + strings.TrimPrefix(strs[i][startIdx:], NamespacePrefix+NamespaceRepositoryHook)
			}

		}
	}

	replace(h.Patterns)
	replace(h.NamespacePaths)
}

// Reserve reserves 'nPatterns'.
func (h *HookPatterns) Reserve(nPatterns int) {
	if h.Patterns == nil {
		h.Patterns = make([]string, 0, nPatterns)
	}

	if h.NamespacePaths == nil {
		h.NamespacePaths = make([]string, 0, nPatterns)
	}
}

const patternInversionPrefix = "!"

// hasInvertPrefix checks a pattern for an inversion prefix "!".
func hasInvertPrefix(p string) bool {
	return strings.HasPrefix(p, patternInversionPrefix)
}

// hasInvertPrefixEscaped checks a pattern for an escaped inversion prefix "\!".
func hasInvertPrefixEscaped(p string) bool {
	return strings.HasPrefix(p, `\`+patternInversionPrefix)
}

// checkPatternInversion checks a pattern for inversion prefix "!" and returns
// the start index where the pattern starts and if its an inverted pattern or not.
func checkPatternInversion(p string) (int, bool) {

	if hasInvertPrefix(p) {
		return 1, true
	} else if hasInvertPrefixEscaped(p) {
		return 1, false
	}

	return 0, false
}

// Matches returns true if `namespacePath` matches any of the patterns and otherwise `false`.
func (h *HookPatterns) Matches(namespacePath string) (matched bool) {

	for _, p := range h.Patterns {

		// Note: Only forward slashes need to be used here in `hookPath`
		cm.DebugAssert(!strings.Contains(namespacePath, `\`),
			"Only forward slashes")

		startIdx, inverted := checkPatternInversion(p)

		// If we currently have a match, only an inversion can revert this...
		// so skip until we find an inversion.
		if matched && !inverted {
			continue
		}

		isMatch, err := cm.GlobMatch(p[startIdx:], namespacePath)
		cm.DebugAssertNoErrorF(err, "List contains malformed pattern '%s'", p)
		if err != nil {
			continue
		}

		if inverted {
			matched = matched && !isMatch
		} else {
			matched = matched || isMatch
		}
	}

	// The full matches can only change the result to `true`
	// They have no invertion "!" prefix.
	matched = matched || strs.Includes(h.NamespacePaths, namespacePath)

	return
}

// IsEmpty checks if there are any patterns stored.
func (h *HookPatterns) IsEmpty() bool {
	return len(h.Patterns)+len(h.NamespacePaths) == 0
}

// IsIgnored returns `true` if the hooksPath is ignored by either the worktree patterns or the user patterns
// and otherwise `false`. The second value is `true` if it was ignored by the user patterns.
func (h *RepoIgnorePatterns) IsIgnored(namespacePath string) (bool, bool) {
	if h.HooksDir.Matches(namespacePath) {
		return true, false
	} else if h.User.Matches(namespacePath) {
		return true, true
	}

	return false, false
}

// GetHookIgnoreFileHooksDir gets ignores files inside the hook directory.
// The `hookName` can be empty.
func GetHookIgnoreFileHooksDir(hooksDir string, hookName string) string {
	return path.Join(hooksDir, hookName, ".ignore.yaml")
}

// GetHookIgnoreFilesHooksDir gets ignores files inside the hook directory.
func GetHookIgnoreFilesHooksDir(hooksDir string, hookNames []string) (files []string) {
	files = make([]string, 0, 1+len(hookNames))

	for _, hookName := range hookNames {
		files = append(files, GetHookIgnoreFileHooksDir(hooksDir, hookName))
	}

	return
}

// GetHookPatternsHooksDir gets all ignored hooks in the hook directory.
func GetHookPatternsHooksDir(
	hooksDir string,
	hookNames []string,
	hookNamespace string) (patterns HookPatterns, err error) {

	files := GetHookIgnoreFilesHooksDir(hooksDir, hookNames)
	patterns.Reserve(2 * len(files)) // nolint: gomnd

	mainFile := GetHookIgnoreFileHooksDir(hooksDir, "")
	if cm.IsFile(mainFile) {
		ps, e := LoadIgnorePatterns(mainFile)
		err = cm.CombineErrors(err, e)

		ps.MakeRelativePatternsAbsolute(hookNamespace, "")
		patterns.Add(&ps)
	}

	for _, hookName := range hookNames {
		file := GetHookIgnoreFileHooksDir(hooksDir, hookName)
		if cm.IsFile(file) {
			ps, e := LoadIgnorePatterns(file)
			err = cm.CombineErrors(err, e)

			ps.MakeRelativePatternsAbsolute(hookNamespace, hookName)
			patterns.Add(&ps)
		}
	}

	return
}

// GetHookIgnoreFileGitDir gets
// the file of all ignored hooks in the current Git directory.
func GetHookIgnoreFileGitDir(gitDir string) string {
	return path.Join(gitDir, ".githooks.ignore.yaml")
}

// getHookPatternsGitDir gets all ignored hooks in the current Git directory.
func getHookPatternsGitDir(gitDir string, hookeNamespace string) (ps HookPatterns, err error) {
	file := GetHookIgnoreFileGitDir(gitDir)

	if cm.IsFile(file) {
		ps, err = LoadIgnorePatterns(file)
		ps.MakeRelativePatternsAbsolute(hookeNamespace, "")
	}

	return
}

// StoreHookPatternsGitDir stores all ignored hooks in the worktrees Git directory `gitDirWorktree`.
func StoreHookPatternsGitDir(patterns HookPatterns, gitDirWorktree string) error {
	return StoreIgnorePatterns(patterns,
		path.Join(gitDirWorktree, ".githooks.ignore.yaml"))
}

// LoadIgnorePatterns loads patterns.
func LoadIgnorePatterns(file string) (patterns HookPatterns, err error) {
	data := createHookIgnoreFile()

	err = cm.LoadYAML(file, &data)
	if err != nil {
		return
	}

	if data.Version == 0 {
		err = cm.ErrorF("Version '%v' needs to be greater than 0.", data.Version)

		return
	}

	patterns.Patterns = data.Patterns
	patterns.NamespacePaths = data.NamespacePaths

	// Filter all malformed patterns and report
	// errors.
	patternIsValid := func(p string) (valid bool) {
		if valid = IsHookPatternValid(p); !valid {
			err = cm.CombineErrors(err, cm.ErrorF("Pattern '%s' is malformed.", p))
		}

		return
	}

	patterns.Patterns = strs.Filter(patterns.Patterns, patternIsValid)

	return
}

// IsHookPatternValid validates a ignore `pattern`.
// This test supports `globstar` syntax.
func IsHookPatternValid(pattern string) bool {
	if pattern == "" {
		return false
	}
	_, e := cm.GlobMatch(pattern, "/test")

	return e == nil
}

// StoreIgnorePatterns stores patterns.
func StoreIgnorePatterns(patterns HookPatterns, file string) (err error) {

	data := hookIgnoreFile{
		Version:        hookIngoreFileVersion,
		Patterns:       strs.MakeUnique(patterns.Patterns),
		NamespacePaths: strs.MakeUnique(patterns.NamespacePaths)}

	return cm.StoreYAML(file, &data)
}

// GetIgnorePatterns loads all ignore patterns in the worktree's hooks dir and
// also in the worktrees Git directory.
func GetIgnorePatterns(
	hooksDir string,
	gitDirWorktree string,
	hookNames []string,
	hookNamespace string) (patt RepoIgnorePatterns, err error) {

	var e error

	patt.HooksDir, e = GetHookPatternsHooksDir(hooksDir, hookNames, hookNamespace)
	if e != nil {
		err = cm.CombineErrors(cm.Error("Could not get worktree ignore patterns."), e)
	}

	patt.User, e = getHookPatternsGitDir(gitDirWorktree, hookNamespace)
	if e != nil {
		err = cm.CombineErrors(err, cm.Error("Could not get user ignore patterns."), e)
	}

	return
}
