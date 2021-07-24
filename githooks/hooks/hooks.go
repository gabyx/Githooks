package hooks

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	strs "github.com/gabyx/githooks/githooks/strings"

	thx "github.com/pbenner/threadpool"
)

// Hook contains the data to an executable hook.
type Hook struct {
	// The executable of the hook.
	cm.IExecutable

	// The path to the file which configured this executable.
	Path string

	// The namespaced path of the hook `<namespace>/<relPath>`.
	NamespacePath string

	// If the hook is not ignored by any ignore patterns.
	// Has priority 1 for execution determination.
	Active bool
	// If the hook is trusted by means of the chechsum store.
	// Has priority 2 for execution determination.
	Trusted bool

	// SHA1 hash of the hook. (if determined)
	SHA1 string

	// BatchName denotes the parallel batch
	BatchName string
}

// HookPrioList is a list of lists of executable hooks.
// Each list contains a set of hooks which can potentially
// be executed in parallel.
type HookPrioList [][]Hook

// Hooks is a collection of all executable hooks.
// Json serialization is only for debug pruposes.
type Hooks struct {
	LocalHooks        HookPrioList
	RepoSharedHooks   HookPrioList
	LocalSharedHooks  HookPrioList
	GlobalSharedHooks HookPrioList
}

// HookResult is the data assembly of the output of an executed hook.
type HookResult struct {
	Hook   *Hook
	Output []byte
	Error  error
}

// TaggedHooksIndex is the index type for hook tags.
type TaggedHooksIndex int
type taggedHooksIndex struct {
	Replaced     TaggedHooksIndex
	Repo         TaggedHooksIndex
	SharedRepo   TaggedHooksIndex
	SharedLocal  TaggedHooksIndex
	SharedGlobal TaggedHooksIndex
	count        int
}

// TaggedHookIndices is a list of indices for all
// possible hooks Githooks supports.
// nolint: gomnd
var TaggedHookIndices = taggedHooksIndex{
	Replaced:     0,
	Repo:         1,
	SharedRepo:   2,
	SharedLocal:  3,
	SharedGlobal: 4,
	count:        5}

// TaggedHooks represents a map for all hooks sorted by tags.
// A list of hooks for each index `TaggedHookIndices`.
type TaggedHooks [][]Hook

// NewTaggedHooks returns a slice of hooks for each index `TaggedHookIndices`.
func NewTaggedHooks(capacity int) (res TaggedHooks) {
	res = make(TaggedHooks, TaggedHookIndices.count)
	for idx := range res {
		res[idx] = make([]Hook, 0, capacity)
	}

	return res
}

const (
	// TagNameReplaced is the hook tag for replaced hooks.
	TagNameReplaced = "replaced"
	// TagNameRepository is the hook tag for repository hooks.
	TagNameRepository = "repo"
	// TagNameSharedRepo is the hook tag for shared hooks inside the repository.
	TagNameSharedRepo = "shared:repo"
	// TagNameSharedLocal is the hook tag for shared hooks in the local Git config.
	TagNameSharedLocal = "shared:local"
	// TagNameSharedGLobal is the hook tag for shared hooks in the global Git config.
	TagNameSharedGLobal = "shared:global"
)

// GetHookTagNameMappings gets the mapping of a hook tag to a name.
// Indexable by `HookTagV`.
func GetHookTagNameMappings() []string {
	return []string{
		TagNameReplaced,
		TagNameRepository,
		TagNameSharedRepo,
		TagNameSharedLocal,
		TagNameSharedGLobal}
}

// IgnoreCallback is the callback type for ignoring hooks.
type IgnoreCallback = func(namespacePath string) (ignored bool)

// TrustCallback is the callback type for trusting hooks.
type TrustCallback = func(hookPath string) (trusted bool, sha1 string)

