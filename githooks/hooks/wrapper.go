package hooks

import (
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gabyx/githooks/githooks/build"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"
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

// InstallLinkRunWrappers installs a link with `core.hooksPath`
// to the maintained run-wrappers by Githooks
// and thus installs Githooks into the Git context, the local repository.
func InstallLinkRunWrappers(
	gitx *git.Context,
	dir string,
) (err error) {
	pathForUseCoreHooksPath, exists := gitx.LookupConfig(GitCKPathForUseCoreHooksPath, git.GlobalScope)

	if !exists || strs.IsEmpty(pathForUseCoreHooksPath) {
		return cm.ErrorF(
			"Githooks has not been installed.\n"+
				"The Git config variable '%s' does not exist or is empty.",
			GitCKPathForUseCoreHooksPath)
	}

	return gitx.SetConfig(git.GitCKCoreHooksPath, pathForUseCoreHooksPath, git.LocalScope)
}

// UninstallLinkRunWrappers uninstalls the link with `core.hooksPath`
// to the maintained run-wrappers by Githooks
// and thus installs Githooks into the Git context, the local repository.
func UninstallLinkRunWrappers(
	gitx *git.Context,
) (err error) {
	return gitx.UnsetConfig(git.GitCKCoreHooksPath, git.LocalScope)
}

// InstallRunWrappers installs run-wrappers for the given `hookNames` in `dir`.
// Existing custom hooks get renamed.
// All deleted hooks by this function get moved to the `tempDir` directory, because
// we should not delete them yet.
// Missing LFS hooks are reinstalled.
// Git context can be `nil` if its the global install into a directory.
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

		err = WriteRunWrapper(dest)
		if err != nil {
			err = cm.CombineErrors(err,
				cm.ErrorF("Could not write Githooks run-wrapper to '%s'.", dest))

			return
		}
	}

	err = cm.TouchFile(path.Join(dir, ".githooks-contains-run-wrappers"), true)
	if err != nil {
		err = cm.CombineErrors(err,
			cm.ErrorF("Could not create marker that directory '%s' contains run-wrappers.", dir))

		return
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

	_ = os.Remove(path.Join(dir, ".githooks-contains-run-wrappers"))

	return
}

// Get all missing LFS hook in `hookNames` and install them into `dir`.
func reinstallLFSHooks(
	dir string,
	hookNames []string,
	lfsHooksCache LFSHooksCache) (count int, err error) {

	if len(hookNames) == 0 {
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
