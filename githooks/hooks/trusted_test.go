package hooks

import (
	"os"
	"os/exec"
	"path"
	"testing"

	cm "github.com/gabyx/githooks/githooks/common"
	"github.com/gabyx/githooks/githooks/git"
	strs "github.com/gabyx/githooks/githooks/strings"
	"github.com/stretchr/testify/assert"
)

func TestTrustedRemoteMatch(t *testing.T) {
	patterns := []string{"https://github.com/my-org/**"}

	isTrusted, pattern := matchesTrustedRemote(
		patterns, "https://github.com/my-org/my-repo.git")
	assert.True(t, isTrusted)
	assert.Equal(t, patterns[0], pattern)

	// Other organizations must not match.
	isTrusted, _ = matchesTrustedRemote(
		patterns, "https://github.com/other-org/my-repo.git")
	assert.False(t, isTrusted)

	// Other hosts must not match.
	isTrusted, _ = matchesTrustedRemote(
		patterns, "https://gitlab.com/my-org/my-repo.git")
	assert.False(t, isTrusted)
}

func TestTrustedRemoteMatchSeparator(t *testing.T) {
	// `*` does not match over `/`, `**` does.
	single := []string{"https://github.com/my-org/*"}
	double := []string{"https://github.com/my-org/**"}

	isTrusted, _ := matchesTrustedRemote(single, "https://github.com/my-org/my-repo.git")
	assert.True(t, isTrusted)

	isTrusted, _ = matchesTrustedRemote(single, "https://github.com/my-org/sub/my-repo.git")
	assert.False(t, isTrusted)

	isTrusted, _ = matchesTrustedRemote(double, "https://github.com/my-org/sub/my-repo.git")
	assert.True(t, isTrusted)
}

func TestTrustedRemoteMatchScpSyntax(t *testing.T) {
	// A `https://` pattern must not match the scp syntax url of the
	// same repository and vice versa.
	https := []string{"https://github.com/my-org/**"}
	scp := []string{"git@github.com:my-org/**"}

	isTrusted, _ := matchesTrustedRemote(https, "git@github.com:my-org/my-repo.git")
	assert.False(t, isTrusted)

	isTrusted, _ = matchesTrustedRemote(scp, "git@github.com:my-org/my-repo.git")
	assert.True(t, isTrusted)

	isTrusted, _ = matchesTrustedRemote(scp, "https://github.com/my-org/my-repo.git")
	assert.False(t, isTrusted)
}

func TestTrustedRemoteMatchNoRemote(t *testing.T) {
	// A repository without a remote url must never be trusted,
	// also not by patterns matching everything.
	isTrusted, _ := matchesTrustedRemote([]string{"*"}, "")
	assert.False(t, isTrusted)

	isTrusted, _ = matchesTrustedRemote([]string{"**"}, "")
	assert.False(t, isTrusted)
}

func TestTrustedRemoteMatchNoPatterns(t *testing.T) {
	isTrusted, _ := matchesTrustedRemote(nil, "https://github.com/my-org/my-repo.git")
	assert.False(t, isTrusted)

	// Empty patterns are skipped and must not match.
	isTrusted, _ = matchesTrustedRemote(
		[]string{"", " "}, "https://github.com/my-org/my-repo.git")
	assert.False(t, isTrusted)
}

func TestTrustedRemoteMatchNotAnchored(t *testing.T) {
	patterns := []string{"https://github.com/my-org/**"}

	// The pattern is matched against the whole url, an url only
	// containing it must not match.
	isTrusted, _ := matchesTrustedRemote(
		patterns, "https://evil.com/x?url=https://github.com/my-org/my-repo.git")
	assert.False(t, isTrusted)

	// A similar looking host must not match.
	isTrusted, _ = matchesTrustedRemote(
		patterns, "https://github.com.evil.com/my-org/my-repo.git")
	assert.False(t, isTrusted)
}

func TestTrustedRemoteMatchFirstPattern(t *testing.T) {
	patterns := []string{
		"https://github.com/other-org/**",
		"https://github.com/my-org/**",
	}

	isTrusted, pattern := matchesTrustedRemote(
		patterns, "https://github.com/my-org/my-repo.git")
	assert.True(t, isTrusted)
	assert.Equal(t, patterns[1], pattern)
}

// runGitIn runs Git inside `dir`.
func runGitIn(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "git %v failed: %s", args, string(out))
}

// makeRepo creates a repository with remote `origin` set to `originURL`
// (if not empty) and an isolated global and system Git configuration.
func makeRepo(t *testing.T, originURL string) string {
	t.Helper()

	dir := t.TempDir()
	globalConfig := path.Join(t.TempDir(), "gitconfig-global")

	// Isolate, such that the users configuration cannot influence the test.
	t.Setenv("GIT_CONFIG_GLOBAL", globalConfig)
	t.Setenv("GIT_CONFIG_SYSTEM", path.Join(t.TempDir(), "gitconfig-system"))
	assert.NoError(t, os.WriteFile(globalConfig, []byte(""), cm.DefaultFileModeFile))

	runGitIn(t, dir, "init")

	if strs.IsNotEmpty(originURL) {
		runGitIn(t, dir, "config", "remote."+TrustedRemoteName+".url", originURL)
	}

	return dir
}

