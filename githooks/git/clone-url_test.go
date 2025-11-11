package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func parsableAsSCP(url string) bool {
	return ParseSCPSyntax(url) != nil
}

func parsableAsRemoteHelper(url string) bool {
	return ParseRemoteHelperSyntax(url) != nil
}

func TestCloneURLs(t *testing.T) {
	// Local URLS
	url := "file://"
	assert.True(t, IsCloneURLALocalURL(url), "Local url")
	assert.False(t, IsCloneURLALocalPath(url), "Local path")
	assert.True(t, IsCloneURLANormalURL(url), "Normal url")

	url = "file://\\a\b\\c\\d\\repo.git"
	assert.True(t, IsCloneURLALocalURL(url), "Local url")
	assert.False(t, IsCloneURLALocalPath(url), "Local path")
	assert.True(t, IsCloneURLANormalURL(url), "Normal url")

	url = "file://"
	assert.True(t, IsCloneURLALocalURL(url), "Local url")
	assert.False(t, IsCloneURLALocalPath(url), "Local path")
	assert.True(t, IsCloneURLANormalURL(url), "Normal url")

	url = "file://///a/b/c/d/repo.git"
	assert.True(t, IsCloneURLALocalURL(url), "Local url")
	assert.False(t, IsCloneURLALocalPath(url), "Local path")
	assert.True(t, IsCloneURLANormalURL(url), "Normal url")

	// Local Paths
	url = "/"
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.True(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")

	url = "\\a\b\\c\\d\\repo.git"
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.True(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")

	url = "../../a"
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.True(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")

	url = "///a/b/c/d/repo.git"
	assert.False(t, IsCloneURLALocalURL(url), "No local path")
	assert.True(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")

	url = "c:/a/b/c/d/repo.git"
	assert.False(t, IsCloneURLALocalURL(url), "No local path")
	assert.True(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")

	// Other protocols
	url = "git://user@server.com:1234/~//a/b/c/d/repo.git"
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.False(t, IsCloneURLALocalPath(url), "Local path")
	assert.True(t, IsCloneURLANormalURL(url), "Normal url")
	assert.True(t, parsableAsSCP(url), "Scp syntax") // not intended but its technically true...

	url = "git://server/a/b/c/d/repo.git"
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.False(t, IsCloneURLALocalPath(url), "Local path")
	assert.True(t, IsCloneURLANormalURL(url), "Normal url")
	assert.True(t, parsableAsSCP(url), "Scp syntax") // not intended but its technically true...

	url = "ssh://git@github.com/shared/hooks-maven.git"
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.False(t, IsCloneURLALocalPath(url), "Local path")
	assert.True(t, IsCloneURLANormalURL(url), "Normal url")
	assert.True(t, parsableAsSCP(url), "Scp syntax") // not intended but its technically true...

	// Short scp syntax
	url = "user@server.com:1231/~here/repo.git"
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.False(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")
	assert.True(t, parsableAsSCP(url), "Scp syntax")

	url = "server:/here/repo.git"
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.False(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")
	assert.True(t, parsableAsSCP(url), "Scp syntax")

	url = "a:~/slightly/wrong/but/ok..."
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.True(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")
	assert.False(t, parsableAsSCP(url), "Scp syntax")

	url = ":~/really/wrong/but/ok..."
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.True(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")
	assert.False(t, parsableAsSCP(url), "Scp syntax")
	assert.False(t, parsableAsRemoteHelper(url), "Remote helper syntax")

	url = "user@server.com/~here/repo.git"
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.True(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")
	assert.False(t, parsableAsSCP(url), "Scp syntax")
	assert.False(t, parsableAsRemoteHelper(url), "Remote helper syntax")

	// Transport helper syntax
	url = "<user@server.com??!super-protocol-stuff>::/~here/repo.git"
	assert.False(t, IsCloneURLALocalURL(url), "Local url")
	assert.False(t, IsCloneURLALocalPath(url), "Local path")
	assert.False(t, IsCloneURLANormalURL(url), "Normal url")
	assert.False(t, parsableAsSCP(url), "Scp syntax")
	assert.True(t, parsableAsRemoteHelper(url), "Remote helper syntax")
}
