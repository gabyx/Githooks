package updates

import (
	"regexp"
	"strings"

	"github.com/gabyx/githooks/githooks/build"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	"github.com/gabyx/githooks/githooks/hooks"
	"github.com/gabyx/githooks/githooks/prompt"
	strs "github.com/gabyx/githooks/githooks/strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
)

// ReleaseStatus contains the status of the release clone.
type ReleaseStatus struct {
	RemoteURL  string
	RemoteName string

	IsNewClone bool

	LocalCommitSHA  string // Current local commit and corresponding to the current installed version.
	LocalTag        string // Current local tag  and corresponding to the current installed version.
	RemoteCommitSHA string // Remote head on the remote branch.

	IsUpdateAvailable bool
	UpdateCommitSHA   string           // The determined update SHA (always <= RemoteCommitSHA).
	UpdateTag         string           // The update tag.
	UpdateVersion     *version.Version // The update version.
	UpdateInfo        []string         // The update info read from the commit.

	Branch       string
	RemoteBranch string
}

// GetCloneURL get the clone url and clone branch.
func GetCloneURL() (url string, branch string) {
	gitx := git.Ctx()
	url = gitx.GetConfig(hooks.GitCKCloneURL, git.GlobalScope)
	branch = gitx.GetConfig(hooks.GitCKCloneBranch, git.GlobalScope)

	return
}

// SetCloneURL get the clone url and clone branch.
// The `branch` can be empty.
func SetCloneURL(url string, branch string) (err error) {
	cm.DebugAssertF(strs.IsNotEmpty(url), "Wrong input")

	err = git.Ctx().SetConfig(hooks.GitCKCloneURL, url, git.GlobalScope)
	if err != nil || strs.IsEmpty(branch) {
		return
	}

	return SetCloneBranch(branch)
}

// SetCloneBranch sets the Githooks clone branch.
func SetCloneBranch(branch string) error {
	cm.DebugAssertF(strs.IsNotEmpty(branch), "Wrong input")

	return git.Ctx().SetConfig(hooks.GitCKCloneBranch, branch, git.GlobalScope)
}

// ResetCloneBranch resets the Githooks clone branch.
func ResetCloneBranch() error {
	return git.Ctx().UnsetConfig(hooks.GitCKCloneBranch, git.GlobalScope)
}

var updateInfoTrailerRe = regexp.MustCompile(`(?m)^Update-Info: *(.*)`)
var unskipTrailerRe = regexp.MustCompile(`(?m)^Update-NoSkip: *true`)

func getNewUpdateCommit(
	gitx *git.Context,
	firstSHA string,
	lastSHA string,
	skipPrerelease bool) (commitF string, tagF string, versionF *version.Version, infoF []string, err error) {

	// Get all commits in (firstSHA, lastSHA]
	commits, err := gitx.GetCommits(firstSHA, lastSHA)
	if err != nil {
		return
	}

	// Iterate from (firstSHA, lastSHA], the list is reversed.
	for i := range commits {

		commit := commits[len(commits)-1-i]

		version, tag, e := git.GetVersionAt(gitx, commit)

		switch {
		case e != nil:
			err = e

			return
		case version == nil || strs.IsEmpty(tag):
			continue // no version tag on this commit
		case skipPrerelease && strs.IsNotEmpty(version.Prerelease()):
			// Skipping prerelease version
			continue
		}

		// Check if it is an unskippable commit:
		// Get message of the tag (or the commit, if no annotated tag)
		mess, e := gitx.Get("tag", "-l", "--format=%(contents)", tag)
		if e != nil {
			err = e

			return
		}

		// We have a valid new version on commit 'commit'
		commitF = commit
		tagF = tag
		versionF = version

		// Add update info to the list.
		info := updateInfoTrailerRe.FindStringSubmatch(mess)
		if info != nil {
			infoF = append(infoF, strs.Fmt("%s : ", version.String())+strings.TrimSpace(info[1]))
		}

		if unskipTrailerRe.MatchString(mess) {
			// We stop at this commit since this update cannot be skipped!
			break
		}
	}

	return
}

// RemoteCheckAction is the action type for the remote check.
type RemoteCheckAction string

const (
	// ErrorOnWrongRemote errors out if wrong remote detected.
	ErrorOnWrongRemote RemoteCheckAction = "error"
	// RecloneOnWrongRemote reclones if wrong remote detected.
	RecloneOnWrongRemote RemoteCheckAction = "reclone"
)

