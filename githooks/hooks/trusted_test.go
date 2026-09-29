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
	log, err := cm.CreateLogContext(false, false)
	assert.NoError(t, err)

	myOrg := "https://github.com/my-org/**"

	tests := []struct {
		name     string
		patterns []string
		url      string
		trusted  bool
		pattern  string
	}{
		{"match", []string{myOrg},
			"https://github.com/my-org/my-repo.git", true, myOrg},
		{"other organization", []string{myOrg},
			"https://github.com/other-org/my-repo.git", false, ""},
		{"other host", []string{myOrg},
			"https://gitlab.com/my-org/my-repo.git", false, ""},
		{"surrounding whitespace in url", []string{myOrg},
			"  https://github.com/my-org/my-repo.git \n", true, myOrg},

		// `*` does not match over `/`, `**` does.
		{"single star", []string{"https://github.com/my-org/*"},
			"https://github.com/my-org/my-repo.git", true, "https://github.com/my-org/*"},
		{"single star over separator", []string{"https://github.com/my-org/*"},
			"https://github.com/my-org/sub/my-repo.git", false, ""},
		{"double star over separator", []string{myOrg},
			"https://github.com/my-org/sub/my-repo.git", true, myOrg},

		// A `https://` pattern must not match the scp syntax url of the
		// same repository and vice versa.
		{"https pattern with scp url", []string{myOrg},
			"git@github.com:my-org/my-repo.git", false, ""},
		{"scp pattern with scp url", []string{"git@github.com:my-org/**"},
			"git@github.com:my-org/my-repo.git", true, "git@github.com:my-org/**"},
		{"scp pattern with https url", []string{"git@github.com:my-org/**"},
			"https://github.com/my-org/my-repo.git", false, ""},

		// A repository without a remote url must never be trusted,
		// also not by patterns matching everything.
		{"no url with star", []string{"*"}, "", false, ""},
		{"no url with double star", []string{"**"}, "", false, ""},
		{"whitespace url with double star", []string{"**"}, "  ", false, ""},

		// Empty patterns are skipped and must not match.
		{"no patterns", nil,
			"https://github.com/my-org/my-repo.git", false, ""},
		{"empty patterns", []string{"", " "},
			"https://github.com/my-org/my-repo.git", false, ""},

		// The pattern is matched against the whole url, an url only
		// containing it must not match.
		{"not anchored", []string{myOrg},
			"https://evil.com/x?url=https://github.com/my-org/my-repo.git", false, ""},
		{"similar looking host", []string{myOrg},
			"https://github.com.evil.com/my-org/my-repo.git", false, ""},

		{"first matching pattern", []string{"https://github.com/other-org/**", myOrg},
			"https://github.com/my-org/my-repo.git", true, myOrg},

		// Malformed patterns are skipped with a warning.
		{"malformed pattern", []string{"https://github.com/my-org/["},
			"https://github.com/my-org/my-repo.git", false, ""},
		{"malformed pattern skipped", []string{"https://github.com/my-org/[", myOrg},
			"https://github.com/my-org/my-repo.git", true, myOrg},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isTrusted, pattern := matchesTrustedRemote(log, test.patterns, test.url)
			assert.Equal(t, test.trusted, isTrusted)
			assert.Equal(t, test.pattern, pattern)
		})
	}
}

// testRepo is a repository with an isolated global and system Git configuration.
type testRepo struct {
	dir string
	env []string
}

// newTestRepo creates a repository with remote `origin` set to `originURL`
// (if not empty) and an isolated global and system Git configuration.
func newTestRepo(t *testing.T, originURL string) testRepo {
	t.Helper()

	configDir := t.TempDir()
	globalConfig := path.Join(configDir, "gitconfig-global")
	assert.NoError(t, os.WriteFile(globalConfig, []byte(""), cm.DefaultFileModeFile))

	r := testRepo{
		dir: t.TempDir(),
		// Isolate, such that the users configuration cannot influence the test.
		env: []string{
			"GIT_CONFIG_GLOBAL=" + globalConfig,
			"GIT_CONFIG_SYSTEM=" + path.Join(configDir, "gitconfig-system"),
		}}

	r.git(t, "init")

	if strs.IsNotEmpty(originURL) {
		r.git(t, "config", "remote."+TrustedRemoteName+".url", originURL)
	}

	return r
}

// git runs Git inside the repository.
func (r *testRepo) git(t *testing.T, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	cmd.Env = append(os.Environ(), r.env...)
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "git %v failed: %s", args, string(out))
}

