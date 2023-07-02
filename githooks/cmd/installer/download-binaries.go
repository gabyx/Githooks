//go:build !mock

package installer

import (
	"net/url"
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
		url, err := url.Parse(cloneURL)
		if err != nil {
			return nil, cm.ErrorF("Cannot parse clone url '%s'.", cloneURL)
		}

		host = url.Host
		owner, repo = path.Split(url.Path)

		owner = strings.TrimSpace(strings.ReplaceAll(owner, "/", ""))
		repo = strings.TrimSpace(strings.TrimSuffix(repo, ".git"))

	} else if userHostPath := git.ParseSCPSyntax(cloneURL); userHostPath != nil { //nolint: gocritic
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

	err := deploySettings.Download(log, versionTag, tempDir)
	log.AssertNoErrorPanicF(err, "Could not download binaries.")

	ext := ""
	if runtime.GOOS == cm.WindowsOsName {
		ext = cm.WindowsExecutableSuffix
	}

	all := []string{
		path.Join(tempDir, "cli"+ext),
		path.Join(tempDir, "runner"+ext),
		path.Join(tempDir, "dialog"+ext)}

	return updates.Binaries{All: all, Cli: all[0], Others: all[1:]}
}