func makeTrustMarker(t *testing.T, dir string) {
	t.Helper()

	assert.NoError(t, os.MkdirAll(path.Join(dir, HooksDirName), cm.DefaultFileModeDirectory))
	assert.NoError(t, os.WriteFile(GetTrustMarkerFile(dir), []byte(""), cm.DefaultFileModeFile))
}

// isRepoTrusted reports `IsRepoTrusted` without and with an initialized
// Git config cache, since the runner uses a cache.
func isRepoTrusted(t *testing.T, dir string) (uncached bool, cached bool) {
	t.Helper()

	uncached, _, _ = IsRepoTrusted(git.NewCtxAt(dir), dir)

	gitx := git.NewCtxAt(dir)
	assert.NoError(t, gitx.InitConfigCache(nil))
	cached, _, _ = IsRepoTrusted(gitx, dir)

	return
}

func TestRepoTrustedByRemoteNotConfigured(t *testing.T) {
	dir := makeRepo(t, "https://github.com/my-org/my-repo.git")

	uncached, cached := isRepoTrusted(t, dir)
	assert.False(t, uncached)
	assert.False(t, cached)
}

func TestRepoTrustedByRemoteLocal(t *testing.T) {
	dir := makeRepo(t, "https://github.com/my-org/my-repo.git")
	runGitIn(t, dir, "config", "--add", GitCKTrustedRemotes, "https://github.com/my-org/**")

	// No trust marker and no user interaction is needed.
	uncached, cached := isRepoTrusted(t, dir)
	assert.True(t, uncached)
	assert.True(t, cached)
}

func TestRepoTrustedByRemoteGlobal(t *testing.T) {
	dir := makeRepo(t, "https://github.com/my-org/my-repo.git")
	runGitIn(t, dir, "config", "--global", "--add",
		GitCKTrustedRemotes, "https://github.com/my-org/**")

	uncached, cached := isRepoTrusted(t, dir)
	assert.True(t, uncached)
	assert.True(t, cached)
}

func TestRepoTrustedByRemoteOtherOrg(t *testing.T) {
	dir := makeRepo(t, "https://github.com/other-org/my-repo.git")
	runGitIn(t, dir, "config", "--global", "--add",
		GitCKTrustedRemotes, "https://github.com/my-org/**")

	uncached, cached := isRepoTrusted(t, dir)
	assert.False(t, uncached)
	assert.False(t, cached)
}

func TestRepoTrustedByRemoteWithoutRemote(t *testing.T) {
	dir := makeRepo(t, "")
	runGitIn(t, dir, "config", "--global", "--add", GitCKTrustedRemotes, "**")

	// A repository without a remote is never trusted.
	uncached, cached := isRepoTrusted(t, dir)
	assert.False(t, uncached)
	assert.False(t, cached)
}

func TestRepoTrustedByRemoteDeniedByUser(t *testing.T) {
	dir := makeRepo(t, "https://github.com/my-org/my-repo.git")
	makeTrustMarker(t, dir)
	runGitIn(t, dir, "config", "--global", "--add",
		GitCKTrustedRemotes, "https://github.com/my-org/**")
	runGitIn(t, dir, "config", GitCKTrustAll, "false")

	// The explicit trust setting of the user wins.
	uncached, cached := isRepoTrusted(t, dir)
	assert.False(t, uncached)
	assert.False(t, cached)
}

func TestRepoTrustedByTrustMarkerOnly(t *testing.T) {
	dir := makeRepo(t, "https://github.com/other-org/my-repo.git")
	makeTrustMarker(t, dir)
	runGitIn(t, dir, "config", GitCKTrustAll, "true")

	uncached, cached := isRepoTrusted(t, dir)
	assert.True(t, uncached)
	assert.True(t, cached)
}

func TestRepoTrustedByRemoteShowsNoPrompt(t *testing.T) {
	dir := makeRepo(t, "https://github.com/my-org/my-repo.git")
	makeTrustMarker(t, dir)
	runGitIn(t, dir, "config", "--global", "--add",
		GitCKTrustedRemotes, "https://github.com/my-org/**")

	isTrusted, hasTrustFile, trustAllSet := IsRepoTrusted(git.NewCtxAt(dir), dir)
	assert.True(t, isTrusted)
	assert.True(t, hasTrustFile)
	assert.False(t, trustAllSet)

	// This is the condition the runner uses to show the trust prompt.
	assert.False(t, !isTrusted && hasTrustFile && !trustAllSet)
}