// FetchUpdates fetches updates in the Githooks clone directory.
// Arguments `url` and `branch` can be empty which triggers.
func FetchUpdates(
	cloneDir string,
	url string,
	branch string,
	checkRemote bool,
	checkRemoteAction RemoteCheckAction,
	skipPrerelease bool) (status ReleaseStatus, err error) {

	cm.AssertOrPanic(strs.IsNotEmpty(cloneDir))

	// Repo check function before fetch is executed.
	check := func(gitx git.Context, url string, branch string) (bool, error) {
		reclone := false

		// Check if clone is dirty, if so error out.
		exitCode, e := gitx.GetExitCode("diff-index", "--quiet", git.HEAD)
		if e != nil {
			return false,
				cm.CombineErrors(cm.ErrorF("Could not check dirty state in '%s'",
					gitx.Cwd),
					e)
		}

		if exitCode != 0 {
			return false, cm.ErrorF("Cannot fetch updates because the clone\n"+
				"'%s'\n"+
				"is dirty! Either fix this or delete the clone\n"+
				"to trigger a new checkout.", gitx.Cwd)
		}

		if checkRemote {
			u, b, e := gitx.GetRemoteURLAndBranch(DefaultRemote)

			if e != nil {
				return false, cm.CombineErrors(cm.ErrorF(
					"Could not check url & branch in repository at '%s'", gitx.Cwd), e)

			} else if u != url || b != branch {
				if checkRemoteAction != RecloneOnWrongRemote {
					// Default action is error out:
					return false, cm.ErrorF("Cannot fetch updates because '%[6]s' of clone\n"+
						"'%[1]s'\n"+
						"points to url:\n"+
						"'%[2]s'\n"+
						"on branch '%[3]s'\n"+
						"which is not requested\n"+
						" - url: '%[4]s'\n"+
						" - branch: '%[5]s'\n"+
						"See 'git hooks config [set|print] clone-url' and\n"+
						"    'git hooks config [set|print] clone-branch'\n"+
						"Either fix this or delete the clone\n"+
						"'%[1]s'\n"+
						"to trigger a new checkout.", gitx.Cwd, u, b, url, branch, DefaultRemote)
				}

				// Do a reclone
				reclone = true
			}
		}

		return reclone, nil
	}

	cURL, cBranch := GetCloneURL()

	// Fallback for url
	if strs.IsEmpty(url) {
		url = cURL
	}
	if strs.IsEmpty(url) {
		url = GetDefaultCloneURL()
	}

	// Fallback for branch.
	if strs.IsEmpty(branch) {
		branch = cBranch
	}
	if strs.IsEmpty(branch) {
		branch = GetDefaultCloneBranch()
	}

	// Fetch or clone the repository:
	// Fetch the whole branch because we need it.
	// The head could be in between tags which we avoid by this.
	depth := -1
	isNewClone, err := git.FetchOrClone(cloneDir, url, branch, depth, "v*", check)
	if err != nil {
		return
	}

	gitx := git.CtxCSanitized(cloneDir)
	resetRemoteTo := ""

	// If branch was empty (default branch), determine it now.
	if strs.IsEmpty(branch) {
		if branch, err = gitx.GetCurrentBranch(); err != nil {
			return
		}
	}

	// Set the url/branch back...
	if err = SetCloneURL(url, branch); err != nil {
		return
	}

	if isNewClone {
		// On a new clone we reset local branch
		// to the matching current release tag descending from HEAD.
		// Remote stays and might trigger a direct update).

		// Check if current tag is reachable from HEAD.
		reachable, e := git.IsRefReachable(gitx, "HEAD", build.BuildTag)
		if e != nil || !reachable {
			err = cm.CombineErrors(
				cm.ErrorF("Current version tag '%v' could not be found on branch '%s'",
					build.BuildTag, branch), e)

			return
		}

		e = gitx.Check("reset", "--hard", build.BuildTag)
		if e != nil {
			err = cm.CombineErrors(
				cm.ErrorF("Could not reset branch '%s' to tag '%s'",
					branch, build.BuildTag), e)

			return
		}
	}

	remoteBranch := DefaultRemote + "/" + branch
	status, err = getStatus(gitx, url, DefaultRemote, branch, remoteBranch, skipPrerelease)

	status.IsNewClone = isNewClone
	if status.IsUpdateAvailable {
		resetRemoteTo = status.UpdateCommitSHA
	}

	if strs.IsNotEmpty(resetRemoteTo) {
		// Reset the release branch to determined update commit sha.
		err = gitx.Check("update-ref", "refs/remotes/"+remoteBranch, resetRemoteTo)
		if err != nil {
			return
		}

		status.RemoteCommitSHA = resetRemoteTo
	}

	return
}

