package hooks

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/hashicorp/go-version"
)

// A file cache containing LFS hooks.
type LFSHooksCache interface {
	// Returns all LFS paths and file names inside the cache.
	GetLFSHooks() ([]string, []string, error)
}

type lfsHooksCache struct {
	lfsHookNames       []string
	repoDir            string
	requiredLFSVersion *version.Version
	initialized        bool

	failure error
}

// Returns all LFS paths and file names inside the cache.
func (l *lfsHooksCache) GetLFSHooks() ([]string, []string, error) {
	if !l.initialized && l.failure == nil {
		l.failure = l.init()
	}

	if l.failure != nil {
		return nil, nil, l.failure
	}

	lfsHookFiles := make([]string, len(l.lfsHookNames))
	for i := range l.lfsHookNames {
		f := path.Join(l.repoDir, "hooks", l.lfsHookNames[i])
		if cm.IsFile(f) {
			lfsHookFiles[i] = f
		}
	}

	return lfsHookFiles, l.lfsHookNames, nil
}

// Creates a new LFS hooks cache.
func NewLFSHooksCache(tempDir string) (_ LFSHooksCache, err error) {
	if !git.IsLFSAvailable() {
		return nil, nil
	}

	var l lfsHooksCache
	l.repoDir = path.Join(tempDir, "lfs-hooks")
	l.requiredLFSVersion, err = git.GetGitLFSVersion()

	return &l, err
}

func gitLFSInstall(gitx *git.Context) (err error) {
	err = gitx.Check("lfs", "install")

	if err != nil {
		err = cm.CombineErrors(err, cm.ErrorF("Could not install Git LFS hooks in\n"+
			"'%s'.\n"+
			"Please try manually by invoking:\n"+
			"  $ git -C '%[1]s' lfs install", gitx.GetCwd()))
	}

	return
}

// Initializes the cache.
func (l *lfsHooksCache) init() (err error) {

	if l.initialized {
		return nil
	}

	versionFile := path.Join(l.repoDir, "lfs-version.info")

	reinit := true

	if cm.IsFile(versionFile) {
		ver, err := os.ReadFile(versionFile)
		if err == nil {
			v, err := version.NewVersion(strings.TrimSpace(string(ver)))
			reinit = err != nil || !v.Equal(l.requiredLFSVersion)
		}
	}

	gitx := git.NewCtxSanitizedAt(l.repoDir)
	hooksDir := path.Join(l.repoDir, "hooks")
	if !cm.IsDirectory(hooksDir) || !gitx.IsGitRepo() {
		reinit = true
	}

	if reinit {
		err = os.MkdirAll(l.repoDir, cm.DefaultFileModeDirectory)

		if err != nil {
			return cm.CombineErrors(err, cm.ErrorF("Could not create LFS hooks cache in '%s'.", l.repoDir))
		}

		err = git.Init(l.repoDir, true)
		if err != nil {
			return
		}

		err = gitLFSInstall(gitx)
		if err != nil {
			return
		}
	}

	for i := range ManagedHookNames {
		hook := path.Join(l.repoDir, "hooks", ManagedHookNames[i])
		if cm.IsFile(hook) {
			l.lfsHookNames = append(l.lfsHookNames, ManagedHookNames[i])
		}
	}

	err = ioutil.WriteFile(versionFile, []byte(l.requiredLFSVersion.String()), cm.DefaultFileModeFile)
	if err != nil {
		err = cm.CombineErrors(err, cm.ErrorF("Could not write version file in LFS hooks cache '%s'.", l.repoDir))
	}

	l.initialized = true

	return
}
