package download

import (
	"io"
	"os"
	"path"

	cm "github.com/gabyx/githooks/githooks/common"

	"code.gitea.io/sdk/gitea"
)

// GiteaDeploySettings are deploy settings for Gitea.
type GiteaDeploySettings struct {
	RepoSettings
	APIUrl string // API url of the Gitea service.

	// If empty, the internal Githooks binary
	// embedded PGP is taken from `.deploy.pgp`.
	PublicPGP string
}

// Download downloads the version with `versionTag` into `dir` from a Gitea instance.
func (s *GiteaDeploySettings) Download(log cm.ILogContext, versionTag string, dir string) error {
	return downloadGitea(log, s.APIUrl, s.Owner, s.Repository, versionTag, dir, s.PublicPGP)
}

// Downloads the Githooks release with tag `versionTag` and
// extracts the matched asset into `dir`.
// The assert matches the OS and architecture of the current runtime.
func downloadGitea(
	log cm.ILogContext,
	url string,
	owner string,
	repo string,
	versionTag string,
	dir string,
	publicPGP string) error {

	client, err := gitea.NewClient(url)
	if err != nil {
		return cm.CombineErrors(err, cm.Error("Cannot initialize Gitea client"))
	}

	rel, _, err := client.GetReleaseByTag(owner, repo, versionTag)
	if err != nil {
		return cm.CombineErrors(err, cm.Error("Failed to get release"))
	}

	// Wrap into our list
	var assets []Asset
	for i := range rel.Attachments {
		assets = append(assets,
			Asset{
				FileName: path.Base(rel.Attachments[i].Name),
				URL:      rel.Attachments[i].DownloadURL})
	}

	target, checksums, err := getGithooksAsset(assets)
	if err != nil {
		return cm.CombineErrors(err,
			cm.ErrorF("Could not select asset in repo '%s/%s' at tag '%s'.", owner, repo, versionTag))
	}

	log.InfoF("Verify signature of checksum file '%s'.", checksums.File.URL)
	checksumData, err := verifyChecksumSignature(checksums, publicPGP)
	if err != nil {
		return cm.CombineErrors(err,
			cm.ErrorF("Signature verification of update failed."+
				"Something is fishy!"))
	}

	log.InfoF("Downloading file '%s'.", target.URL)
	response, err := GetFile(target.URL)
	if err != nil {
		return cm.CombineErrors(err, cm.ErrorF("Could not download url '%s'.", target.URL))
	}
	defer response.Body.Close()

	// Store into temp. file.
	err = os.MkdirAll(dir, cm.DefaultFileModeDirectory)
	if err != nil {
		return cm.ErrorF("Could create dir '%s'.", dir)
	}

	temp, err := os.CreateTemp(dir, "*-"+target.FileName)
	if err != nil {
		return cm.ErrorF("Could open temp file '%s' for download.", target.FileName)
	}
	_, err = io.Copy(temp, response.Body)
	if err != nil {
		return cm.CombineErrors(err, cm.ErrorF("Could not store download in '%s'.", temp.Name()))
	}
	temp.Close()

	log.InfoF("Validate checksums.")
	err = checkChecksum(temp.Name(), checksumData)
	if err != nil {
		return cm.CombineErrors(err, cm.ErrorF("Checksum validation failed."))
	}

	err = Extract(temp.Name(), target.Extension, dir)
	if err != nil {
		return cm.CombineErrors(err,
			cm.ErrorF("Archive extraction from url '%s' failed.", url))
	}

	return nil
}