// GetStatus returns the status of the release clone.
func GetStatus(cloneDir string, checkRemote, skipPrerelease bool) (status ReleaseStatus, err error) {

	gitx := git.CtxCSanitized(cloneDir)

	var url, branch string
	url, branch, err = gitx.GetRemoteURLAndBranch(DefaultRemote)
	cm.DebugAssert(strs.IsNotEmpty(url) &&
		strs.IsNotEmpty(branch), "Wrong output!")
	if err != nil {
		return
	}

	if checkRemote {
		configURL, configBranch := GetCloneURL()

		if url != configURL || branch != configBranch {
			// Default action is error out:
			err = cm.ErrorF("Cannot get status because '%s' of clone\n"+
				"'%[1]s'\n"+
				"points to url:\n"+
				"'%[2]s'\n"+
				"on branch '%[3]s'\n"+
				"which is not requested\n"+
				" - url: '%[4]s'\n"+
				" - branch: '%[5]s'\n"+
				"See 'git hooks config [set|print] clone-url' and\n"+
				"    'git hooks config [set|print] clone-branch'\n"+
				"Either fix this or delete the clone\n"+
				"'%[1]s'\n"+
				"to trigger a new checkout.", gitx.Cwd, url, branch, configURL, configBranch)

			return
		}
	}

	remoteBranch := DefaultRemote + "/" + branch

	return getStatus(gitx, url, DefaultRemote, branch, remoteBranch, skipPrerelease)
}

func getStatus(
	gitx *git.Context,
	url string,
	remoteName string,
	branch string,
	remoteBranch string,
	skipPrerelease bool) (status ReleaseStatus, err error) {

	localSHA, err := gitx.Get("rev-parse", branch)
	if err != nil {
		return
	}

	// Get the tag (corresponding to a version)
	// at the latest commit.
	_, tag, _ := git.GetVersionAt(gitx, localSHA)

	remoteSHA, err := gitx.Get("rev-parse", remoteBranch)
	if err != nil {
		return
	}

	var updateInfo []string
	updateCommit := ""
	updateTag := ""
	var updateVersion *version.Version

	if localSHA != remoteSHA {
		// We have a potential update available...
		// Get the latest update commit in the range (localSHA, remoteSHA]
		// - Skip prerelease versions
		// - also never skip annotated (Git trailers) "non-skip" versions.
		updateCommit, updateTag, updateVersion, updateInfo, err =
			getNewUpdateCommit(gitx, localSHA, remoteSHA, skipPrerelease)

		if err != nil {
			return
		}
	}

	status = ReleaseStatus{
		RemoteURL:  url,
		RemoteName: remoteName,

		LocalCommitSHA: localSHA,
		LocalTag:       tag,

		RemoteCommitSHA: remoteSHA,

		IsUpdateAvailable: updateVersion != nil,
		UpdateVersion:     updateVersion,
		UpdateCommitSHA:   updateCommit,
		UpdateTag:         updateTag,
		UpdateInfo:        updateInfo,

		Branch:       branch,
		RemoteBranch: remoteBranch}

	return
}

// MergeUpdates merges updates in the Githooks clone directory.
// Only a fast-forward merge of the remote branch into the local
// branch is performed, no remote fetch is performed.
// Returns the commit SHA after the fast-forward.
func MergeUpdates(cloneDir string, dryRun bool) (currentSHA string, err error) {
	if !cm.IsDirectory(cloneDir) {
		err = cm.ErrorF("Clone directory '%s' does not exist.", cloneDir)
		return //nolint: nlreturn
	}

	// Get configured branches...
	_, branch := GetCloneURL()
	remoteBranch := "origin/" + branch

	gitx := git.CtxCSanitized(cloneDir)

	// Safety check that branches are the same.
	currentBranch, err := gitx.Get("rev-parse", "--abbrev-ref", git.HEAD)
	if err != nil {
		return
	}

	if currentBranch != branch {
		err = cm.ErrorF("Current branch of clone directory\n'%s'\n"+
			"does not point to the configured branch '%s'\n"+
			"but instead to '%s'.", cloneDir, branch, currentBranch)
		return //nolint: nlreturn
	}

	if dryRun {
		// Checkout a temporary branch from the current
		// and merge the remote to see if it works.
		oldBranch := branch
		branch = "dryrunmerge-" + uuid.New().String() // ust this dry-run branch from now
		if err = gitx.Check("branch", branch, oldBranch); err != nil {
			return
		}

		// Delete the branch on exit.
		defer func() {
			err = cm.CombineErrors(err, gitx.Check("branch", "-D", branch))
		}()

		// Checkout the branch
		if err = gitx.Check("checkout", branch); err != nil {
			return
		}

		// Checkout the old branch on exit.
		defer func() {
			err = cm.CombineErrors(err, gitx.Check("checkout", oldBranch))
		}()
	}

	// Fast-forward merge.
	if err = gitx.Check("merge", "--ff-only", remoteBranch); err != nil {
		return
	}

	// Get the current commit SHA1 after the merge.
	currentSHA, err = gitx.Get("rev-parse", branch)

	return
}

