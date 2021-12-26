package hooks

import (
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/gabyx/githooks/githooks/build"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"
	"github.com/hashicorp/go-version"
)

var runWrapperDetectionRegex = regexp.MustCompile(`https://github\.com/(gabyx|rycus86)/githooks`)

// IsRunWrapper answers the question if `filePath`
// is a Githooks hook template file.
func IsRunWrapper(filePath string) (bool, error) {
	return cm.MatchLineRegexInFile(filePath, runWrapperDetectionRegex)
}

// GetHookReplacementFileName returns the file name of a replaced custom Git hook.
func GetHookReplacementFileName(fileName string) string {
	return path.Base(fileName) + ".replaced.githook"
}

// GetRunWrapperContent gets the bytes of the hook template.
func getRunWrapperContent() ([]byte, error) {
	return build.Asset("embedded/run-wrapper.sh")
}

// WriteRunWrapper writes the run-wrapper to the file `filePath`.
func WriteRunWrapper(filePath string) (err error) {
	runWrapperContent, err := getRunWrapperContent()
	cm.AssertNoErrorPanic(err, "Could not get embedded run-wrapper content.")

	file, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write(runWrapperContent)
	if err != nil {
		return
	}
	err = file.Sync()
	if err != nil {
		return
	}

	// Make executable
	_ = file.Close()
	err = cm.MakeExecutable(filePath)

	return
}

var lfsDetectionRe = regexp.MustCompile(`(git\s+lfs|git-lfs)`)

// HookDisableOption are the options
// how to disable a hook.
type HookDisableOption int

const (
	// BackupHook defines that a hook file gets backed up.
	BackupHook HookDisableOption = 1
	// DeleteHook defines that a hook file gets deleted.
	DeleteHook HookDisableOption = 2
)

func disableHookIfLFSDetected(
	filePath string,
	disableCallBack func(file string) HookDisableOption) (disabled bool, deleted bool, err error) {

	found, err := cm.MatchLineRegexInFile(filePath, lfsDetectionRe)
	if err != nil {
		return
	}

	if found {
		disableOption := disableCallBack(filePath)

		switch disableOption {
		default:
			fallthrough
		case BackupHook:
			err = os.Rename(filePath, filePath+".disabled.githooks")
			disabled = true
		case DeleteHook:
			// The file cannot be potentially be opened/read.
			// Only a run-wrapper can be running (triggering an update)
			// and a run-wrapper gets not detected as an LFS hook.
			err = os.Remove(filePath)
			disabled = true
			deleted = true
		}
	}

	return
}

// Moves existing hook `dest` to `dir(path)/GetHookReplacementFileName(dest)`
// if its not a Githooks run-wrapper.
// If it is a run-wrapper dont do anything.
func moveExistingHooks(
	dest string,
	disableHookIfLFS func(file string) HookDisableOption,
	log cm.ILogContext) error {

	// Check there is already a Git hook in place and replace it.
	if !cm.IsFile(dest) {
		return nil
	}

	isRunWrapper, err := IsRunWrapper(dest)

	if err != nil {
		return cm.CombineErrors(err,
			cm.ErrorF("Could not detect if '%s' is a Githooks run-wrapper.", dest))
	} else if isRunWrapper {
		return nil
	}

	// Try to detect a potential LFS statements and
	// disable the hook (backup or delete).
	if disableHookIfLFS != nil {
		_, _, err := disableHookIfLFSDetected(dest, disableHookIfLFS)
		if err != nil {
			return err
		}
	}

	// Replace the file normally if it is still existing.
	if cm.IsFile(dest) {
		newDest := path.Join(path.Dir(dest), GetHookReplacementFileName(dest))
		if log != nil {
			log.InfoF("Saving existing Git hook '%s' to '%s'.", dest, newDest)
		}

		err = os.Rename(dest, newDest)
		if err != nil {
			return cm.CombineErrors(err,
				cm.ErrorF("Could not rename file '%s' to '%s'.", dest, newDest))
		}
	}

	return nil
}

// getHookDirTemp gets the Githooks temp. directory inside Git's hook directory.
func getHookDirTemp(hookDir string) string {
	return path.Join(hookDir, ".githooks-tmp")
}

// DeleteHookDirTemp deletes the temporary director inside the Git's hook directory.
func DeleteHookDirTemp(hookDir string) (err error) {
	dir := getHookDirTemp(hookDir)
	if cm.IsDirectory(dir) {
		return os.RemoveAll(dir)
	}

	return nil
}

// Cleans the temporary director inside the Git's hook directory.
func assertHookDirTemp(hookDir string) (dir string, err error) {
	dir = getHookDirTemp(hookDir)
	err = os.MkdirAll(dir, cm.DefaultFileModeDirectory)

	return
}

