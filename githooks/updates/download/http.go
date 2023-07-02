package download

import (
	cm "github.com/gabyx/githooks/githooks/common"
)

// HTTPDeploySettings are deploy settings for HTTP downloading.
type HTTPDeploySettings struct {
	// Path template string which can contain
	// - `{{VersionTag}}` : The version tag to download.
	// - `{{Version}}` : The version to download (removed prefix 'v' of `VersionTag`).
	// - `{{Os}}` : The `runtime.GOOS` variable with the operating system.
	// - `{{Arch}}` : The `runtime.GOARCH` for type architecture.
	// pointing to the compressed archive of the Githooks binaries.
	// In the same url directory need to be a checksum file
	// `githooks.checksums`
	// and a checksum signature file.
	// `githooks.checksums.sig` which is validated using
	// the `PublicPGP`.
	URLTemplate string

	// If empty, the internal Githooks binary
	// embedded PGP is taken from `.deploy.pgp`.
	PublicPGP string
}

// Download downloads the Githooks from a template URL and
// extracts it into `dir`.
func (s *HTTPDeploySettings) Download(log cm.ILogContext, versionTag string, dir string) error {
	cm.Panic("Not implemented.")

	return nil
}