// gitx returns a Git context inside the repository.
func (r *testRepo) gitx() *git.Context {
	return git.NewCtxAt(r.dir, git.WithModifications(
		func(builder *cm.CmdContextBuilder) *cm.CmdContextBuilder {
			return builder.AddEnv(r.env)
		}))
}

func (r *testRepo) makeTrustMarker(t *testing.T) {
	t.Helper()

	assert.NoError(t, os.MkdirAll(path.Join(r.dir, HooksDirName), cm.DefaultFileModeDirectory))
	assert.NoError(t, os.WriteFile(GetTrustMarkerFile(r.dir), []byte(""), cm.DefaultFileModeFile))
}

// isRepoTrusted reports `IsRepoTrusted` without and with an initialized
// Git config cache, since the runner uses a cache.
func (r *testRepo) isRepoTrusted(t *testing.T) (uncached bool, cached bool) {
	t.Helper()

	log, err := cm.CreateLogContext(false, false)
	assert.NoError(t, err)

	uncached, _, _ = IsRepoTrusted(log, r.gitx(), r.dir)

	gitx := r.gitx()
	assert.NoError(t, gitx.InitConfigCache(nil))
	cached, _, _ = IsRepoTrusted(log, gitx, r.dir)

	return
}

func TestRepoTrustedByRemoteNotConfigured(t *testing.T) {
	r := newTestRepo(t, "https://github.com/my-org/my-repo.git")

	uncached, cached := r.isRepoTrusted(t)
	assert.False(t, uncached)
	assert.False(t, cached)
}

func TestRepoTrustedByRemoteLocal(t *testing.T) {
	r := newTestRepo(t, "https://github.com/my-org/my-repo.git")
	r.git(t, "config", "--add", GitCKTrustedRemotes, "https://github.com/my-org/**")

	// No trust marker and no user interaction is needed.
	uncached, cached := r.isRepoTrusted(t)
	assert.True(t, uncached)
	assert.True(t, cached)
}

func TestRepoTrustedByRemoteGlobal(t *testing.T) {
	r := newTestRepo(t, "https://github.com/my-org/my-repo.git")
	r.git(t, "config", "--global", "--add",
		GitCKTrustedRemotes, "https://github.com/my-org/**")

	uncached, cached := r.isRepoTrusted(t)
	assert.True(t, uncached)
	assert.True(t, cached)
}

func TestRepoTrustedByRemoteOtherOrg(t *testing.T) {
	r := newTestRepo(t, "https://github.com/other-org/my-repo.git")
	r.git(t, "config", "--global", "--add",
		GitCKTrustedRemotes, "https://github.com/my-org/**")

	uncached, cached := r.isRepoTrusted(t)
	assert.False(t, uncached)
	assert.False(t, cached)
}

func TestRepoTrustedByRemoteWithoutRemote(t *testing.T) {
	r := newTestRepo(t, "")
	r.git(t, "config", "--global", "--add", GitCKTrustedRemotes, "**")

	// A repository without a remote is never trusted.
	uncached, cached := r.isRepoTrusted(t)
	assert.False(t, uncached)
	assert.False(t, cached)
}

func TestRepoTrustedByTrustMarkerAndDeniedByUser(t *testing.T) {
	r := newTestRepo(t, "https://github.com/my-org/my-repo.git")
	r.makeTrustMarker(t)
	r.git(t, "config", GitCKTrustAll, "true")

	// Trusted by the trust marker only, no trusted remotes are configured.
	uncached, cached := r.isRepoTrusted(t)
	assert.True(t, uncached)
	assert.True(t, cached)

	r.git(t, "config", "--global", "--add",
		GitCKTrustedRemotes, "https://github.com/my-org/**")
	r.git(t, "config", GitCKTrustAll, "false")

	// The explicit trust setting of the user wins over the trusted remotes.
	uncached, cached = r.isRepoTrusted(t)
	assert.False(t, uncached)
	assert.False(t, cached)
}

func TestRepoTrustedByRemoteShowsNoPrompt(t *testing.T) {
	r := newTestRepo(t, "https://github.com/my-org/my-repo.git")
	r.makeTrustMarker(t)
	r.git(t, "config", "--global", "--add",
		GitCKTrustedRemotes, "https://github.com/my-org/**")

	log, err := cm.CreateLogContext(false, false)
	assert.NoError(t, err)

	isTrusted, hasTrustFile, trustAllSet := IsRepoTrusted(log, r.gitx(), r.dir)
	assert.True(t, isTrusted)
	assert.True(t, hasTrustFile)
	assert.False(t, trustAllSet)

	// This is the condition the runner uses to show the trust prompt.
	assert.False(t, !isTrusted && hasTrustFile && !trustAllSet)
}