// AcceptUpdateCallback is the callback type
// for accepting/rejecting updates.
type AcceptUpdateCallback func(status *ReleaseStatus) bool

// RunUpdate runs the procedure of updating Githooks.
func RunUpdate(
	installDir string,
	acceptUpdate AcceptUpdateCallback,
	run func() error) (updateAvailable bool, accepted bool, err error) {

	err = RecordUpdateCheckTimestamp()

	if err != nil {
		err = cm.Error("Could not record update check timestamp.")

		return
	}

	cloneDir := hooks.GetReleaseCloneDir(installDir)
	skipPrerelease := true
	status, err := FetchUpdates(cloneDir, "", "", true, ErrorOnWrongRemote, skipPrerelease)
	if err != nil {
		err = cm.CombineErrors(cm.Error("Could not fetch updates."), err)

		return
	}

	updateAvailable = status.IsUpdateAvailable

	if status.IsUpdateAvailable {
		accepted = acceptUpdate(&status)
	}

	if updateAvailable && accepted {

		_, err = MergeUpdates(cloneDir, true) // Dry run the merge...
		if err != nil {
			err = cm.CombineErrors(cm.ErrorF(
				"Update cannot run:\n"+
					"Your release clone '%s' cannot be fast-forward merged.\n"+
					"Either fix this or delete the clone to retry.",
				cloneDir), err)

			return
		}

		err = run()
	}

	return
}

// formatUpdateInfo formats all update infos.
func formatUpdateInfo(updateInfo []string) string {
	return strs.Fmt("\nUpdate Info:\n%s",
		strings.Join(strs.Map(updateInfo,
			func(s string) string {
				return "  - " + strings.ReplaceAll(s, "\n", "\n    ")
			}), "\n"))
}

// DefaultAcceptUpdateCallback creates a default accept update callback
// which prompts the user.
func DefaultAcceptUpdateCallback(
	log cm.ILogContext,
	promptCtx prompt.IContext,
	acceptIfNoPrompt bool) AcceptUpdateCallback {

	return func(status *ReleaseStatus) bool {
		log.DebugF("Fetch status: '%v'", status)
		cm.DebugAssert(status.IsUpdateAvailable, "Wrong input.")

		versionText := strs.Fmt(
			"Current Version: '%s'\n"+
				"New Version: '%s'",
			build.GetBuildVersion(),
			status.UpdateVersion.String())

		isMajorUpdate := build.GetBuildVersion().Segments()[0] < status.UpdateVersion.Segments()[0]

		promptHint := "(Yes/no)"
		promptDefault := "Y/n"
		if isMajorUpdate {
			versionText += " (Major Update)"
			promptHint = "(yes/No)"
			promptDefault = "y/N"
		}

		if status.UpdateInfo != nil {
			versionText += formatUpdateInfo(status.UpdateInfo)
		}

		if promptCtx != nil {
			question := "There is a new Githooks update available:\n" +
				versionText + "\n" +
				"Would you like to install it now?"

			answer, err := promptCtx.ShowOptions(question,
				promptHint,
				promptDefault,
				"Yes", "No")
			log.AssertNoErrorF(err, "Could not show prompt.")

			if answer == "y" {

				return true
			}

		} else {
			log.InfoF("There is a new Githooks update available:\n%s", versionText)

			if acceptIfNoPrompt {

				return true
			}
		}

		return false
	}
}

// RunUpdateOverExecutbale executes the installer to run the update.
func RunUpdateOverExecutable(
	installDir string,
	execC cm.IExecContext,
	pipeSetup cm.PipeSetupFunc,
	args ...string) error {

	installer := hooks.GetInstallerExecutable(installDir)

	execX := cm.ExecContext{Cwd: execC.GetWorkingDir()}
	execX.Env = git.SanitizeEnv(execC.GetEnv())

	if !cm.IsFile(installer.Cmd) {
		return cm.ErrorF(
			"Could not execute update, because Githooks executable:\n"+
				"'%s'\n"+
				"is not existing.", installer.Cmd)
	}

	if pipeSetup == nil {
		output, err := cm.GetCombinedOutputFromExecutable(
			&execX,
			&installer,
			nil,
			args...)

		if err != nil {
			return cm.CombineErrors(err, cm.ErrorF("Update output:\n%s", output))
		}

	} else {
		err := cm.RunExecutable(
			&execX,
			&installer,
			pipeSetup,
			args...)

		if err != nil {
			return cm.CombineErrors(err, cm.Error("Update failed. See output"))
		}
	}

	return nil
}
