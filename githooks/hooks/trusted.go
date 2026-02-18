package hooks

import (
	"os"
	"path"
	"path/filepath"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"
)

// GetTrustMarkerFile get the trust marker file in the current repo.
func GetTrustMarkerFile(repoDir string) string {
	return path.Join(GetGithooksDir(repoDir), "trust-all")
}

// GetTrustAllSetting gets the trust-all setting in the local Git configuration.
func GetTrustAllSetting(gitx *git.Context) (trustall bool, isSet bool) {
	conf := gitx.GetConfig(GitCKTrustAll, git.LocalScope)

	isSet = strs.IsNotEmpty(conf)
	trustall = conf == git.GitCVTrue

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
// It is only trusted if the trust marker is present and
// the `trustAll` settings is set to `trusted`.
// On any error `false` is reported together with the error.
func IsRepoTrusted(
	gitx *git.Context,
	repoPath string) (isTrusted bool, hasTrustFile bool, trustAllSet bool) {

	trustFile := GetTrustMarkerFile(repoPath)

	if cm.IsFile(trustFile) {
		hasTrustFile = true
		isTrusted, trustAllSet = GetTrustAllSetting(gitx)
	}

	return
}

// SetSkipUntrustedHooks sets the settings if the hook runner should fail on active non-trusted hooks.
func SetSkipUntrustedHooks(gitx *git.Context, enable bool, reset bool, scope git.ConfigScope) error {
	switch {
	case reset:
		return gitx.UnsetConfig(GitCKSkipUntrustedHooks, scope)
	default:
		return gitx.SetConfig(GitCKSkipUntrustedHooks, enable, scope)
	}
}

// SkipUntrustedHooks gets the settings if the hook runner should fail on active non-trusted hooks.
func SkipUntrustedHooks(gitx *git.Context, scope git.ConfigScope) (enabled bool, isSet bool) {
	var conf string
	conf, set := os.LookupEnv("GITHOOKS_SKIP_UNTRUSTED_HOOKS")
	if !set {
		conf = gitx.GetConfig(GitCKSkipUntrustedHooks, scope)
	}

	switch {
	case strs.IsEmpty(conf) || conf == git.GitCVFalse:
		return
	default:
		return conf == git.GitCVTrue, true
	}
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
	// checksumDir is the path to the checksum directories containing files
	// with file name equal to the checksum.
	checksumDir string

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

// AddChecksums sets the search directory to `path`.
func (t *ChecksumStore) SetSearchDirectory(path string) {
	t.checksumDir = path
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

// SyncChecksumAdd adds SHA1 checksums of a path to the search directory.
func (t *ChecksumStore) SyncChecksumAdd(checksums ...ChecksumResult) error {
	if strs.IsEmpty(t.checksumDir) {
		return cm.Error("No checksum directory.")
	}

	for i := range checksums {
		checksum := &checksums[i]

		cm.DebugAssertF(len(checksum.SHA1) == 40, "Wrong SHA1 hash '%s'", checksum.SHA1) // nolint: mnd

		dir := path.Join(t.checksumDir, checksum.SHA1[0:2])
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
// of a path from the search directory.
func (t *ChecksumStore) SyncChecksumRemove(sha1s ...string) (removed int, err error) {

	if strs.IsEmpty(t.checksumDir) {
		err = cm.Error("No checksum directory.")

		return
	}

	for _, sha1 := range sha1s {

		cm.DebugAssertF(len(sha1) == 40, "Wrong SHA1 hash '%s'", sha1) // nolint: mnd

		dir := path.Join(t.checksumDir, sha1[0:2])
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

	// Check first search directory ...
	if strs.IsNotEmpty(t.checksumDir) {
		bucket := sha1[0:2]
		exists, err := cm.IsPathExisting(path.Join(t.checksumDir, bucket, sha1[2:]))
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
			"and directory search path '%v'.",
		len(t.checksums),
		t.checksumDir)
}

// GetChecksumDirectoryGitDir gets the checksum file inside the Git directory.
func GetChecksumDirectoryGitDir(gitDir string) string {
	var conf string
	gitx := git.NewCtxAt(gitDir)
	scope := git.Traverse
	conf, set := os.LookupEnv("GITHOOKS_CHECKSUMS_DIR")
	if !set {
		conf = gitx.GetConfig(GitCKChecksumsDir, scope)
	}

	switch {
	case !strs.IsEmpty(conf) && filepath.IsAbs(conf):
		return conf
	default:
		return path.Join(gitDir, ChecksumsDir)
	}
}

// GetChecksumStorage loads the checksum store from the
// current Git directory.
func GetChecksumStorage(gitDirWorktree string) (store ChecksumStore, err error) {

	cacheDir := GetChecksumDirectoryGitDir(gitDirWorktree)

	fi, e := os.Lstat(cacheDir)

	if e == nil && fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		cacheDir, err = os.Readlink(cacheDir)
		if err != nil {
			return
		}

		if !filepath.IsAbs(cacheDir) {
			err = cm.ErrorF("Checksum store symbolic link '%v' needs to be absolute.", cacheDir)

			return
		}
	}

	store.SetSearchDirectory(cacheDir)

	return
}
