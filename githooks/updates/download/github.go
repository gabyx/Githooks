package download

import (
	"context"
	cm "gabyx/githooks/common"
	"io"
	"os"
	"path"

	"github.com/google/go-github/v33/github"
)

// RepoSettings holds repo data for web based Git services such as Github or Gitea.
type RepoSettings struct {
	Owner      string // The owner of the repository.
	Repository string // The repository name.
}

// GithubDeploySettings are deploy settings for Github.
type GithubDeploySettings struct {
	RepoSettings

	// If empty, the internal Githooks binary
	// embedded PGP is taken from `.deploy.pgp`.
	PublicPGP string
}

// Download downloads the version with `versionTag` to `dir` from a Github instance.
func (s *GithubDeploySettings) Download(versionTag string, dir string) error {
	return downloadGithub(s.Owner, s.Repository, versionTag, dir, s.PublicPGP)
}

// Downloads the Githooks release with tag `versionTag` and
// extracts the matched asset into `dir`.
// The assert matches the OS and architecture of the current runtime.
func downloadGithub(owner string, repo string, versionTag string, dir string, publicPGP string) error {

	client := github.NewClient(nil)
	rel, _, err := client.Repositories.GetReleaseByTag(context.Background(),
		owner, repo, versionTag)
	if err != nil {
		return cm.CombineErrors(err, cm.Error("Failed to get release"))
	}

	// Wrap into our list
	var assets []Asset
	for i := range rel.Assets {
		assets = append(assets,
			Asset{
				FileName: path.Base(rel.Assets[i].GetName()),
				URL:      rel.Assets[i].GetBrowserDownloadURL()})
	}

	target, checksums, err := getGithooksAsset(assets)
	if err != nil {
		return cm.CombineErrors(err,
			cm.ErrorF("Could not select asset in repo '%s/%s' at tag '%s'.", owner, repo, versionTag))
	}

	checksumData, err := verifyChecksums(checksums, publicPGP)
	if err != nil {
		return cm.CombineErrors(err,
			cm.ErrorF("Signature verification of update failed."+
				"Something is fishy!"))
	}

	response, err := DownloadFile(target.URL)
	if err != nil {
		return cm.CombineErrors(err, cm.ErrorF("Could not download url '%s'.", target.URL))
	}
	defer response.Body.Close()

	// Store into temp. file.
	err = os.MkdirAll(dir, cm.DefaultFileModeDirectory)
	if err != nil {
		return cm.ErrorF("Could create dir '%s'.", dir)
	}

	tempFile := cm.GetTempPath(dir, target.FileName)
	temp, err := os.Create(tempFile)
	if err != nil {
		return cm.ErrorF("Could open temp file '%s' for download.", tempFile)
	}
	_, err = io.Copy(temp, response.Body)
	if err != nil {
		return cm.CombineErrors(err, cm.ErrorF("Could not store download in '%s'.", tempFile))
	}

	// Validate checksum.
	err = checkChecksum(tempFile, checksumData)
	if err != nil {
		return cm.CombineErrors(err, cm.ErrorF("Checksum validation failed."))
	}

	// Extract the file.
	err = Extract(tempFile, target.Extension, dir)
	if err != nil {
		return cm.CombineErrors(err,
			cm.ErrorF("Archive extraction from url '%s' failed.", target.URL))
	}

	return nil
}
