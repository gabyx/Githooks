package hooks

import (
	cm "gabyx/githooks/common"
	"gabyx/githooks/git"
	strs "gabyx/githooks/strings"
	"os"
	"path"
	"path/filepath"
)

// GetTrustMarkerFile get the trust marker file in the current repo.
func GetTrustMarkerFile(repoDir string) string {
	return path.Join(GetGithooksDir(repoDir), "trust-all")
}

// GetTrustAllSetting gets the trust-all setting in the local Git configuration.
func GetTrustAllSetting(gitx *git.Context) (trustall bool, isSet bool) {
	conf := gitx.GetConfig(GitCKTrustAll, git.LocalScope)

	isSet = strs.IsNotEmpty(conf)
	trustall = conf == "true"

	return
}

// SetTrustAllSetting sets the trust-all setting in the local Git configuration.
func SetTrustAllSetting(gitx *git.Context, enable bool, reset bool) error {
	switch {
	case reset:
		return gitx.UnsetConfig(GitCKTrustAll, git.LocalScope)
	default:
		return gitx.SetConfig(GitCKTrustAll, enable, git.LocalScope)
	}
}

// IsRepoTrusted tells if the repository `repoPath` is trusted.
// On any error `false` is reported together with the error.
func IsRepoTrusted(
	gitx *git.Context,
	repoPath string) (isTrusted bool, hasTrustFile bool) {

	trustFile := GetTrustMarkerFile(repoPath)

	if cm.IsFile(trustFile) {
		hasTrustFile = true
		isTrusted, _ = GetTrustAllSetting(gitx)
	}

	return
}

const (
	// SHA1Length is the string length of a SHA1 hash.
	SHA1Length = 40
)

// ChecksumResult defines the SHA1 hash and the path it was computed with together with the
// namespaced path.
type ChecksumResult struct {
	SHA1          string // SHA1 hash.
	Path          string // Path.
	NamespacePath string // Namespaced path.
}

// ChecksumStore represents a set of checksum which
// can be consulted to check if a hook is trusted or not.
type ChecksumStore struct {
	// checksumDirs are the paths to the checksum directories containing files
	// with file name equal to the checksum.
	checksumDirs []string

	// Checksums are the checksums manually added to this store
	checksums map[string]ChecksumData
}

// ChecksumData represents the data for one checksum which was stored.
type ChecksumData struct {
	Paths []string
}

type checksumFile struct {
	Path string
}

func newChecksumData(paths ...string) ChecksumData {
	return ChecksumData{paths}
}

// NewChecksumStore creates a checksum store from `path` (file or directory).
func NewChecksumStore(path string, addAsDirIfNonExisting bool) (ChecksumStore, error) {
	c := ChecksumStore{}
	err := c.AddChecksums(path, addAsDirIfNonExisting)

	return c, err
}

// AddChecksums adds checksum data from `path` (file or directory) to the store.
func (t *ChecksumStore) AddChecksums(path string, addAsDirIfNonExisting bool) error {

	if cm.IsDirectory(path) || addAsDirIfNonExisting {
		t.checksumDirs = append(t.checksumDirs, path)
	}

	return nil
}

func (t *ChecksumStore) assertData() {
	if t.checksums == nil {
		t.checksums = make(map[string]ChecksumData)
	}
}

// AddChecksum adds a SHA1 checksum of a path and returns if it was added (or it existed already).
func (t *ChecksumStore) AddChecksum(sha1 string, filePath string) bool {
	t.assertData()
	filePath = filepath.ToSlash(filePath)
	if data, exists := t.checksums[sha1]; exists {
		p := &data.Paths
		*p = append(*p, filePath)

		return true
	}

	t.checksums[sha1] = newChecksumData(filePath)

	return false
}

// SyncChecksumAdd adds SHA1 checksums of a path to the first search directory.
func (t *ChecksumStore) SyncChecksumAdd(checksums ...ChecksumResult) error {
	if len(t.checksumDirs) == 0 {
		return cm.Error("No checksum directory.")
	}

	for i := range checksums {
		checksum := &checksums[i]

		cm.DebugAssertF(len(checksum.SHA1) == 40, "Wrong SHA1 hash '%s'", checksum.SHA1) // nolint:gomnd

		dir := path.Join(t.checksumDirs[0], checksum.SHA1[0:2])
		err := os.MkdirAll(dir, cm.DefaultFileModeDirectory)
		if err != nil {
			return err
		}

		err = cm.StoreYAML(path.Join(dir, checksum.SHA1[2:]), &checksumFile{checksum.Path})
		if err != nil {
			return err
		}
	}

	return nil
}

// SyncChecksumRemove removes SHA1 checksums
// of a path from the first search directory.
func (t *ChecksumStore) SyncChecksumRemove(sha1s ...string) (removed int, err error) {

	if len(t.checksumDirs) == 0 {
		err = cm.Error("No checksum directory.")

		return
	}

	for _, sha1 := range sha1s {

		cm.DebugAssertF(len(sha1) == 40, "Wrong SHA1 hash '%s'", sha1) // nolint:gomnd

		dir := path.Join(t.checksumDirs[0], sha1[0:2])
		file := path.Join(dir, sha1[2:])

		if cm.IsFile(file) {
			if err = os.Remove(file); err != nil {
				return
			}

			removed++
		}
	}

	return
}

// IsTrusted checks if a path has been trusted.
func (t *ChecksumStore) IsTrusted(filePath string) (bool, string, error) {

	sha1, err := cm.GetSHA1HashFile(filePath)
	if err != nil {
		return false, "",
			cm.CombineErrors(cm.ErrorF("Could not get hash for '%s'", filePath), err)
	}

	// Check first all directories ...
	for _, dir := range t.checksumDirs {
		bucket := sha1[0:2]
		exists, err := cm.IsPathExisting(path.Join(dir, bucket, sha1[2:]))
		if exists {
			return true, sha1, nil
		} else if err != nil {
			return false, sha1, err
		}
	}

	// Check all checksums ...
	_, ok := t.checksums[sha1]
	if ok {
		return true, sha1, nil
	}

	return false, sha1, nil
}

// Summary returns a summary of the checksum store.
func (t *ChecksumStore) Summary() string {
	return strs.Fmt(
		"Checksum store contains '%v' checksums\n"+
			"and '%v' directory search paths.",
		len(t.checksums),
		len(t.checksumDirs))
}

// GetChecksumDirectoryGitDir gets the checksum file inside the Git directory.
func GetChecksumDirectoryGitDir(gitDir string) string {
	return path.Join(gitDir, ".githooks.checksums")
}

// GetChecksumStorage loads the checksum store from the Git config
// 'GitCKChecksumCacheDir' and if not possible from the
// current Git directory.
func GetChecksumStorage(gitx *git.Context, gitDirWorktree string) (store ChecksumStore, err error) {

	// Get the store from the config variable and fallback to Git dir if not existing.
	cacheDir := gitx.GetConfig(GitCKChecksumCacheDir, git.Traverse)
	loadFallback := strs.IsEmpty(cacheDir)

	if !loadFallback {
		e := store.AddChecksums(cacheDir, true)
		if e != nil {
			loadFallback = true
			err = cm.CombineErrors(err, cm.ErrorF("Could not add checksums from '%s'.", cacheDir), e)
		}
	}

	if loadFallback {
		cacheDir = GetChecksumDirectoryGitDir(gitDirWorktree)
		e := store.AddChecksums(cacheDir, true)
		if e != nil {
			err = cm.CombineErrors(err, cm.ErrorF("Could not add checksums from '%s'.", cacheDir), e)
		}
	}

	return
}
