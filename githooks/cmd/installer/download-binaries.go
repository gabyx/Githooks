//go:build !mock && !download_mock

package installer

import (
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/gabyx/githooks/githooks/build"
	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"
	"github.com/gabyx/githooks/githooks/updates"
	"github.com/gabyx/githooks/githooks/updates/download"
)

// ghToken returns the GitHub/Gitea API token from the environment.
// It checks GH_TOKEN first, then falls back to GITHUB_TOKEN.
// The token is read here, at the application boundary, and threaded
// into the download library as a parameter.
func ghToken() string {
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token
	}

	return os.Getenv("GITHUB_TOKEN")
}

// IsRunningCoverage tells if we are running coverage.
const IsRunningCoverage = false

// detectDeploySettings tries to detect the deploy settings.
// Currently that works for Github automatically.
// For Gitea you need to specify the deploy api `deployAPI`.
// Others will fail and need a special deploy settings config file.
func detectDeploySettings(cloneURL string, deployAPI string) (download.IDeploySettings, error) {
	publicPGP, err := build.Asset("embedded/.deploy-pgp")
	cm.AssertNoErrorPanic(err, "Could not get embedded deploy PGP.")

	isLocal := git.IsCloneURLALocalPath(cloneURL) ||
		git.IsCloneURLALocalURL(cloneURL)
	if isLocal {
		return nil, cm.ErrorF(
			"Url '%s' points to a local directory.", cloneURL)
	}

	owner := ""
	repo := ""

	// Parse the url.
	host := ""

	if git.IsCloneURLANormalURL(cloneURL) {
		// Parse normal URL.
		url, e := url.Parse(cloneURL)
		if e != nil {
			return nil, cm.ErrorF("Cannot parse clone url '%s'.", cloneURL)
		}

		host = url.Host
		owner, repo = path.Split(url.Path)

		owner = strings.TrimSpace(strings.ReplaceAll(owner, "/", ""))
		repo = strings.TrimSpace(strings.TrimSuffix(repo, ".git"))
	} else if userHostPath := git.ParseSCPSyntax(cloneURL); userHostPath != nil {
		// Parse SCP Syntax.
		host = userHostPath[1]
		owner, repo = path.Split(userHostPath[2])

		owner = strings.TrimSpace(strings.TrimPrefix(owner, "/"))
		repo = strings.TrimSpace(strings.TrimSuffix(repo, ".git"))
	} else {
		return nil,
			cm.ErrorF("Cannot auto-determine deploy API for url '%s'.", cloneURL)
	}

	// If deploy API hint is not given,
	// define it by the parsed host.
	if strs.IsEmpty(deployAPI) {
		switch {
		case strings.Contains(host, "github"):
			deployAPI = "github"
		default:
			return nil,
				cm.ErrorF("Cannot auto-determine deploy API for host '%s'.", host)
		}
	}

	switch deployAPI {
	case "github":
		return &download.GithubDeploySettings{
			RepoSettings: download.RepoSettings{
				Owner:      owner,
				Repository: repo},
			PublicPGP: string(publicPGP)}, nil
	case "gitea":
		return &download.GiteaDeploySettings{
			APIUrl: "https://" + host + "/api/v1",
			RepoSettings: download.RepoSettings{
				Owner:      owner,
				Repository: repo},
			PublicPGP: string(publicPGP)}, nil
	default:
		return nil, cm.ErrorF("Deploy settings auto-detection for\n"+
			"deploy api '%s' not supported.",
			deployAPI)
	}
}

func downloadBinaries(
	log cm.ILogContext,
	deploySettings download.IDeploySettings,
	tempDir string,
	versionTag string) updates.Binaries {
	log.PanicIfF(deploySettings == nil,
		"Could not determine deploy settings.")

	err := deploySettings.Download(log, versionTag, tempDir, ghToken())
	log.AssertNoErrorPanicF(err, "Could not download binaries.")

	ext := ""
	if runtime.GOOS == cm.WindowsOsName {
		ext = cm.WindowsExecutableSuffix
	}

	cli := path.Join(tempDir, "githooks-cli"+ext)
	dialog := path.Join(tempDir, "githooks-dialog"+ext)
	runner := path.Join(tempDir, "githooks-runner"+ext)

	// Handle old binary names.
	if exists, _ := cm.IsPathExisting(path.Join(tempDir, "runner")); exists {
		e := os.Rename(path.Join(tempDir, "cli"+ext), cli)
		log.AssertNoErrorPanic(e, "Could not rename executable 'cli'.")

		e = os.Rename(path.Join(tempDir, "runner"+ext), runner)
		log.AssertNoErrorPanic(e, "Could not rename executable 'runner'.")

		e = os.Rename(path.Join(tempDir, "dialog"+ext), dialog)
		log.AssertNoErrorPanic(e, "Could not rename executable 'dialog'.")
	}

	all := []string{cli, runner, dialog}

	return updates.Binaries{All: all, Cli: all[0], Others: all[1:]}
}