// GetAllHooksIn gets all hooks with name `hookName`
// in hooks dir `hookDir`.
// The reported `maxBatches` might include empty ones.
func GetAllHooksIn(
	rootDir string,
	hooksDir string,
	hookName string,
	hookNamespace string,
	isIgnored IgnoreCallback,
	isTrusted TrustCallback,
	lazyIfIgnored bool,
	parseRunnerConfig bool) (allHooks []Hook, maxBatches int, err error) {

	appendHook := func(prefix, hookPath, hookNamespace, batchName string) error {

		prefix += "/"

		// Prefix should always be removed! (we only have '/' in paths!)
		cm.DebugAssertF(strings.TrimPrefix(hookPath, prefix) != hookPath,
			"Prefix could not be removed '%s', '%s'.", prefix, hookPath)

		if strs.IsNotEmpty(hookNamespace) {
			hookNamespace = NamespacePrefix + hookNamespace
		}

		namespacedPath := path.Join(hookNamespace, strings.TrimPrefix(hookPath, prefix))
		ignored := isIgnored(namespacedPath)

		trusted := false
		sha := ""
		var runCmd cm.IExecutable

		if !ignored || !lazyIfIgnored {
			trusted, sha = isTrusted(hookPath)

			runCmd, err = GetHookRunCmd(hookPath, parseRunnerConfig, rootDir)
			if err != nil {
				return cm.CombineErrors(err,
					cm.ErrorF("Could not detect runner for hook\n'%s'", hookPath))
			}
		}

		allHooks = append(allHooks,
			Hook{
				IExecutable:   runCmd,
				Path:          hookPath,
				NamespacePath: namespacedPath,
				Active:        !ignored,
				Trusted:       trusted,
				SHA1:          sha,
				BatchName:     batchName})

		return nil
	}

	dirOrFile := path.Join(hooksDir, hookName)

	switch {
	case cm.IsDirectory(dirOrFile):

		var batchName string

		// If `.all-parallel` exists, throw all hooks into
		// the same batch with name 'all'.
		allParallel := cm.IsFile(path.Join(dirOrFile, ".all-parallel"))
		if allParallel {
			batchName = "all"
			maxBatches = 1
		}

		// Ignore file or skip folder.
		ignorePath := func(info os.FileInfo) error {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		// Collect all files starting from `hookPath`
		collectFiles := func(p string, info os.FileInfo) error {

			base := path.Base(p)

			if !allParallel {
				maxBatches++
				// The basename of the hook file or batch dir
				// defines the batch name.
				batchName = base
			}

			// Ignore `.dotfile` files
			if strings.HasPrefix(base, ".") {
				return ignorePath(info)
			}

			if info.IsDir() {

				// Get all files in the parallel batch folder
				err := cm.WalkPaths(p, func(p string, info os.FileInfo) error {
					// Ignore `.dotfile` files
					if strings.HasPrefix(path.Base(p), ".") {
						return ignorePath(info)
					}

					if !info.IsDir() {
						return appendHook(hooksDir, p, hookNamespace, batchName)
					}

					return nil
				})

				if err != nil {
					return err
				}

				// Skip this batch folder...
				return filepath.SkipDir
			}

			return appendHook(hooksDir, p, hookNamespace, batchName)
		}

		// Collect all hooks in e.g. `path/pre-commit/*`
		err = cm.WalkPaths(dirOrFile, collectFiles)

		if err != nil {
			err = cm.CombineErrors(cm.ErrorF("Errors while walking '%s'", dirOrFile), err)

			return
		}

	case cm.IsFile(dirOrFile):
		maxBatches++
		// Check hook in `path/pre-commit`
		err = appendHook(hooksDir, dirOrFile, hookNamespace, path.Base(dirOrFile))
	default:
		// Check hook in `path/pre-commit.yaml`
		runConfig := dirOrFile + ".yaml"
		if cm.IsFile(runConfig) {
			maxBatches++
			err = appendHook(hooksDir, runConfig, hookNamespace, path.Base(runConfig))
		}
	}

	return
}

// ExecuteHooksParallel executes hooks in parallel over a thread pool.
func ExecuteHooksParallel(
	pool *thx.ThreadPool,
	exec cm.IExecContext,
	hs *HookPrioList,
	res []HookResult,
	outputCallback func(res ...HookResult),
	args ...string) ([]HookResult, error) {

	// Count number of results we need
	nResults := 0
	for _, hooksGroup := range *hs {
		nResults += len(hooksGroup)
	}

	// Assert results is the right size
	if nResults > len(res) {
		res = append(res, make([]HookResult, nResults-len(res))...)
	} else {
		res = res[:nResults]
	}

	call := func(hookRes *HookResult, hook *Hook) {
		var err error
		hookRes.Output, err =
			cm.GetCombinedOutputFromExecutable(
				exec,
				hook,
				cm.UseOnlyStdin(os.Stdin),
				args...)

		hookRes.Error = err
		hookRes.Hook = hook
	}

	currIdx := 0
	for _, hooksGroup := range *hs {
		nHooks := len(hooksGroup)

		if nHooks == 0 {
			continue
		}

		if pool == nil {
			for idx := range hooksGroup {
				hookRes := &res[currIdx+idx]
				hook := &hooksGroup[idx]
				call(hookRes, hook)
				outputCallback(*hookRes)
			}
		} else {
			g := pool.NewJobGroup()

			err := pool.AddRangeJob(0, nHooks, g,
				func(idx int, pool thx.ThreadPool, erf func() error) error {
					hookRes := &res[currIdx+idx]
					hook := &hooksGroup[idx]
					call(hookRes, hook)

					return nil
				})

			if err != nil {
				return nil, err
			}

			if err = pool.Wait(g); err != nil {
				return nil, err
			}

			outputCallback(res[currIdx : currIdx+nHooks]...)
		}

		currIdx += nHooks
	}

	return res, nil
}

// StoreJSON stores the hooks priority list in JSON to the writer.
func (h *Hooks) StoreJSON(writer io.Writer) error {
	return cm.WriteJSON(writer, h)
}

// GetHooksCount gets the number of all hooks in the priority list.
func (h *HookPrioList) GetHooksCount() (count int) {
	for i := range *h {
		count += len((*h)[i])
	}

	return
}

// GetHooksCount gets the number of all hooks.
func (h *Hooks) GetHooksCount() int {
	return h.LocalHooks.GetHooksCount() + h.RepoSharedHooks.GetHooksCount() +
		h.LocalSharedHooks.GetHooksCount() + h.GlobalSharedHooks.GetHooksCount()
}

// AllHooksSuccessful returns `true`.
func AllHooksSuccessful(results []HookResult) bool {
	for _, h := range results {
		if h.Error != nil {
			return false
		}
	}

	return true
}

// AssertSHA1 ensures that the hook has its SHA1 computed.
func (h *Hook) AssertSHA1() (err error) {
	if strs.IsEmpty(h.SHA1) {
		h.SHA1, err = cm.GetSHA1HashFile(h.Path)
	}

	return
}
