package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func isLocalURL(url string) bool {
	return !IsCloneURLALocalPath(url) && IsCloneURLALocalURL(url)
}

func isLocalPath(url string) bool {
	return IsCloneURLALocalPath(url) && !IsCloneURLALocalURL(url)
}

func parsableAsSCP(url string) bool {
	return ParseSCPSyntax(url) != nil
}

func parsableAsRemoteHelper(url string) bool {
	return ParseRemoteHelperSyntax(url) != nil
}

func TestCoverage(t *testing.T) {

	// Local URLS
	url := "file://"
	assert.True(t, isLocalURL(url), "Local url")

	url = "file://\\a\b\\c\\d\\repo.git"
	assert.True(t, isLocalURL(url), "Local url")

	url = "file://"
	assert.True(t, isLocalURL(url), "Local url")

	url = "file://///a/b/c/d/repo.git"
	assert.True(t, isLocalURL(url), "Local url")
	assert.False(t, isLocalPath(url), "Local path")

	// Local Paths
	url = "/"
	assert.True(t, isLocalPath(url), "Local path")

	url = "\\a\b\\c\\d\\repo.git"
	assert.True(t, isLocalPath(url), "Local path")

	url = "../../a"
	assert.True(t, isLocalPath(url), "Local path")

	url = "///a/b/c/d/repo.git"
	assert.True(t, isLocalPath(url), "Local path")
	assert.False(t, isLocalURL(url), "No local path")

	// Other protocols
	url = "git://user@server.com:1234/~//a/b/c/d/repo.git"
	assert.False(t, isLocalURL(url), "Local url")
	assert.False(t, isLocalPath(url), "Local path")

	url = "git://server/a/b/c/d/repo.git"
	assert.False(t, isLocalURL(url), "Local url")
	assert.False(t, isLocalPath(url), "Local path")

	// Short scp syntax
	url = "user@server.com:1231/~here/repo.git"
	assert.False(t, isLocalURL(url), "Local url")
	assert.False(t, isLocalPath(url), "Local path")
	assert.True(t, parsableAsSCP(url), "Scp syntax")

	url = "server:/here/repo.git"
	assert.False(t, isLocalURL(url), "Local url")
	assert.False(t, isLocalPath(url), "Local path")
	assert.True(t, parsableAsSCP(url), "Scp syntax")

	url = "a:~/slightly/wrong/but/ok..."
	assert.False(t, isLocalURL(url), "Local url")
	assert.False(t, isLocalPath(url), "Local path")
	assert.True(t, parsableAsSCP(url), "Scp syntax")

	url = ":~/really/wrong/but/ok..."
	assert.False(t, isLocalURL(url), "Local url")
	assert.True(t, isLocalPath(url), "Local path")
	assert.False(t, parsableAsSCP(url), "Scp syntax")

	url = "user@server.com/~here/repo.git"
	assert.False(t, isLocalURL(url), "Local url")
	assert.True(t, isLocalPath(url), "Local path")
	assert.False(t, parsableAsSCP(url), "Scp syntax")

	// Transport helper syntax
	url = "<user@server.com??!super-protocol-stuff>::/~here/repo.git"
	assert.False(t, isLocalURL(url), "Local url")
	assert.False(t, isLocalPath(url), "Local path")
	assert.False(t, parsableAsSCP(url), "Scp syntax")
	assert.True(t, parsableAsRemoteHelper(url), "Remote helper syntax")

}