// InstallRunWrappers installs run-wrappers for the given `hookNames` in `dir`.
// Existing custom hooks get renamed.
// All deleted hooks by this function get moved to the `tempDir` directory, because
// we should not delete them yet.
// Missing LFS hooks are reinstalled.
func InstallRunWrappers(
	dir string,
	hookNames []string,
	beforeSaveCallback func(file string),
	disableHookIfLFS func(file string) HookDisableOption,
	lfsHooksCache LFSHooksCache,
	log cm.ILogContext) (nLFSHooks int, err error) {

	// Uninstall all other hooks.
	otherHooks := GetAllOtherHooks(hookNames)
	nLFSHooks, err = uninstallRunWrappers(otherHooks, dir, lfsHooksCache)
	if err != nil {
		return
	}

	// Install all maintained hooks.
	for _, hookName := range hookNames {

		dest := path.Join(dir, hookName)

		err = moveExistingHooks(dest, disableHookIfLFS, log)
		if err != nil {
			return
		}

		if beforeSaveCallback != nil {
			beforeSaveCallback(dest)
		}

		if cm.IsFile(dest) {
			// If still existing, it is a run-wrapper:

			// The file `dest` could be currently running,
			// therefore we move it to the temporary directly.
			// On Unix we could simply remove the file.
			// But on Windows, an opened file (mostly) cannot be deleted.
			// it might work, but is ugly.

			if runtime.GOOS == cm.WindowsOsName {
				backupDir, e := assertHookDirTemp(dir)
				if e != nil {
					err = cm.CombineErrors(e,
						cm.ErrorF("Could not create temp. dir in '%s'.", dir))

					return
				}

				moveDest := cm.GetTempPath(backupDir, "-"+path.Base(dest))
				err = os.Rename(dest, moveDest)
				if err != nil {
					err = cm.CombineErrors(err,
						cm.ErrorF("Could not move file '%s' to '%s'.", dest, moveDest))

					return
				}

			} else {
				// On Unix we simply delete the file, because that works even if the file is
				// open at the moment.
				err = os.Remove(dest)
				if err != nil {
					err = cm.CombineErrors(err,
						cm.ErrorF("Could not delete file '%s'.", dest))

					return
				}
			}

		}

		err = WriteRunWrapper(dest)
		if err != nil {
			err = cm.CombineErrors(err,
				cm.ErrorF("Could not write Githooks run-wrapper to '%s'.", dest))

			return
		}
	}

	return nLFSHooks, nil
}

// UninstallRunWrappers deletes run-wrappers in `dir`.
// Existing replaced hooks get renamed.
func UninstallRunWrappers(dir string, lfsHooksCache LFSHooksCache) (int, error) {
	return uninstallRunWrappers(ManagedHookNames, dir, lfsHooksCache)
}

func uninstallRunWrappers(
	hookNames []string,
	dir string,
	lfsHooksCache LFSHooksCache) (nLFSCount int, err error) {

	var e error
	var isRunWrapper bool

	for _, hookName := range hookNames {

		dest := path.Join(dir, hookName)

		if !cm.IsFile(dest) {
			continue
		}

		isRunWrapper, e = IsRunWrapper(dest)

		if e != nil {
			err = cm.CombineErrors(err,
				cm.ErrorF("Run-wrapper detection for '%s' failed.", dest))
		} else if isRunWrapper {
			// Delete the run-wrapper
			e := os.Remove(dest)

			if e == nil {
				// Move replaced hook (if existing) back in place.
				replacedHook := path.Join(path.Dir(dest), GetHookReplacementFileName(dest))

				if cm.IsFile(replacedHook) {
					if e := os.Rename(replacedHook, dest); e != nil {
						err = cm.CombineErrors(err,
							cm.ErrorF("Could not rename file '%s' to '%s'.",
								replacedHook, dest))
					}
				}

			} else {
				err = cm.CombineErrors(err, cm.ErrorF("Could not delete file '%s'.", dest))
			}
		}
	}

	if lfsHooksCache != nil {
		nLFSCount, e = reinstallLFSHooks(dir, hookNames, lfsHooksCache)
		if e != nil {
			err = cm.CombineErrors(err, e,
				cm.ErrorF("Could not reinstall LFS hooks into '%s'.", dir))
		}
	}

	return
}

// Get all missing LFS hook in `hookNames` and install them into `dir`.
func reinstallLFSHooks(
	dir string,
	hookNames []string,
	lfsHooksCache LFSHooksCache) (count int, err error) {

	if hookNames == nil {
		return
	}

	lfsHookPaths, lfsHookNames, err := lfsHooksCache.GetLFSHooks()

	if err != nil {
		return
	}

	for i := range lfsHookNames {

		if !strs.Includes(hookNames, lfsHookNames[i]) {
			// no LFS hooks
			continue
		}

		count += 1
		src := lfsHookPaths[i]
		dest := path.Join(dir, lfsHookNames[i])

		// Copy LFS hooks to destination.
		if cm.IsFile(dest) {

			equal, e := cm.AreChecksumsIdentical(src, dest)
			if equal {
				continue // This is obviously a LFS hook already. Skip it.
			}

			err = cm.CombineErrors(err, e)

			file, e := os.ReadFile(src)
			lfsContent := "  | " + strings.ReplaceAll(string(file), "\n", "\n  | ")

			err = cm.CombineErrors(err, e, cm.ErrorF("Cannot install LFS hook at '%s' because it already exists\n"+
				"and contains no 'git lfs' statement.\n"+
				"Either delete the hook and rerun the command or incorporate the following\n"+
				"content into the file '%s':\n"+
				"%s", dest, dest, string(lfsContent)))

			continue
		}

		e := cm.CopyFileOrDirectory(src, dest)
		if e != nil {
			err = cm.CombineErrors(e, cm.ErrorF("Cannot move LFS hook from '%s' to '%s'.", src, dest))

			continue
		}
	}

	return
}

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

func gitLFSInstall(gitDir string) (err error) {
	err = git.NewCtxAt(gitDir).Check("lfs", "install")

	if err != nil {
		err = cm.CombineErrors(err, cm.ErrorF("Could not install Git LFS hooks in\n"+
			"'%s'.\n"+
			"Please try manually by invoking:\n"+
			"  $ git -C '%[1]s' lfs install", gitDir))
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

	hooksDir := path.Join(l.repoDir, "hooks")
	if !cm.IsDirectory(hooksDir) || !git.NewCtxAt(l.repoDir).IsGitRepo() {
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

		err = gitLFSInstall(l.repoDir)
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
